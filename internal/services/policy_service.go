package services

import (
	"errors"

	"tridorian-ztna/internal/models"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type PolicyService struct {
	db    *gorm.DB
	cache *redis.Client
}

func NewPolicyService(db *gorm.DB, cache *redis.Client) *PolicyService {
	return &PolicyService{db: db, cache: cache}
}

func (s *PolicyService) ListAccessPolicies(tenantID uuid.UUID) ([]models.AccessPolicy, error) {
	var policies []models.AccessPolicy
	if err := s.db.Scopes(models.TenantScope(tenantID)).
		Preload("DestinationApp").
		Preload("DestinationApp.CIDRs").
		Preload("Nodes").
		Order("priority asc").
		Find(&policies).Error; err != nil {
		return nil, err
	}

	for i := range policies {
		if policies[i].RootNodeID != nil {
			node, err := s.LoadNodeRecursive(*policies[i].RootNodeID)
			if err == nil {
				policies[i].RootNode = *node
			}
		}
	}
	return policies, nil
}

func (s *PolicyService) ListAccessPoliciesByNodeID(tenantID uuid.UUID, nodeID uuid.UUID) ([]models.AccessPolicy, error) {
	var policies []models.AccessPolicy
	// Find policies where the node is in the association
	if err := s.db.Scopes(models.TenantScope(tenantID)).
		Preload("DestinationApp").
		Preload("DestinationApp.CIDRs").
		Preload("Nodes").
		Joins("JOIN access_policy_nodes ON access_policy_nodes.access_policy_id = access_policies.id").
		Where("access_policy_nodes.node_id = ? AND access_policies.enabled = ?", nodeID, true).
		Order("priority asc").
		Find(&policies).Error; err != nil {
		return nil, err
	}

	for i := range policies {
		if policies[i].RootNodeID != nil {
			node, err := s.LoadNodeRecursive(*policies[i].RootNodeID)
			if err == nil {
				policies[i].RootNode = *node
			}
		}
	}
	return policies, nil
}

func (s *PolicyService) CreateAccessPolicy(tenantID uuid.UUID, policy *models.AccessPolicy) (*models.AccessPolicy, error) {
	policy.TenantID = tenantID
	policy.Enabled = true

	s.setTenantIDOnTree(tenantID, &policy.RootNode)

	// Create policy first, omitting Nodes to prevent insertion of empty Node objects
	if err := s.db.Omit("Nodes").Create(policy).Error; err != nil {
		return nil, err
	}

	// Update the Many-to-Many association if provided
	if len(policy.Nodes) > 0 {
		if err := s.db.Model(policy).Association("Nodes").Replace(policy.Nodes); err != nil {
			return nil, err
		}
	}

	return policy, nil
}

func (s *PolicyService) UpdateAccessPolicy(tenantID uuid.UUID, policy *models.AccessPolicy) (*models.AccessPolicy, error) {
	policy.TenantID = tenantID

	// 1. Get the old policy to find the root node
	var oldPolicy models.AccessPolicy
	if err := s.db.First(&oldPolicy, "id = ?", policy.ID).Error; err == nil {
		// 2. Delete the old tree recursively
		if oldPolicy.RootNodeID != nil {
			s.DeleteNodeRecursive(*oldPolicy.RootNodeID)
		}
	}

	s.setTenantIDOnTree(tenantID, &policy.RootNode)

	// 3. Clear IDs in the new tree to force recreation
	s.clearIDsRecursive(&policy.RootNode)

	// Update the policy fields (except associations)
	if err := s.db.Omit("Nodes").Save(policy).Error; err != nil {
		return nil, err
	}

	// Check existing associations to avoid redundant updates
	var currentNodes []models.Node
	// Create a clean struct to ensure GORM only uses the ID for the lookup
	lookupPolicy := &models.AccessPolicy{}
	lookupPolicy.ID = policy.ID

	if err := s.db.Model(lookupPolicy).Association("Nodes").Find(&currentNodes); err != nil {
		return nil, err
	}

	shouldUpdateNodes := false
	if len(currentNodes) != len(policy.Nodes) {
		shouldUpdateNodes = true
	} else {
		// Maps for O(n) lookup
		existingIDs := make(map[uuid.UUID]bool)
		for _, n := range currentNodes {
			existingIDs[n.ID] = true
		}
		for _, n := range policy.Nodes {
			if !existingIDs[n.ID] {
				shouldUpdateNodes = true
				break
			}
		}
	}

	// Update the Many-to-Many association explicitly only if changed
	if shouldUpdateNodes {
		if err := s.db.Model(policy).Association("Nodes").Replace(policy.Nodes); err != nil {
			return nil, err
		}
	}

	return policy, nil
}

func (s *PolicyService) DeleteAccessPolicy(tenantID uuid.UUID, policyID uuid.UUID) error {
	result := s.db.Scopes(models.TenantScope(tenantID)).Delete(&models.AccessPolicy{}, "id = ?", policyID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("policy not found")
	}
	return nil
}

func (s *PolicyService) ListSignInPolicies(tenantID uuid.UUID) ([]models.SignInPolicy, error) {
	var policies []models.SignInPolicy
	if err := s.db.Scopes(models.TenantScope(tenantID)).
		Order("priority asc").
		Find(&policies).Error; err != nil {
		return nil, err
	}

	for i := range policies {
		if policies[i].RootNodeID != uuid.Nil {
			node, err := s.LoadNodeRecursive(policies[i].RootNodeID)
			if err == nil {
				policies[i].RootNode = *node
			}
		}
	}
	return policies, nil
}

func (s *PolicyService) LoadNodeRecursive(nodeID uuid.UUID) (*models.PolicyNode, error) {
	var node models.PolicyNode
	if err := s.db.Preload("Condition").Preload("Children").First(&node, "id = ?", nodeID).Error; err != nil {
		return nil, err
	}

	for i := range node.Children {
		child, err := s.LoadNodeRecursive(node.Children[i].ID)
		if err == nil {
			node.Children[i] = *child
		}
	}
	return &node, nil
}

func (s *PolicyService) CreateSignInPolicy(tenantID uuid.UUID, policy *models.SignInPolicy) (*models.SignInPolicy, error) {
	policy.TenantID = tenantID
	policy.Enabled = true

	s.setTenantIDOnTree(tenantID, &policy.RootNode)

	if err := s.db.Session(&gorm.Session{FullSaveAssociations: true}).Create(policy).Error; err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *PolicyService) UpdateSignInPolicy(tenantID uuid.UUID, policy *models.SignInPolicy) (*models.SignInPolicy, error) {
	policy.TenantID = tenantID

	// 1. Get the old policy to find the root node
	var oldPolicy models.SignInPolicy
	if err := s.db.First(&oldPolicy, "id = ?", policy.ID).Error; err == nil {
		// 2. Delete the old tree recursively
		s.DeleteNodeRecursive(oldPolicy.RootNodeID)
	}

	s.setTenantIDOnTree(tenantID, &policy.RootNode)

	// 3. Clear IDs in the new tree to force recreation (this is safer than trying to reconcile)
	s.clearIDsRecursive(&policy.RootNode)

	if err := s.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(policy).Error; err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *PolicyService) clearIDsRecursive(node *models.PolicyNode) {
	if node == nil {
		return
	}
	node.ID = uuid.Nil
	if node.Condition != nil {
		node.Condition.ID = uuid.Nil
	}
	for i := range node.Children {
		s.clearIDsRecursive(&node.Children[i])
	}
}

func (s *PolicyService) DeleteNodeRecursive(nodeID uuid.UUID) error {
	if nodeID == uuid.Nil {
		return nil
	}

	var node models.PolicyNode
	if err := s.db.Preload("Children").First(&node, "id = ?", nodeID).Error; err != nil {
		return err
	}

	// Delete children first
	for _, child := range node.Children {
		s.DeleteNodeRecursive(child.ID)
	}

	// Delete condition
	s.db.Where("node_id = ?", nodeID).Delete(&models.PolicyCondition{})

	// Delete node itself
	return s.db.Delete(&node).Error
}

func (s *PolicyService) setTenantIDOnTree(tenantID uuid.UUID, node *models.PolicyNode) {
	if node == nil {
		return
	}
	node.TenantID = tenantID
	if node.Condition != nil {
		node.Condition.TenantID = tenantID
	}
	for i := range node.Children {
		s.setTenantIDOnTree(tenantID, &node.Children[i])
	}
}

func (s *PolicyService) DeleteSignInPolicy(tenantID uuid.UUID, policyID uuid.UUID) error {
	var policy models.SignInPolicy
	if err := s.db.Scopes(models.TenantScope(tenantID)).First(&policy, "id = ?", policyID).Error; err != nil {
		return err
	}

	// Delete root node and tree
	s.DeleteNodeRecursive(policy.RootNodeID)

	result := s.db.Scopes(models.TenantScope(tenantID)).Delete(&models.SignInPolicy{}, "id = ?", policyID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("policy not found")
	}
	return nil
}
