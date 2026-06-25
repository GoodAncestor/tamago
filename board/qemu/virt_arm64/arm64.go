// QEMU virt support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package virt_arm64 provides hardware initialization, automatically on import,
// for a QEMU virt ARM64 virtual machine configured with one core.
package virt_arm64

import (
	"runtime/goos"
	_ "unsafe"

	tamagoarm64 "github.com/usbarmory/tamago/arm64"
	"github.com/usbarmory/tamago/arm64/gic"
	"github.com/usbarmory/tamago/dma"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/kvm/virtio"
	"github.com/usbarmory/tamago/soc/arm/pl011"
)

const (
	dmaStart = 0x58000000
	dmaSize  = 0x08000000 // 128MB
)

// Peripheral registers.
const (
	RAM_START = 0x40000000

	GICD_BASE = 0x08000000
	GICR_BASE = 0x080a0000
	GITS_BASE = 0x08080000

	UART0_BASE = 0x09000000

	VIRTIO_MMIO_BASE = 0x0a000000
	VIRTIO_MMIO_SIZE = 0x00000200
	VIRTIO_MMIO_NUM  = 32
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

	// First QEMU virtio-mmio transport slot.
	VirtioMMIO = &virtio.MMIO{
		Base: VIRTIO_MMIO_BASE,
	}
)

//go:linkname nanotime runtime/goos.Nanotime
func nanotime() int64 {
	return ARM64.GetTime()
}

// VirtioMMIODevice returns the first QEMU virtio-mmio transport exposing the
// requested VirtIO subsystem device ID.
func VirtioMMIODevice(deviceID uint32) (*virtio.MMIO, bool) {
	for i := uint32(0); i < VIRTIO_MMIO_NUM; i++ {
		base := VIRTIO_MMIO_BASE + i*VIRTIO_MMIO_SIZE

		if reg.Read(base+virtio.Magic) != virtio.MAGIC {
			continue
		}

		if reg.Read(base+virtio.DeviceID) == deviceID {
			return &virtio.MMIO{Base: base}, true
		}
	}

	return nil, false
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

	goos.Exit = func(_ int32) {
		for {
			ARM64.WaitInterrupt()
		}
	}
}

func init() {
	dma.Init(dmaStart, dmaSize)
}
