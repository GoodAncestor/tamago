// Intel Peripheral Component Interconnect (PCI) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build amd64

package pci

import "github.com/usbarmory/tamago/internal/reg"

func configPortRead(port uint16) uint32 {
	return reg.In32(port)
}

func configPortWrite(port uint16, val uint32) {
	reg.Out32(port, val)
}
