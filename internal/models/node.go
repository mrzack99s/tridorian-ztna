package models

import (
	"time"

	"github.com/google/uuid"
)

type Node struct {
	BaseTenant
	BaseModel

	// Identity
	Name         string `gorm:"size:255;not null" json:"name,omitempty"`
	Hostname     string `gorm:"size:255" json:"hostname,omitempty"`
	AgentVersion string `gorm:"size:50" json:"agent_version,omitempty"`

	// Authentication & Security
	AuthToken string `gorm:"uniqueIndex;not null" json:"auth_token,omitempty"`

	// License
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `gorm:"default:true" json:"is_active,omitempty"`

	// Status (Heartbeat)
	Status        string     `gorm:"size:20;default:'OFFLINE'" json:"status,omitempty"` // ONLINE, OFFLINE, ERROR
	LastSeenAt    *time.Time `json:"last_seen_at,omitempty"`
	ConfigPending bool       `gorm:"default:false;not null" json:"config_pending,omitempty"`

	AccessPolicies []AccessPolicy `gorm:"many2many:access_policy_nodes;" json:"access_policies,omitempty"`
	SignInPolicies []SignInPolicy `gorm:"many2many:sign_in_policy_nodes;" json:"sign_in_policies,omitempty"`

	// ผูกว่า Node นี้เกิดจาก License ไหน
	NodeSkuID uuid.UUID `gorm:"index" json:"node_sku_id,omitempty"`
	NodeSku   NodeSku   `json:"node_sku,omitempty"`
}
