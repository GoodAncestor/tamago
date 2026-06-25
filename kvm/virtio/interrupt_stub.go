// VirtIO over PCI driver (non-transitional)
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

//go:build !amd64

package virtio

import "errors"

func interruptVector(_ int) (addr uint64, data uint32, err error) {
	return 0, 0, errors.New("MSI-X interrupt routing is not implemented for this architecture")
}
