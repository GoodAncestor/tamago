// QEMU virt support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package virt_arm64

import (
	_ "unsafe"
)

//go:linkname ramStart runtime/goos.RamStart
var ramStart uint64 = RAM_START

//go:linkname ramSize runtime/goos.RamSize
var ramSize uint64 = 0x18000000 // 384MB, leaving top 128MB for DMA
