package models

import (
	"github.com/google/uuid"
)

type SignInPolicy struct {
	BasePolicy
	BaseTenant

	RootNodeID uuid.UUID  `json:"root_node_id,omitempty"`
	RootNode   PolicyNode `json:"root_node,omitempty"`

	Block bool   `json:"block,omitempty"`
	Stage string `gorm:"size:20;default:'pre_auth'" json:"stage,omitempty"` // "pre_auth" | "post_auth"
}
