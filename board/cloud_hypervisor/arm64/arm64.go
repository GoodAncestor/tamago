// Cloud Hypervisor support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package arm64 provides hardware initialization, automatically on import, for
// a Cloud Hypervisor ARM64 virtual machine configured with one core.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package arm64

import (
	"runtime/goos"
	_ "unsafe"

	tamagoarm64 "github.com/usbarmory/tamago/arm64"
	"github.com/usbarmory/tamago/arm64/gic"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/soc/arm/pl011"
	"github.com/usbarmory/tamago/soc/intel/pci"
)

const (
	dmaStart = 0x58000000
	dmaSize  = 0x08000000 // 128MB
)

// Peripheral registers.
const (
	RAM_START = 0x40000000

	GICD_BASE = 0x08ff0000
	GICR_BASE = 0x08fd0000
	GITS_BASE = 0x08fb0000

	UART0_BASE = 0x09000000

	PCIE_ECAM_BASE = 0x30000000

	// VirtIO Networking
	VIRTIO_NET_PCI_VENDOR = 0x1af4 // Red Hat, Inc.
	VIRTIO_NET_PCI_DEVICE = 0x1041 // Virtio 1.0 network device

	VIRTIO_NET_BAR1 = 0x10058000
	VIRTIO_NET_BAR4 = 0x10040000
)

// Peripheral instances.
var (
	// CPU instance.
	ARM64 = &tamagoarm64.CPU{}

	// Generic Interrupt Controller.
	GIC = &gic.GIC{
		GICD: GICD_BASE,
		GICR: GICR_BASE,
	}

	// Serial port.
	UART0 = &pl011.UART{
		Base: UART0_BASE,
	}
)

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return ARM64.GetTime()
}

// Init takes care of lower-level initialization triggered early in runtime
// setup (post World start).
//
//go:linkname Init runtime/goos.Hwinit1
func Init() {
	ARM64.Init()
	ARM64.InitGenericTimers(0, 0)

	GIC.Init()
	GIC.EnableInterrupt(tamagoarm64.TIMER_IRQ)

	UART0.Init()
	ARM64.ConfigureMMU(dmaStart, dmaStart+dmaSize, 0, tamagoarm64.NormalNonCacheableAttributes|tamagoarm64.TTE_EXECUTE_NEVER)
	pci.ConfigureECAM(PCIE_ECAM_BASE)

	if dev := pci.Probe(0, VIRTIO_NET_PCI_VENDOR, VIRTIO_NET_PCI_DEVICE); dev != nil {
		// set I/O Space, Memory Space and Bus Master Enable
		dev.Write(0, pci.Command, 0x7)
		// assign BARs inside the generic PCI host bridge memory window
		dev.Write(0, pci.Bar1, VIRTIO_NET_BAR1)
		dev.Write(0, pci.Bar4, VIRTIO_NET_BAR4)
		dev.Write(0, pci.Bar4+4, 0)
	}

	goos.Exit = func(_ int32) {
		for {
			ARM64.WaitInterrupt()
		}
	}
}

func init() {
	dma.Init(dmaStart, dmaSize)
}
