// Intel Peripheral Component Interconnect (PCI) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package pci implements a driver for Intel Peripheral Component Interconnect
// (PCI) controllers adopting the following reference
// specifications:
//   - PCI Local Bus Specification, revision 3.0, PCI Special Interest Group
//
// This package is only meant to be used with `GOOS=tamago` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package pci

import (
	"sync/atomic"
	"unsafe"

	"github.com/usbarmory/tamago/bits"
)

const (
	CONFIG_ADDRESS = 0x0cf8
	CONFIG_DATA    = 0x0cfc
)

const (
	maxBuses   = 256
	maxDevices = 32
)

var ecamBase uint64

// Header Type 0x0 offsets
const (
	VendorID           = 0x00
	Command            = 0x04
	RevisionID         = 0x08
	Bar0               = 0x10
	Bar1               = 0x14
	Bar2               = 0x18
	Bar3               = 0x1c
	Bar4               = 0x20
	Bar5               = 0x24
	CapabilitiesOffset = 0x34
)

// Device represents a PCI device.
type Device struct {
	// Bus number
	Bus uint32
	// Vendor ID
	Vendor uint16
	// Device ID
	Device uint16

	// PCI Slot
	Slot uint32
}

func (d *Device) address(fn uint32, off uint32) uint32 {
	return 1<<31 | d.Bus<<16 | d.Slot<<11 | fn<<8 | off&0xfc
}

func (d *Device) ecamAddress(fn uint32, off uint32) uint64 {
	return ecamBase + uint64(d.Bus)<<20 + uint64(d.Slot)<<15 + uint64(fn)<<12 + uint64(off&0xfc)
}

// ConfigureECAM configures PCI Configuration Space access through an Enhanced
// Configuration Access Mechanism base address.
func ConfigureECAM(base uint64) {
	ecamBase = base
}

func read32(addr uint64) uint32 {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	return atomic.LoadUint32(reg)
}

func write32(addr uint64, val uint32) {
	reg := (*uint32)(unsafe.Pointer(uintptr(addr)))
	atomic.StoreUint32(reg, val)
}

// Read reads the device configuration space for a given function and
// register offset.
func (d *Device) Read(fn uint32, off uint32) uint32 {
	if ecamBase != 0 {
		return read32(d.ecamAddress(fn, off)) >> ((off & 2) * 8)
	}

	configPortWrite(CONFIG_ADDRESS, d.address(fn, off))
	return configPortRead(CONFIG_DATA) >> ((off & 2) * 8)
}

// Write writes the device configuration space for a given function and
// register offset, the offset must be 32-bit aligned.
func (d *Device) Write(fn uint32, off uint32, val uint32) {
	if (off&2)*8 != 0 {
		return
	}

	if ecamBase != 0 {
		write32(d.ecamAddress(fn, off), val)
		return
	}

	configPortWrite(CONFIG_ADDRESS, d.address(fn, off))
	configPortWrite(CONFIG_DATA, val)
}

// BaseAddress returns a device Base Address register (BAR).
func (d *Device) BaseAddress(n int) uint {
	if n > 5 {
		return 0
	}

	off := Bar0 + uint32(n)*4
	bar := d.Read(0, Bar0+uint32(n)*4)

	// decode BAR Type
	switch bits.GetN(&bar, 1, 0b11) {
	case 0b00:
		return uint(bar) & 0xfffffff0
	case 0b10:
		return uint(d.Read(0, off+4))<<32 | uint(bar)&0xfffffff0
	}

	return 0
}

func (d *Device) probe() bool {
	if d.Bus > maxBuses {
		return false
	}

	val := d.Read(0, VendorID)

	if d.Vendor = uint16(val); d.Vendor == 0xffff {
		return false
	}

	d.Device = uint16(val >> 16)

	return true
}

// Probe probes a PCI device.
func Probe(bus int, vendor uint16, device uint16) *Device {
	d := &Device{
		Bus: uint32(bus),
	}

	for slot := range uint32(maxDevices) {
		d.Slot = slot

		if d.probe() && d.Vendor == vendor && d.Device == device {
			return d
		}
	}

	return nil
}

// Devices returns all found PCI devices on a given bus.
func Devices(bus int) (devices []*Device) {
	for slot := range uint32(maxDevices) {
		d := &Device{
			Bus:  uint32(bus),
			Slot: slot,
		}

		if d.probe() {
			devices = append(devices, d)
		}
	}

	return
}
