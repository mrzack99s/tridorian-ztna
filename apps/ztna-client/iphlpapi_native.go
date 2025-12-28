//go:build windows

package main

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

// --- 1. DLL Loading ---
var (
	modIphlpapi = syscall.NewLazyDLL("iphlpapi.dll")

	procConvertInterfaceLuidToIndex     = modIphlpapi.NewProc("ConvertInterfaceLuidToIndex")
	procInitializeUnicastIpAddressEntry = modIphlpapi.NewProc("InitializeUnicastIpAddressEntry")
	procCreateUnicastIpAddressEntry     = modIphlpapi.NewProc("CreateUnicastIpAddressEntry")
	procInitializeIpForwardEntry        = modIphlpapi.NewProc("InitializeIpForwardEntry")
	procCreateIpForwardEntry2           = modIphlpapi.NewProc("CreateIpForwardEntry2")
	procDeleteIpForwardEntry2           = modIphlpapi.NewProc("DeleteIpForwardEntry2")
)

// --- 2. Struct Definitions ---

type LUID uint64

// SOCKADDR_INET: Union in C
type RawSockAddrInet struct {
	Family uint16
	Data   [26]byte
}

// MIB_UNICASTIPADDRESS_ROW
type MibUnicastIpAddressRow struct {
	Address            RawSockAddrInet // 28 bytes
	_                  [4]byte         // Padding
	InterfaceLuid      LUID
	InterfaceIndex     uint32
	PrefixOrigin       int32 // Enum
	SuffixOrigin       int32 // Enum
	ValidLifetime      uint32
	PreferredLifetime  uint32
	OnLinkPrefixLength uint8
	SkipAsSource       bool
	DadState           int32 // Enum
	ScopeId            uint32
	CreationTimeStamp  int64
}

// MIB_IPFORWARD_ROW2
type MibIpForwardRow2 struct {
	InterfaceLuid        LUID
	InterfaceIndex       uint32
	DestinationPrefix    IpAddressPrefix
	NextHop              RawSockAddrInet
	SitePrefixLength     uint8
	ValidLifetime        uint32
	PreferredLifetime    uint32
	Metric               uint32
	Protocol             int32
	Loopback             bool
	AutoconfigureAddress bool
	Publish              bool
	Immortal             bool
	Age                  uint32
	Origin               int32
}

type IpAddressPrefix struct {
	Prefix       RawSockAddrInet
	PrefixLength uint8
	_            [3]byte // Padding
}

// --- 3. Helper Functions ---

func SetAdapterIP(luid LUID, ipStr string) error {
	ip, ipNet, err := net.ParseCIDR(ipStr)
	if err != nil {
		return err
	}
	ipv4 := ip.To4()
	ones, _ := ipNet.Mask.Size()

	var row MibUnicastIpAddressRow

	procInitializeUnicastIpAddressEntry.Call(uintptr(unsafe.Pointer(&row)))

	row.InterfaceLuid = luid
	row.Address.Family = 2 // AF_INET

	copy(row.Address.Data[2:], ipv4)

	row.OnLinkPrefixLength = uint8(ones)
	row.PrefixOrigin = 3 // Manual
	row.SuffixOrigin = 3 // Manual
	row.ValidLifetime = 0xffffffff
	row.PreferredLifetime = 0xffffffff
	row.DadState = 4 // IpDadStatePreferred (4)

	ret, _, _ := procCreateUnicastIpAddressEntry.Call(uintptr(unsafe.Pointer(&row)))
	if ret != 0 {
		if ret == 5010 { // Object already exists
			return nil
		}
		return fmt.Errorf("CreateUnicastIpAddressEntry failed: Code %d", ret)
	}
	return nil
}

func SetAdapterRoute(luid LUID, destCIDR string) error {
	_, ipNet, err := net.ParseCIDR(destCIDR)
	if err != nil {
		return err
	}
	ones, _ := ipNet.Mask.Size()
	destIP := ipNet.IP.To4()

	var ifIndex uint32
	ret, _, _ := procConvertInterfaceLuidToIndex.Call(uintptr(unsafe.Pointer(&luid)), uintptr(unsafe.Pointer(&ifIndex)))
	if ret != 0 {
		return fmt.Errorf("ConvertInterfaceLuidToIndex failed: %d", ret)
	}

	var row MibIpForwardRow2
	procInitializeIpForwardEntry.Call(uintptr(unsafe.Pointer(&row)))

	row.InterfaceLuid = luid
	row.InterfaceIndex = ifIndex

	row.DestinationPrefix.Prefix.Family = 2
	copy(row.DestinationPrefix.Prefix.Data[2:], destIP)
	row.DestinationPrefix.PrefixLength = uint8(ones)

	row.NextHop.Family = 2
	row.Metric = 0
	row.Protocol = 3 // Static

	ret, _, _ = procCreateIpForwardEntry2.Call(uintptr(unsafe.Pointer(&row)))

	if ret != 0 && ret != 5010 {
		return fmt.Errorf("CreateIpForwardEntry2 failed: Code %d", ret)
	}

	return nil
}

func RemoveWinsRoute(luid LUID, destCIDR string) error {
	_, ipNet, err := net.ParseCIDR(destCIDR)
	if err != nil {
		return err
	}
	ones, _ := ipNet.Mask.Size()
	destIP := ipNet.IP.To4()

	var ifIndex uint32
	ret, _, _ := procConvertInterfaceLuidToIndex.Call(uintptr(unsafe.Pointer(&luid)), uintptr(unsafe.Pointer(&ifIndex)))
	if ret != 0 {
		return fmt.Errorf("ConvertInterfaceLuidToIndex failed: %d", ret)
	}

	var row MibIpForwardRow2
	procInitializeIpForwardEntry.Call(uintptr(unsafe.Pointer(&row)))

	row.InterfaceLuid = luid
	row.InterfaceIndex = ifIndex

	row.DestinationPrefix.Prefix.Family = 2
	copy(row.DestinationPrefix.Prefix.Data[2:], destIP)
	row.DestinationPrefix.PrefixLength = uint8(ones)

	row.NextHop.Family = 2

	ret, _, _ = procDeleteIpForwardEntry2.Call(uintptr(unsafe.Pointer(&row)))
	if ret != 0 {
		if ret == 1168 { // Element not found
			return nil
		}
		return fmt.Errorf("DeleteIpForwardEntry2 failed: Code %d", ret)
	}
	return nil
}

// AddRoute adds a route to the adapter.
// This acts as a wrapper for SetAdapterRoute to maintain naming consistency.
func AddWinsRoute(luid LUID, destCIDR string) error {
	return SetAdapterRoute(luid, destCIDR)
}
