package utils

import (
	"math/big"
	"net"
	"net/http"
	"strings"
)

// GetClientIP extracts the client IP address from the request.
// It checks X-Forwarded-For and X-Real-IP headers first.
func GetClientIP(r *http.Request) string {
	// 1. X-Forwarded-For
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		parts := strings.Split(xff, ",")
		if len(parts) > 0 {
			return strings.TrimSpace(parts[0])
		}
	}

	// 2. X-Real-IP
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// 3. RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func IncrementIP(ip net.IP) net.IP {
	nextIP := make(net.IP, len(ip))
	copy(nextIP, ip)
	for j := len(nextIP) - 1; j >= 0; j-- {
		nextIP[j]++
		if nextIP[j] > 0 {
			break
		}
	}
	return nextIP
}

func IPToBigInt(ip net.IP) *big.Int {
	if ip4 := ip.To4(); ip4 != nil {
		return big.NewInt(0).SetBytes(ip4)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}

func BigIntToIP(i *big.Int, isIPv4 bool) net.IP {
	if isIPv4 {
		ip := make(net.IP, 4)
		b := i.Bytes()
		copy(ip[4-len(b):], b)
		return ip
	}
	ip := make(net.IP, 16)
	b := i.Bytes()
	copy(ip[16-len(b):], b)
	return ip
}

func GetIPRange(cidr string) (net.IP, net.IP, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, nil, err
	}

	firstIP := ipNet.IP
	lastIP := make(net.IP, len(firstIP))
	for i := 0; i < len(firstIP); i++ {
		lastIP[i] = firstIP[i] | ^ipNet.Mask[i]
	}

	return firstIP, lastIP, nil
}
