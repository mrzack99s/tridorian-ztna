package firewall

import (
	"encoding/binary"
	"log"
	"net/netip"
	"sort"
	"strings"
)

var CurrentConfig *AgentExecutionConfig
var Engine *EngineType

type EngineType struct {
	Rules []ParsedRule
}

type ParsedRule struct {
	Name           string
	Priority       int
	Allow          bool
	SourceType     string
	SourceIdentity string
	DestType       string
	DestNet        []netip.Prefix
	DestIdentity   string
}

type ValType struct {
	Addr     netip.Addr
	Identity string
}

func NewEngine() {
	var parsedRules []ParsedRule

	for _, r := range CurrentConfig.ConditionalAccessPolicies {
		isAllow := strings.ToUpper(r.Action) == "ALLOW"
		var dstPrefixes []netip.Prefix
		if r.DestinationTagType == "CIDR" {
			trimedStr := strings.TrimSpace(r.DestinationMatchValue)
			for _, v := range strings.Split(trimedStr, ",") {
				dstPrefix, err := netip.ParsePrefix(v)
				if err != nil {
					log.Printf("⚠️ Invalid Destination CIDR in rule %s: %v", r.Name, err)
					continue
				}
				dstPrefixes = append(dstPrefixes, dstPrefix)
			}
		}

		parsedRules = append(parsedRules, ParsedRule{
			Name:           r.Name,
			Priority:       r.Priority,
			Allow:          isAllow,
			SourceType:     r.SourceTagType,
			SourceIdentity: r.SourceMatchValue,
			DestType:       r.DestinationTagType,
			DestNet:        dstPrefixes,
			DestIdentity:   r.DestinationMatchValue,
		})
	}

	sort.Slice(parsedRules, func(i, j int) bool {
		return parsedRules[i].Priority > parsedRules[j].Priority
	})

	log.Printf("✅ Firewall Engine loaded %d rules", len(parsedRules))
	Engine = &EngineType{Rules: parsedRules}
}

func MatchSNI(packet []byte, sni string) SniResponseType {
	if len(packet) < 40 {
		return SNI_RESPONSE_BYPASS
	}
	if packet[9] != 6 {
		return SNI_RESPONSE_BYPASS
	}
	ihl := int(packet[0]&0x0F) * 4
	if len(packet) < ihl+20 {
		return SNI_RESPONSE_BYPASS
	}
	tcpPayload := packet[ihl:]
	dataOffset := int(tcpPayload[12]>>4) * 4
	if len(tcpPayload) < dataOffset {
		return SNI_RESPONSE_BYPASS
	}
	payload := tcpPayload[dataOffset:]
	if len(payload) < 5 {
		return SNI_RESPONSE_BYPASS
	}
	if payload[0] != 0x16 || payload[1] != 0x03 {
		return SNI_RESPONSE_BYPASS
	}
	if len(payload) < 9 || payload[5] != 0x01 {
		return SNI_RESPONSE_BYPASS
	}
	cursor := 9
	cursor += 34
	if cursor >= len(payload) {
		return SNI_RESPONSE_BYPASS
	}
	sessionIDLen := int(payload[cursor])
	cursor += 1 + sessionIDLen
	if cursor >= len(payload) {
		return SNI_RESPONSE_BYPASS
	}
	if cursor+2 > len(payload) {
		return SNI_RESPONSE_BYPASS
	}
	cipherLen := int(binary.BigEndian.Uint16(payload[cursor : cursor+2]))
	cursor += 2 + cipherLen
	if cursor >= len(payload) {
		return SNI_RESPONSE_BYPASS
	}
	if cursor+1 > len(payload) {
		return SNI_RESPONSE_BYPASS
	}
	compLen := int(payload[cursor])
	cursor += 1 + compLen
	if cursor >= len(payload) {
		return SNI_RESPONSE_BYPASS
	}
	if cursor+2 > len(payload) {
		return SNI_RESPONSE_BYPASS
	}
	extBlockLen := int(binary.BigEndian.Uint16(payload[cursor : cursor+2]))
	cursor += 2
	endOfExt := min(cursor+extBlockLen, len(payload))
	for cursor < endOfExt {
		if cursor+4 > endOfExt {
			break
		}
		extType := binary.BigEndian.Uint16(payload[cursor : cursor+2])
		extLen := int(binary.BigEndian.Uint16(payload[cursor+2 : cursor+4]))
		cursor += 4
		if extType == 0x0000 {
			if cursor+5 > endOfExt {
				break
			}
			sniLen := int(binary.BigEndian.Uint16(payload[cursor+3 : cursor+5]))
			startSNI := cursor + 5
			endSNI := startSNI + sniLen
			if endSNI <= endOfExt {
				domain := string(payload[startSNI:endSNI])
				if domain == sni {
					return SNI_RESPONSE_MATCH
				}
			}
			return SNI_RESPONSE_UNMATCH
		}
		cursor += extLen
	}
	return SNI_RESPONSE_NOT_FOUND
}

func (e *EngineType) IsAllowed(packetData []byte, sourceVal ValType, destVal ValType) bool {
	for _, rule := range e.Rules {
		matchSrc := false
		switch rule.SourceType {
		case "Identity":
			if rule.SourceIdentity == sourceVal.Identity {
				matchSrc = true
			}
		}
		if !matchSrc {
			continue
		}
		matchDst := false
		switch rule.DestType {
		case "CIDR":
			for _, dst := range rule.DestNet {
				if dst.Contains(destVal.Addr) {
					matchDst = true
					break
				}
			}
		case "SNI":
			switch MatchSNI(packetData, rule.DestIdentity) {
			case SNI_RESPONSE_BYPASS, SNI_RESPONSE_MATCH:
				matchDst = true
			default:
				matchDst = false
			}
			destVal.Identity = rule.DestIdentity
		}
		if !matchDst {
			continue
		}
		if rule.Allow {
			return true
		} else {
			log.Printf("🔴 Denied by rule: %s (Src: %s -> Dst: %s)", rule.Name, sourceVal.Identity, destVal.Identity)
			return false
		}
	}

	log.Printf("🛡️ Blocked by Default Policy (Src: %s -> Dst: %s)", sourceVal.Identity, destVal.Identity)
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
