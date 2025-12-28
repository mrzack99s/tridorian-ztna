package services

import (
	"crypto/sha256"
	"fmt"
	"strings"
	"tridorian-ztna/internal/models"
	pb "tridorian-ztna/internal/proto/gateway/v1"
)

// GenerateGatewayPolicies converts internal AccessPolicy models into the Protobuf format expected by the Gateway.
// It flattens the logical tree of conditions into a list of specific rules that the Gateway's simple firewall engine can enforce.
func GenerateGatewayPolicies(policies []models.AccessPolicy) []*pb.GetConfigResponse_Policy {
	var result []*pb.GetConfigResponse_Policy

	for _, p := range policies {
		// Flatten the tree logic regarding Source (Identity, Device, etc.)
		sourceRules := flattenPolicyNode(p.RootNode)

		// If no source rules are returned (e.g. empty tree or unsupported conditions),
		// we might optionally create a "catch-all" or skip.
		if len(sourceRules) == 0 {
			continue
		}

		for _, src := range sourceRules {
			// Map Destination
			destTag := ""
			destVal := ""

			switch p.DestinationType {
			case "cidr":
				destTag = "CIDR"
				destVal = p.DestinationCIDR
			case "sni":
				destTag = "SNI"
				destVal = p.DestinationSNI
			case "app":
				// Support for Applications (multiple CIDRs)
				if p.DestinationApp != nil && len(p.DestinationApp.CIDRs) > 0 {
					var cidrs []string
					for _, c := range p.DestinationApp.CIDRs {
						cidrs = append(cidrs, c.CIDR)
					}
					destTag = "CIDR"
					// The Gateway Firewall Engine supports comma-separated CIDRs in a single rule
					destVal = strings.Join(cidrs, ",")
				}
			}

			if destTag == "" {
				continue
			}

			// Map Action
			action := "ALLOW" // Default to ALLOW for Access Policies unless specified
			if strings.EqualFold(p.Effect, "deny") {
				action = "DENY"
			}

			gp := &pb.GetConfigResponse_Policy{
				Name:                  p.Name,
				Action:                action,
				Priority:              int32(p.Priority),
				SourceTagType:         src.TagType,
				SourceMatchValue:      src.Value,
				DestinationTagType:    destTag,
				DestinationMatchValue: destVal,
			}

			result = append(result, gp)
		}
	}
	return result
}

type SourceRule struct {
	TagType string
	Value   string
}

// flattenPolicyNode traverses the condition tree and returns a list of valid Source Rules (OR logic).
func flattenPolicyNode(node models.PolicyNode) []SourceRule {
	var rules []SourceRule

	// Leaf Node with Condition
	if node.Condition != nil {
		t, v := mapCondition(*node.Condition)
		if t != "" {
			rules = append(rules, SourceRule{TagType: t, Value: v})
		}
		return rules
	}

	// Recursive Children
	// We intentionally flatten both AND/OR into a list.
	if len(node.Children) > 0 {
		for _, child := range node.Children {
			rules = append(rules, flattenPolicyNode(child)...)
		}
	}

	return rules
}

func mapCondition(c models.PolicyCondition) (string, string) {
	switch c.Type {
	case "User":
		if c.Field == "group" {
			if !strings.HasPrefix(c.Value, "group:") {
				return "Identity", "group:" + c.Value
			}
		}
		return "Identity", c.Value
	case "Group":
		if !strings.HasPrefix(c.Value, "group:") {
			return "Identity", "group:" + c.Value
		}
		return "Identity", c.Value
	case "Device":
		if c.Field == "os" {
			return "DeviceOS", c.Value
		}
	}
	return "", ""
}

func CalculateConfigHash(policies []*pb.GetConfigResponse_Policy) string {
	if len(policies) == 0 {
		return "empty"
	}
	// Simple string concatenation of all fields to generate a hash
	var builder strings.Builder
	for _, p := range policies {
		builder.WriteString(p.Name)
		builder.WriteString(p.Action)
		builder.WriteString(fmt.Sprintf("%d", p.Priority))
		builder.WriteString(p.SourceTagType)
		builder.WriteString(p.SourceMatchValue)
		builder.WriteString(p.DestinationTagType)
		builder.WriteString(p.DestinationMatchValue)
		builder.WriteString("|")
	}

	sum := sha256.Sum256([]byte(builder.String()))
	return fmt.Sprintf("%x", sum)
}
