package types

type AgentExecutionConfig struct {
	// 1. Identity Part (มาจาก Node Table)
	NodeID   string `json:"node_id"`
	Hostname string `json:"hostname"`

	VPNServer struct {
		Port        string `json:"port"`
		HostAddress string `json:"host_address"`
	} `json:"vpn_server"`

	IPAM struct {
		CIDR string `json:"cidr"`
	} `json:"ipam"`

	ConditionalAccessPolicies []ConditionalAccessPolicy `json:"conditional_access_policies"`
}

type ConditionalAccessPolicy struct {
	Name        string `gorm:"size:255;not null"`
	Description string

	Priority int `gorm:"default:0"`

	// Condition Logic
	// TagType: "SourceCIDR", "Identity", "DeviceOS"
	NetworkTagType string `gorm:"size:50;not null"`

	// MatchValue: "192.168.1.0/24", "group:admin@domain.com", "email:user1@domain.com", "windows"
	MatchValue string `gorm:"size:255;not null"`

	// Destination: "10.200.0.5/32" หรือ "0.0.0.0/0"
	DestinationNetworkCIDR string `gorm:"size:50;not null"`

	// Action: "ALLOW", "DENY", "LOG"
	Action string `gorm:"size:20;not null"`
}
