package main

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"
)

// --- 1. DLL Loading (เหมือนเดิม) ---
var (
	modIphlpapi = syscall.NewLazyDLL("iphlpapi.dll")

	procConvertInterfaceLuidToIndex     = modIphlpapi.NewProc("ConvertInterfaceLuidToIndex")
	procInitializeUnicastIpAddressEntry = modIphlpapi.NewProc("InitializeUnicastIpAddressEntry")
	procCreateUnicastIpAddressEntry     = modIphlpapi.NewProc("CreateUnicastIpAddressEntry")
	procInitializeIpForwardEntry        = modIphlpapi.NewProc("InitializeIpForwardEntry")
	procCreateIpForwardEntry2           = modIphlpapi.NewProc("CreateIpForwardEntry2")
)

// --- 2. Struct Definitions (แก้ใหม่ให้ถูกต้อง) ---

type LUID uint64

// SOCKADDR_INET: ใน C เป็น Union ต้องจองพื้นที่เท่าตัวที่ใหญ่สุด (IPv6 = 28 bytes)
// แม้เราจะใช้ IPv4 เราก็ต้องส่งก้อนนี้ไปทั้งก้อน
type RawSockAddrInet struct {
	Family uint16
	Data   [26]byte // รวมเป็น 2 + 26 = 28 bytes
}

// MIB_UNICASTIPADDRESS_ROW
// สำคัญ: Go จะ Align struct ไม่เหมือน C ในบางครั้ง
// เราต้องแน่ใจว่า LUID เริ่มต้นที่ Offset ที่ถูกต้อง
type MibUnicastIpAddressRow struct {
	Address            RawSockAddrInet // 28 bytes
	_                  [4]byte         // Padding 4 bytes เพื่อให้ LUID เริ่มที่ offset 32 (8-byte aligned)
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

	// 1. Initialize (Windows จะเติมค่า Default ให้)
	// สำคัญมาก: ต้องส่ง pointer ไป
	procInitializeUnicastIpAddressEntry.Call(uintptr(unsafe.Pointer(&row)))

	// 2. เติมค่า IPv4
	row.InterfaceLuid = luid
	row.Address.Family = 2 // AF_INET

	// Copy IPv4 (4 bytes) ไปใส่ใน Data แต่ต้องระวัง Port (ข้ามไป 2 byte แรกของ Data ถ้าเทียบกับ struct sockaddr_in)
	// Layout ของ sockaddr_in (IPv4) คือ: Family(2) + Port(2) + Addr(4) + Zero(8)
	// ใน RawSockAddrInet ของเราคือ: Family(2) + Data[0-1](Port) + Data[2-5](Addr) + ...

	copy(row.Address.Data[2:], ipv4) // Data[2] คือจุดเริ่มของ IP ใน sockaddr_in structure

	row.OnLinkPrefixLength = uint8(ones)
	row.PrefixOrigin = 3 // Manual
	row.SuffixOrigin = 3 // Manual
	row.ValidLifetime = 0xffffffff
	row.PreferredLifetime = 0xffffffff
	row.DadState = 4 // IpDadStatePreferred (4) ไม่ต้องรอ check duplicate

	// 3. Create
	ret, _, _ := procCreateUnicastIpAddressEntry.Call(uintptr(unsafe.Pointer(&row)))
	if ret != 0 {
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

	// หา Index
	var ifIndex uint32
	ret, _, _ := procConvertInterfaceLuidToIndex.Call(uintptr(luid), uintptr(unsafe.Pointer(&ifIndex)))
	if ret != 0 {
		return fmt.Errorf("ConvertInterfaceLuidToIndex failed: %d", ret)
	}

	var row MibIpForwardRow2
	procInitializeIpForwardEntry.Call(uintptr(unsafe.Pointer(&row)))

	row.InterfaceLuid = luid
	row.InterfaceIndex = ifIndex

	// Set Destination
	row.DestinationPrefix.Prefix.Family = 2
	copy(row.DestinationPrefix.Prefix.Data[2:], destIP) // Offset เดียวกับข้างบน
	row.DestinationPrefix.PrefixLength = uint8(ones)

	// Set NextHop (0.0.0.0 = On-Link)
	row.NextHop.Family = 2

	row.Metric = 0
	row.Protocol = 3 // Static

	ret, _, _ = procCreateIpForwardEntry2.Call(uintptr(unsafe.Pointer(&row)))

	// 5010 = Object Exists (ยอมรับได้)
	if ret != 0 && ret != 5010 {
		return fmt.Errorf("CreateIpForwardEntry2 failed: Code %d", ret)
	}

	return nil
}
