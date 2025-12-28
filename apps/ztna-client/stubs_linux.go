//go:build !windows

package main

import "fmt"

type NativeWintun struct{}

func CreateNativeWintun(name, tunnelType string) (*NativeWintun, error) {
	return nil, fmt.Errorf("not supported on this OS")
}

func OpenNativeWintun(name string) (*NativeWintun, error) {
	return nil, fmt.Errorf("not supported on this OS")
}

func DeleteWintunAdapter(name string) error {
	return nil
}

func (w *NativeWintun) Close()                        {}
func (w *NativeWintun) GetLUID() (uint64, error)      { return 0, fmt.Errorf("not supported") }
func (w *NativeWintun) ReadPacket() ([]byte, error)   { return nil, fmt.Errorf("not supported") }
func (w *NativeWintun) WritePacket(data []byte) error { return fmt.Errorf("not supported") }

type LUID uint64

func SetAdapterIP(luid LUID, ipStr string) error {
	return fmt.Errorf("not supported")
}

func AddWinsRoute(luid LUID, destCIDR string) error {
	return fmt.Errorf("not supported")
}

func RemoveWinsRoute(luid LUID, destCIDR string) error {
	return fmt.Errorf("not supported")
}
