package firewall

import (
	"encoding/binary"
	"fmt"
	"log"
	"net/netip"
	"sort"
	"strings"
)

var CurrentConfig *AgentExecutionConfig
var Engine *EngineType

type EngineType struct {
	DefaultRule ConditionalAccessDefaultPolicies

	Rules []ParsedRule
}

type ParsedRule struct {
	Name     string
	Priority int
	Allow    bool // True = ALLOW, False = DENY

	// Criteria
	SourceType     string //  "Identity", "DeviceOS"
	SourceIdentity string

	DestType     string //  "CIDR, SNI, Tag"
	DestNet      []netip.Prefix
	DestIdentity string
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

		// Append เข้า List
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

		// if parsedRules[i].Priority != parsedRules[j].Priority {
		return parsedRules[i].Priority > parsedRules[j].Priority
		// }

		// dstBitsI := parsedRules[i].DestNet.Bits()
		// dstBitsJ := parsedRules[j].DestNet.Bits()

		// return dstBitsI > dstBitsJ
	})

	log.Printf("✅ Firewall Engine loaded %d rules", len(parsedRules))
	Engine = &EngineType{Rules: parsedRules, DefaultRule: CurrentConfig.ConditionalAccessDefaultPolicies}
}

func MatchSNI(packet []byte, sni string) SniResponseType {
	if len(packet) < 40 {
		return SNI_RESPONSE_BYPASS
	}

	// Check TCP Protocol (Byte 9 of IP Header)
	if packet[9] != 6 {
		return SNI_RESPONSE_BYPASS
	}

	//  IP Header (IHL)
	ihl := int(packet[0]&0x0F) * 4
	if len(packet) < ihl+20 {
		return SNI_RESPONSE_BYPASS
	}

	// 2. TCP Header
	tcpPayload := packet[ihl:]
	// Data Offset
	dataOffset := int(tcpPayload[12]>>4) * 4
	if len(tcpPayload) < dataOffset {
		return SNI_RESPONSE_BYPASS
	}

	// 3. TLS Record Layer
	payload := tcpPayload[dataOffset:]
	if len(payload) < 5 {
		return SNI_RESPONSE_BYPASS
	}

	// TLS Record Type: 0x16 (Handshake)
	// TLS Version: 0x03 (SSL 3.0 / TLS 1.x)
	if payload[0] != 0x16 || payload[1] != 0x03 {
		return SNI_RESPONSE_BYPASS // non TLS Handshake
	}

	// 4. Handshake Layer
	// Handshake Type: 0x01 (Client Hello)
	// skip Record Header 5 bytes
	if len(payload) < 9 || payload[5] != 0x01 {
		return SNI_RESPONSE_BYPASS
	}

	// --- Deep Dive into Client Hello (Zero Copy Parsing) ---
	// Pointer start after Handshake Header (5 + 4 = 9)
	cursor := 9

	// Skip Protocol Version (2 bytes) + Random (32 bytes)
	cursor += 34
	if cursor >= len(payload) {
		return SNI_RESPONSE_BYPASS
	}

	// Skip Session ID
	sessionIDLen := int(payload[cursor])
	cursor += 1 + sessionIDLen
	if cursor >= len(payload) {
		return SNI_RESPONSE_BYPASS
	}

	// Skip Cipher Suites
	if cursor+2 > len(payload) {
		return SNI_RESPONSE_BYPASS
	}
	cipherLen := int(binary.BigEndian.Uint16(payload[cursor : cursor+2]))
	cursor += 2 + cipherLen
	if cursor >= len(payload) {
		return SNI_RESPONSE_BYPASS
	}

	// Skip Compression Methods
	if cursor+1 > len(payload) {
		return SNI_RESPONSE_BYPASS
	}
	compLen := int(payload[cursor])
	cursor += 1 + compLen
	if cursor >= len(payload) {
		return SNI_RESPONSE_BYPASS
	}

	// --- Extensions Block ---
	if cursor+2 > len(payload) {
		return SNI_RESPONSE_BYPASS
	}
	extBlockLen := int(binary.BigEndian.Uint16(payload[cursor : cursor+2]))
	cursor += 2

	endOfExt := min(cursor+extBlockLen, len(payload))

	// Loop find Extension ID 0x0000 (SNI)
	for cursor < endOfExt {
		if cursor+4 > endOfExt {
			break
		}
		extType := binary.BigEndian.Uint16(payload[cursor : cursor+2])
		extLen := int(binary.BigEndian.Uint16(payload[cursor+2 : cursor+4]))
		cursor += 4

		if extType == 0x0000 { // found SNI!
			// extract Server Name
			if cursor+5 > endOfExt {
				break
			}
			// ข้าม List Len (2) + Type (1)
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

	// 2. Loop Through Rules (Sorted)
	for _, rule := range e.Rules {

		// --- Check Source ---
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

		// --- Check Destination ---
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

	fmt.Println(e.DefaultRule.BlockByDefault)
	if e.DefaultRule.BlockByDefault {
		log.Printf("🛡️ Blocked by Default Policy (Src: %s -> Dst: %s)", sourceVal.Identity, destVal.Identity)
		return false
	}

	return true

}

// Helper: ตัด Port ทิ้งถ้ามี
func parseIP(s string) (netip.Addr, error) {
	// ลอง Parse แบบมี Port ก่อน (1.1.1.1:80)
	if ap, err := netip.ParseAddrPort(s); err == nil {
		return ap.Addr(), nil
	}
	// ถ้าไม่ผ่าน ลอง Parse แบบ IP เพียวๆ
	return netip.ParseAddr(s)
}
