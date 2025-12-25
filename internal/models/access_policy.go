package models

import "github.com/google/uuid"

type AccessPolicy struct {
	BasePolicy
	BaseTenant

	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"`

	RootNodeID uuid.UUID  `json:"root_node_id,omitempty"`
	RootNode   PolicyNode `json:"root_node,omitempty"`

	Effect string `json:"effect,omitempty"` // Allow / Deny / Limit

	Nodes []PolicyNode `gorm:"many2many:access_policy_nodes;" json:"nodes,omitempty"`
}
