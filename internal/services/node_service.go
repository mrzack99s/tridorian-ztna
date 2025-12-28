package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"context"
	"fmt"
	"tridorian-ztna/internal/models"
	pb "tridorian-ztna/internal/proto/gateway/v1"
	"tridorian-ztna/pkg/utils"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type NodeService struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewNodeService(db *gorm.DB, cache *redis.Client) *NodeService {
	return &NodeService{db: db, cache: cache}
}

func (s *NodeService) ListNodes(tenantID uuid.UUID) ([]models.Node, error) {
	var nodes []models.Node
	if err := s.db.Scopes(models.TenantScope(tenantID)).Preload("AccessPolicies").Preload("NodeSku").Find(&nodes).Error; err != nil {
		return nil, err
	}

	// Update status based on Valkey liveness
	if s.cache != nil {
		for i := range nodes {
			if nodes[i].Status == "PENDING_REGISTRATION" {
				continue
			}
			key := fmt.Sprintf("node:liveness:%s", nodes[i].ID.String())
			exists, _ := s.cache.Exists(context.Background(), key).Result()
			if exists == 0 {
				nodes[i].Status = "DISCONNECTED"
			} else {
				nodes[i].Status = "CONNECTED"
			}
		}
	}

	return nodes, nil
}

func (s *NodeService) GetSessionIP(nodeID uuid.UUID, userID, email string) (string, error) {
	if s.cache == nil {
		return "", errors.New("cache not available")
	}

	var node models.Node
	if err := s.db.First(&node, "id = ?", nodeID).Error; err != nil {
		return "", err
	}

	if node.ClientCIDR == "" {
		return "", errors.New("node has no client CIDR configured")
	}

	ctx := context.Background()
	assignmentKey := fmt.Sprintf("ip:user:%s:%s", node.TenantID, userID)

	// 1. Check for sticky assignment (1 hour TTL)
	assignedIP, err := s.cache.Get(ctx, assignmentKey).Result()
	if err == nil && assignedIP != "" {
		// Refresh TTL
		s.cache.Expire(ctx, assignmentKey, 1*time.Hour)
		return assignedIP, nil
	}

	// 2. Allocate new IP
	firstIP, lastIP, err := utils.GetIPRange(node.ClientCIDR)
	if err != nil {
		return "", err
	}

	// Usually skip network and broadcast address
	currentIP := utils.IncrementIP(firstIP) // Skip .0 (network)

	// Check if .1 is typically gateway. If so, skip it too.
	// Let's assume user IPs start from .2 or .10
	// For simplicity, let's start from .2
	currentIP = utils.IncrementIP(currentIP)

	for {
		ipStr := currentIP.String()
		if currentIP.String() == lastIP.String() {
			return "", errors.New("no more IPs available in CIDR")
		}

		reverseKey := fmt.Sprintf("ip:allocated:%s:%s", node.TenantID, ipStr)

		// Try to claim this IP
		success, err := s.cache.SetNX(ctx, reverseKey, userID, 1*time.Hour).Result()
		if err == nil && success {
			// Save assignment
			s.cache.Set(ctx, assignmentKey, ipStr, 1*time.Hour)
			return ipStr, nil
		}

		currentIP = utils.IncrementIP(currentIP)
	}
}

func (s *NodeService) SyncSessions(nodeID uuid.UUID, sessions []*pb.SyncSessionsRequest_Session) error {
	if s.cache == nil {
		return nil
	}

	var node models.Node
	if err := s.db.First(&node, "id = ?", nodeID).Error; err != nil {
		return err
	}

	ctx := context.Background()
	sessionsKey := fmt.Sprintf("node:sessions:%s", nodeID)

	// 1. Clear old sessions for this node or just overwrite?
	// To be accurate, we should probably use a Set.
	s.cache.Del(ctx, sessionsKey)

	for _, sess := range sessions {
		// Store session info as JSON
		sessionData := map[string]interface{}{
			"user_id":      sess.UserId,
			"user_email":   sess.UserEmail,
			"ip_address":   sess.IpAddress,
			"connected_at": sess.ConnectedAt,
		}
		jsonData, _ := json.Marshal(sessionData)
		s.cache.HSet(ctx, sessionsKey, sess.UserId, jsonData)

		// Refresh IP assignment TTL
		assignmentKey := fmt.Sprintf("ip:user:%s:%s", node.TenantID, sess.UserId)
		reverseKey := fmt.Sprintf("ip:allocated:%s:%s", node.TenantID, sess.IpAddress)
		s.cache.Expire(ctx, assignmentKey, 1*time.Hour)
		s.cache.Expire(ctx, reverseKey, 1*time.Hour)
	}

	return nil
}

func (s *NodeService) ListSessions(tenantID uuid.UUID, nodeID uuid.UUID) ([]map[string]interface{}, error) {
	if s.cache == nil {
		return nil, nil
	}

	// Verify node belongs to tenant
	var node models.Node
	if err := s.db.Scopes(models.TenantScope(tenantID)).First(&node, "id = ?", nodeID).Error; err != nil {
		return nil, err
	}

	ctx := context.Background()
	sessionsKey := fmt.Sprintf("node:sessions:%s", nodeID)

	data, err := s.cache.HGetAll(ctx, sessionsKey).Result()
	if err != nil {
		return nil, err
	}

	var sessions []map[string]interface{}
	for _, val := range data {
		var session map[string]interface{}
		if err := json.Unmarshal([]byte(val), &session); err == nil {
			sessions = append(sessions, session)
		}
	}

	return sessions, nil
}

func (s *NodeService) UpdateHeartbeat(nodeID uuid.UUID) error {
	if s.cache == nil {
		return nil
	}

	ctx := context.Background()
	key := fmt.Sprintf("node:liveness:%s", nodeID.String())

	// Set liveness key with 60 second expiration
	err := s.cache.Set(ctx, key, "1", 60*time.Second).Err()
	if err != nil {
		return err
	}

	// Also update last_seen_at in DB occasionally or on every heartbeat?
	// To minimize DB load, we can skip updating DB every heartbeat since we have Valkey.
	// But the user might want LastSeenAt to be persisted.
	// Let's update it in DB as well but maybe via a background job or just here.
	return s.db.Model(&models.Node{}).Where("id = ?", nodeID).Update("last_seen_at", time.Now()).Error
}

func (s *NodeService) ListNodeSkus() ([]models.NodeSku, error) {
	var skus []models.NodeSku
	if err := s.db.Find(&skus).Error; err != nil {
		return nil, err
	}
	return skus, nil
}

func (s *NodeService) CreateNode(tenantID uuid.UUID, name string, skuID uuid.UUID, clientCIDR string) (*models.Node, error) {
	node := models.Node{
		BaseTenant: models.BaseTenant{TenantID: tenantID},
		Name:       name,
		Status:     "PENDING_REGISTRATION", // New status
		IsActive:   true,
		NodeSkuID:  skuID,
		ClientCIDR: clientCIDR,
		AuthToken:  nil, // Token generated on registration
	}

	if err := s.db.Create(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
}

// RegisterGateway handles the gateway enrollment process
func (s *NodeService) RegisterGateway(nodeID uuid.UUID, hostname string, deviceHash string, ipAddress string, version string) (string, error) {
	var node models.Node
	if err := s.db.First(&node, "id = ?", nodeID).Error; err != nil {
		return "", errors.New("gateway node not found")
	}

	// 1. If already registered (AuthToken exists)
	if node.AuthToken != nil && *node.AuthToken != "" {
		if node.DeviceHash == deviceHash {
			// Update version and last seen on re-registration/reconnect
			s.db.Model(&node).Updates(map[string]interface{}{
				"gateway_version": version,
				"ip_address":      ipAddress,
				"last_seen_at":    time.Now(),
			})
			_ = s.UpdateHeartbeat(node.ID)
			return *node.AuthToken, nil
		}
		return "", errors.New("gateway already registered with different device")
	}

	// 2. Not registered yet - Generate Token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	token := hex.EncodeToString(bytes)

	node.AuthToken = &token
	node.Hostname = hostname
	node.IPAddress = ipAddress
	node.DeviceHash = deviceHash
	node.GatewayVersion = version
	node.Status = "CONNECTED"
	now := time.Now()
	node.LastSeenAt = &now

	updates := map[string]interface{}{
		"auth_token":      token,
		"hostname":        hostname,
		"ip_address":      ipAddress,
		"device_hash":     deviceHash,
		"gateway_version": version,
		"status":          "CONNECTED",
		"last_seen_at":    now,
	}

	if err := s.db.Model(&node).Updates(updates).Error; err != nil {
		return "", err
	}

	_ = s.UpdateHeartbeat(node.ID)

	return token, nil
}

func (s *NodeService) DeleteNode(tenantID uuid.UUID, nodeID uuid.UUID) error {
	result := s.db.Scopes(models.TenantScope(tenantID)).Delete(&models.Node{}, "id = ?", nodeID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("node not found")
	}
	return nil
}

func (s *NodeService) GetNodeByToken(token string) (*models.Node, error) {
	var node models.Node
	if err := s.db.Preload("NodeSku").First(&node, "auth_token = ?", token).Error; err != nil {
		return nil, errors.New("invalid token")
	}
	return &node, nil
}
