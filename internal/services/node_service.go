package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"

	"tridorian-ztna/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NodeService struct {
	db *gorm.DB
}

func NewNodeService(db *gorm.DB) *NodeService {
	return &NodeService{db: db}
}

func (s *NodeService) ListNodes(tenantID uuid.UUID) ([]models.Node, error) {
	var nodes []models.Node
	if err := s.db.Scopes(models.TenantScope(tenantID)).Find(&nodes).Error; err != nil {
		return nil, err
	}
	return nodes, nil
}

func (s *NodeService) CreateNode(tenantID uuid.UUID, name string) (*models.Node, error) {
	// Generate random Auth Token
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return nil, err
	}
	token := hex.EncodeToString(bytes)

	node := models.Node{
		BaseTenant: models.BaseTenant{TenantID: tenantID},
		Name:       name,
		AuthToken:  token,
		Status:     "OFFLINE",
		IsActive:   true,
	}

	if err := s.db.Create(&node).Error; err != nil {
		return nil, err
	}
	return &node, nil
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
