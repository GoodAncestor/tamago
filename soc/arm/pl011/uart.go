// ARM PrimeCell PL011 UART driver
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package pl011 implements a minimal ARM PrimeCell PL011 UART driver.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm64` as
// supported by the TamaGo framework for bare metal Go, see
// https://github.com/usbarmory/tamago.
package pl011

import (
	"runtime"

	"github.com/usbarmory/tamago/internal/reg"
)

const (
	DR   = 0x000
	FR   = 0x018
	IBRD = 0x024
	FBRD = 0x028
	LCRH = 0x02c
	CR   = 0x030
	IMSC = 0x038
	ICR  = 0x044

	FR_RXFE = 4
	FR_TXFF = 5

	LCRH_FEN  = 4
	LCRH_WLEN = 5

	CR_UARTEN = 0
	CR_TXE    = 8
	CR_RXE    = 9
)

// UART represents a PL011 serial port instance.
type UART struct {
	// Base register
	Base uint32
}

// Init initializes the UART for basic polling I/O.
func (hw *UART) Init() {
	if hw.Base == 0 {
		panic("invalid UART controller instance")
	}

	reg.Write(hw.Base+CR, 0)
	reg.Write(hw.Base+ICR, 0x7ff)
	reg.Write(hw.Base+IMSC, 0)

	// Leave baud-rate divisors untouched. Hypervisors generally preconfigure
	// the emulated PL011, and the console path only needs polled FIFO access.
	reg.Write(hw.Base+LCRH, 0b11<<LCRH_WLEN|1<<LCRH_FEN)
	reg.Write(hw.Base+CR, 1<<CR_UARTEN|1<<CR_TXE|1<<CR_RXE)
}

// Tx transmits a single character to the serial port.
func (hw *UART) Tx(c byte) {
	for reg.Get(hw.Base+FR, FR_TXFF) {
		// wait for TX FIFO to have room for a character
	}

	reg.Write(hw.Base+DR, uint32(c))
}

// Rx receives a single character from the serial port.
func (hw *UART) Rx() (c byte, valid bool) {
	if reg.Get(hw.Base+FR, FR_RXFE) {
		return
	}

	return byte(reg.Read(hw.Base + DR)), true
}

// Write data from buffer to serial port.
func (hw *UART) Write(buf []byte) (n int, _ error) {
	for n = range buf {
		hw.Tx(buf[n])
	}

	return
}

// Read available data to buffer from serial port.
func (hw *UART) Read(buf []byte) (n int, _ error) {
	var valid bool

	for n = range buf {
		buf[n], valid = hw.Rx()

		if !valid {
			if n == 0 {
				runtime.Gosched()
			}

			break
		}
	}

	return
}
