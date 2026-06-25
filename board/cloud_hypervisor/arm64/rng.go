// Cloud Hypervisor support for tamago/arm64
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	_ "unsafe"

	"github.com/usbarmory/tamago/internal/rng"
)

var rngState uint64 = 0x9e3779b97f4a7c15

//go:linkname initRNG runtime/goos.InitRNG
func initRNG() {
	rng.GetRandomDataFn = getRandomData
}

func getRandomData(b []byte) {
	for i := range b {
		rngState ^= rngState << 7
		rngState ^= rngState >> 9
		rngState ^= rngState << 8

		b[i] = byte(rngState)
	}
}
