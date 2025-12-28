package models

import (
	"github.com/google/uuid"
)

type PolicyCondition struct {
	BaseModel
	BaseTenant

	NodeID uuid.UUID `json:"node_id,omitempty"`

	Type string `gorm:"size:50" json:"type,omitempty"` // "User", "Network", "Device", ...

	Field string `gorm:"size:100" json:"field,omitempty"` // "group", "ip", "os"
	Op    string `gorm:"size:20" json:"op,omitempty"`     // "equals", "in", "cidr", "contains"

	Value string `gorm:"size:255" json:"value,omitempty"` // JSON / string / array
}
