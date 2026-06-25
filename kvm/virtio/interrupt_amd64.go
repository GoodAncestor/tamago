// VirtIO over PCI driver (non-transitional)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build amd64

package virtio

import "github.com/usbarmory/tamago/amd64"

func interruptVector(id int) (addr uint64, data uint32, err error) {
	return uint64(amd64.LAPIC_BASE), uint32(id), nil
}
