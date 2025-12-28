package models

import "github.com/google/uuid"

type AccessPolicy struct {
	BasePolicy
	BaseTenant

	ResourceType    string `json:"resource_type,omitempty"`
	ResourceID      string `json:"resource_id,omitempty"`
	DestinationType string `json:"destination_type,omitempty"` // "cidr", "app", "sni"

	// For type "cidr"
	DestinationCIDR string `gorm:"column:destination_cidr" json:"destination_cidr,omitempty"`

	// For type "app"
	DestinationAppID *uuid.UUID   `gorm:"type:uuid" json:"destination_app_id,omitempty"`
	DestinationApp   *Application `json:"destination_app,omitempty"`

	// For type "sni"
	DestinationSNI string `json:"destination_sni,omitempty"`

	RootNodeID *uuid.UUID `json:"root_node_id,omitempty"`
	RootNode   PolicyNode `json:"root_node,omitempty"`

	Effect string `json:"effect,omitempty"` // Allow / Deny / Limit

	Nodes []Node `gorm:"many2many:access_policy_nodes;" json:"nodes,omitempty"`
}
