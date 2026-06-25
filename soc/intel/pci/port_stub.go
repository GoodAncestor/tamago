// Intel Peripheral Component Interconnect (PCI) driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !amd64

package pci

func configPortRead(_ uint16) uint32 {
	return 0xffffffff
}

func configPortWrite(_ uint16, _ uint32) {}
