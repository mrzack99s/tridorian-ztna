package geoip

import (
	"bufio"
	"encoding/binary"
	"net"
	"os"
	"sort"
	"strings"
)

type IPRange struct {
	Start uint32
	End   uint32
	Code  string
}

type GeoIP struct {
	Ranges []IPRange
}

func New() *GeoIP {
	return &GeoIP{}
}

func (g *GeoIP) Load(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var ranges []IPRange
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		startIP := net.ParseIP(parts[0])
		endIP := net.ParseIP(parts[1])
		country := parts[2]

		if startIP == nil || endIP == nil {
			continue
		}

		startFunc := ipToUint32(startIP)
		endFunc := ipToUint32(endIP)

		ranges = append(ranges, IPRange{
			Start: startFunc,
			End:   endFunc,
			Code:  country,
		})
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	// Sort explicitly just in case, though the file is usually sorted
	sort.Slice(ranges, func(i, j int) bool {
		return ranges[i].Start < ranges[j].Start
	})

	g.Ranges = ranges
	return nil
}

func (g *GeoIP) Lookup(ipStr string) string {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return ""
	}

	// Ensure IPv4
	ip = ip.To4()
	if ip == nil {
		return ""
	}

	if isPrivateIP(ip) {
		return "PRIVATE"
	}

	val := ipToUint32(ip)

	// Binary search
	idx := sort.Search(len(g.Ranges), func(i int) bool {
		return g.Ranges[i].End >= val
	})

	if idx < len(g.Ranges) && g.Ranges[idx].Start <= val {
		return g.Ranges[idx].Code
	}

	return ""
}

func ipToUint32(ip net.IP) uint32 {
	if len(ip) == 16 {
		return binary.BigEndian.Uint32(ip[12:16])
	}
	return binary.BigEndian.Uint32(ip)
}

func isPrivateIP(ip net.IP) bool {
	// 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 127.0.0.0/8
	privateBlocks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
	}

	for _, block := range privateBlocks {
		_, ipNet, _ := net.ParseCIDR(block)
		if ipNet.Contains(ip) {
			return true
		}
	}
	return false
}
