package domain

import "time"

type Tenant struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain" gorm:"unique"`
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GatewayNode struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	TenantID  string    `json:"tenant_id" gorm:"index"`
	Hostname  string    `json:"hostname"`
	IPAddress string    `json:"ip_address"`
	Status    string    `json:"status"` // ONLINE, OFFLINE, ERROR
	LastSeen  time.Time `json:"last_seen"`
}

type Policy struct {
	ID       string `json:"id" gorm:"primaryKey"`
	TenantID string `json:"tenant_id" gorm:"index"`
	Name     string `json:"name"`
	Rules    string `json:"rules"` // JSON or DSL
}
