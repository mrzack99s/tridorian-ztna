package firewall

type SniResponseType string

type AgentExecutionConfig struct {
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

	SourceTagType    string `gorm:"size:50;not null"`
	SourceMatchValue string `gorm:"size:255;not null"`

	DestinationTagType    string `gorm:"size:50;not null"`
	DestinationMatchValue string `gorm:"size:255;not null"`

	Action string `gorm:"size:20;not null"`
}
