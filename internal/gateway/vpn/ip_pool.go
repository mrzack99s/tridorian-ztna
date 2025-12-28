package vpn

import (
	"errors"
	"fmt"
	"net"
	"sync"
)

// FastIPPool manages IP address allocation from a CIDR block using a channel for O(1) allocation.
type FastIPPool struct {
	CIDR         string
	AvailableIPs chan string
	Used         sync.Map          // Currently active sessions
	Reservations map[string]string // Sticky IP map: identity -> ip
	mu           sync.Mutex
}

// NewFastIPPool creates a new IP pool.
// It generates IPs covering a large range (effectively /16 if loop goes to 255) to the channel.
// Note: Ensure the CIDR passed matches what is expected, although this implementation mostly uses the base IP and iterates.
func NewFastIPPool(cidr string) *FastIPPool {
	pool := &FastIPPool{
		AvailableIPs: make(chan string, 65536),
		CIDR:         cidr,
		Reservations: make(map[string]string),
	}
	go func() {
		ip, ipnet, _ := net.ParseCIDR(cidr)
		prefix, _ := ipnet.Mask.Size()
		ip = ip.To4()

		// Generate IPs.
		// The user example had `for i := range 255`.
		// If CIDR is /23 (10.8.0.0/23), valid third octets are 0 and 1.
		// If we want to support up to 65k IPs as per comment, we need /16.
		// We will adapt to follow the logic: iterate 0..255 for 3rd octet.
		for i := 0; i < 255; i++ {
			for j := 2; j < 254; j++ {
				ip[2], ip[3] = byte(i), byte(j)
				pool.AvailableIPs <- ip.String() + "/" + fmt.Sprint(prefix)
			}
		}
	}()
	return pool
}

func (p *FastIPPool) GetHostIPAddress() string {
	ip, ipnet, _ := net.ParseCIDR(p.CIDR)
	prefix, _ := ipnet.Mask.Size()
	ip = ip.To4()
	ip[3] = byte(1)

	return ip.String() + "/" + fmt.Sprint(prefix)
}

// Assign an IP to an email/identity
func (p *FastIPPool) Assign(identity string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 1. Check if user has a sticky IP
	if ip, ok := p.Reservations[identity]; ok {
		p.Used.Store(ip, identity)
		return ip, nil
	}

	// 2. Allocate new IP
	select {
	case ip := <-p.AvailableIPs:
		p.Reservations[identity] = ip
		p.Used.Store(ip, identity)
		return ip, nil
	default:
		return "", errors.New("IP Pool Empty")
	}
}

func (p *FastIPPool) Release(ip string) {
	p.Used.Delete(ip)
	// Sticky IP: Do not return to AvailableIPs.
	// The IP remains reserved in p.Reservations for this user.
}
