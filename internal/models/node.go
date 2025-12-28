package models

import (
	"time"

	"github.com/google/uuid"
)

type Node struct {
	BaseTenant
	BaseModel

	// Identity
	Name           string `gorm:"size:255;not null" json:"name,omitempty"`
	Hostname       string `gorm:"size:255" json:"hostname,omitempty"`
	IPAddress      string `gorm:"size:50" json:"ip_address,omitempty"` // Public IP of the gateway
	GatewayVersion string `gorm:"size:50" json:"gateway_version,omitempty"`
	ClientCIDR     string `gorm:"size:50;column:client_cidr" json:"client_cidr,omitempty"`
	DeviceHash     string `gorm:"size:255" json:"device_hash,omitempty"`

	// Authentication & Security
	AuthToken    *string `gorm:"uniqueIndex" json:"auth_token,omitempty"`
	PublicKeyPEM string  `gorm:"type:text" json:"public_key_pem,omitempty"`

	// License
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	IsActive  bool       `gorm:"default:true" json:"is_active,omitempty"`

	// Status (Heartbeat)
	Status        string     `gorm:"size:20;default:'OFFLINE'" json:"status,omitempty"` // ONLINE, OFFLINE, ERROR
	LastSeenAt    *time.Time `json:"last_seen_at,omitempty"`
	ConfigPending bool       `gorm:"default:false;not null" json:"config_pending,omitempty"`

	AccessPolicies []AccessPolicy `gorm:"many2many:access_policy_nodes;" json:"access_policies,omitempty"`

	NodeSkuID uuid.UUID `gorm:"index" json:"node_sku_id,omitempty"`
	NodeSku   NodeSku   `json:"node_sku,omitempty"`
}
