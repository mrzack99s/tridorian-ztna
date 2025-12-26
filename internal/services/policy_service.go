package services

import (
	"errors"

	"tridorian-ztna/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PolicyService struct {
	db *gorm.DB
}

func NewPolicyService(db *gorm.DB) *PolicyService {
	return &PolicyService{db: db}
}

func (s *PolicyService) ListAccessPolicies(tenantID uuid.UUID) ([]models.AccessPolicy, error) {
	var policies []models.AccessPolicy
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

func (s *PolicyService) CreateAccessPolicy(tenantID uuid.UUID, policy *models.AccessPolicy) (*models.AccessPolicy, error) {
	policy.TenantID = tenantID
	policy.Enabled = true

	s.setTenantIDOnTree(tenantID, &policy.RootNode)

	if err := s.db.Session(&gorm.Session{FullSaveAssociations: true}).Create(policy).Error; err != nil {
		return nil, err
	}
	return policy, nil
}

func (s *PolicyService) UpdateAccessPolicy(tenantID uuid.UUID, policy *models.AccessPolicy) (*models.AccessPolicy, error) {
	policy.TenantID = tenantID

	// 1. Get the old policy to find the root node
	var oldPolicy models.AccessPolicy
	if err := s.db.First(&oldPolicy, "id = ?", policy.ID).Error; err == nil {
		// 2. Delete the old tree recursively
		s.DeleteNodeRecursive(oldPolicy.RootNodeID)
	}

	s.setTenantIDOnTree(tenantID, &policy.RootNode)

	// 3. Clear IDs in the new tree to force recreation
	s.clearIDsRecursive(&policy.RootNode)

	if err := s.db.Session(&gorm.Session{FullSaveAssociations: true}).Save(policy).Error; err != nil {
		return nil, err
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
