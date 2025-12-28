package models

import (
	"time"

	"github.com/google/uuid"
)

type AccessPolicyNode struct {
	AccessPolicyID uuid.UUID `gorm:"index" json:"access_policy_id,omitempty"`
	NodeID         uuid.UUID `gorm:"index" json:"node_id,omitempty"`

	AccessPolicy AccessPolicy `json:"access_policy,omitempty"`
	Node         Node         `json:"node,omitempty"`

	CreatedAt time.Time `json:"created_at,omitempty"`
}
