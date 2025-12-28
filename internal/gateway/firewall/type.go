package firewall

type SniResponseType string

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

	ConditionalAccessDefaultPolicies ConditionalAccessDefaultPolicies `json:"conditional_access_default_policies"`
	ConditionalAccessPolicies        []ConditionalAccessPolicy        `json:"conditional_access_policies"`
}

type ConditionalAccessDefaultPolicies struct {
	BlockByDefault bool `json:"cidr"`
}

type ConditionalAccessPolicy struct {
	Name        string `gorm:"size:255;not null"`
	Description string

	Priority int `gorm:"default:0"`

	// Condition Logic
	// TagType: "Identity", "DeviceOS", "Geo"
	SourceTagType string `gorm:"size:50;not null"`

	// MatchValue: "group:admin@domain.com", "email:user1@domain.com", "windows"
	SourceMatchValue string `gorm:"size:255;not null"`

	// Condition Logic
	// TagType: "CIDR", "SNI"
	DestinationTagType string `gorm:"size:50;not null"`

	// MatchValue: "0.0.0.0/0", "tridorian.com"
	DestinationMatchValue string `gorm:"size:255;not null"`

	// Action: "ALLOW", "DENY", "LOG"
	Action string `gorm:"size:20;not null"`
}
