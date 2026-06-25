// ARM64 processor support
// https://github.com/usbarmory/tamago
//
// Copyright (c) The TamaGo Authors. All Rights Reserved.
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package arm64

import (
	"runtime"

	"github.com/usbarmory/tamago/internal/reg"
)

const (
	l1pageTableOffset = 0x4000
	l1pageTableSize   = 512

	l2pageTableOffset = 0x5000
	l2pageTableSize   = 512

	// another L2 table is appended at 0x6000

	l3pageTableOffset = 0x7000
	l3pageTableSize   = 512
)

// Memory region attributes
//
// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// G5.7.
const (
	TTE_XN   = 53
	TTE_AF   = 10
	TTE_SH   = 8
	TTE_AP   = 6
	TTE_ATTR = 2
	TTE_DESC = 0

	TTE_BLOCK         uint64 = (0b01 << TTE_DESC)
	TTE_TABLE         uint64 = (0b11 << TTE_DESC)
	TTE_PAGE          uint64 = (0b11 << TTE_DESC)
	TTE_NON_SH        uint64 = (0b00 << TTE_SH)
	TTE_OUTER_SH      uint64 = (0b10 << TTE_SH)
	TTE_INNER_SH      uint64 = (0b11 << TTE_SH)
	TTE_EXECUTE_NEVER uint64 = (0b11 << TTE_XN)

	// Device-nGnRnE
	DeviceRegion uint64 = 0b00000000
	// Normal, Inner/Outer WB/WA/RA
	MemoryRegion uint64 = 0b11111111
	// Normal, Inner/Outer Non-cacheable
	NormalNonCacheableRegion uint64 = 0b01000100

	deviceAttributeIndex   = 0
	memoryAttributeIndex   = 1
	normalNCAttributeIndex = 2

	deviceAttributes   = 1<<TTE_AF | TTE_OUTER_SH | TTE_AP_00<<TTE_AP | deviceAttributeIndex<<TTE_ATTR
	memoryAttributes   = 1<<TTE_AF | TTE_INNER_SH | TTE_AP_00<<TTE_AP | memoryAttributeIndex<<TTE_ATTR
	normalNCAttributes = 1<<TTE_AF | TTE_INNER_SH | TTE_AP_00<<TTE_AP | normalNCAttributeIndex<<TTE_ATTR

	NormalNonCacheableAttributes = normalNCAttributes | TTE_BLOCK
)

// MMU access permissions
//
// ARM Architecture Reference Manual ARMv8,for ARMv8-A architecture profile
// Table D5-25, Data access permissions for stage 1 translations.
const (
	// EL1 or above: read/write, EL0: none
	TTE_AP_00 = 0b00
	// EL1 or above: read/write, EL0: read/write
	TTE_AP_01 = 0b01
	// EL1 or above: read-only, EL0: none
	TTE_AP_10 = 0b10
	// EL1 or above: read-only, EL0: read-only
	TTE_AP_11 = 0b11
)

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D12.2.103 TCR_EL1, Translation Control Register (EL1).
const (
	TCR_IPS   = 32
	TCR_TBID  = 29
	TCR_TG0   = 14
	TCR_SH0   = 12
	TCR_ORGN0 = 10
	TCR_IRGN0 = 8
	TCR_T0SZ  = 0

	// 32-bit or 40-bit intermediate physical address size
	tcr uint64 = 0b010<<TCR_IPS |
		// 4KB granule
		0b00<<TCR_TG0 |
		// inner shareable
		0b11<<TCR_SH0 |
		// outer cacheability (normal, cacheable)
		0b01<<TCR_ORGN0 |
		// inner cacheability (normal, cacheable)
		0b01<<TCR_IRGN0 |
		// memory region size offset 0:5 (39-bits VA space)
		25<<TCR_T0SZ
)

// defined in mmu.s
func flush_tlb()
func write_mair_el1(val uint64)
func write_tcr_el1(val uint64)
func set_ttbr0_el1(addr uint64)

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D5.3.1 Translation table level 0, level 1, and level 2 descriptor formats.
func (cpu *CPU) initL1Table(entry int, ttbr uint64, section uint64) {
	n := 30 // 1GB

	ramStart, ramEnd := runtime.MemRegion()
	_, textEnd := runtime.TextRegion()

	memoryRegion := memoryAttributes | TTE_BLOCK
	deviceRegion := deviceAttributes | TTE_BLOCK

	for i := uint64(entry); i < l1pageTableSize; i++ {
		page := ttbr + 8*i
		addr := section + (i << n)

		switch {
		case addr < textEnd && (addr+(1<<n)) > textEnd:
			// skip first L2 table, pointing to L3
			l2pageTableStart := ramStart + l2pageTableOffset
			base := l2pageTableStart + l2pageTableSize*8

			// use L2 table to end non-executable boundary
			// precisely at textStart
			reg.Write64(page, base|TTE_TABLE)
			cpu.initL2Table(0, base, addr)
		case addr >= ramStart && addr < textEnd:
			reg.Write64(page, addr|memoryRegion)
		case addr >= ramStart && addr < ramEnd:
			reg.Write64(page, addr|memoryRegion|TTE_EXECUTE_NEVER)
		default:
			reg.Write64(page, addr|deviceRegion|TTE_EXECUTE_NEVER)
		}
	}
}

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D5.3.1 Translation table level 0, level 1, and level 2 descriptor formats.
func (cpu *CPU) initL2Table(entry int, base uint64, section uint64) {
	n := 21 // 2MB

	ramStart, ramEnd := runtime.MemRegion()
	_, textEnd := runtime.TextRegion()

	memoryRegion := memoryAttributes | TTE_BLOCK
	deviceRegion := deviceAttributes | TTE_BLOCK

	for i := uint64(entry); i < l2pageTableSize; i++ {
		page := base + 8*i
		addr := section + (i << n)

		switch {
		case addr < textEnd && (addr+(1<<n)) > textEnd:
			// skip first L3 table, reserved to trap null pointers
			l3pageTableStart := ramStart + l3pageTableOffset
			base := l3pageTableStart + l3pageTableSize*8

			// use L3 table to end non-executable boundary
			// precisely at textStart
			reg.Write64(page, base|TTE_TABLE)
			cpu.initL3Table(0, base, addr)
		case addr >= ramStart && addr < textEnd:
			reg.Write64(page, addr|memoryRegion)
		case addr >= ramStart && addr < ramEnd:
			reg.Write64(page, addr|memoryRegion|TTE_EXECUTE_NEVER)
		default:
			reg.Write64(page, addr|deviceRegion|TTE_EXECUTE_NEVER)
		}
	}
}

// ARM Architecture Reference Manual ARMv8, for ARMv8-A architecture profile
// D5.3.2 ARMv8 translation table level 3 descriptor formats.
func (cpu *CPU) initL3Table(entry int, base uint64, section uint64) {
	n := 12 // 4KB

	ramStart, ramEnd := runtime.MemRegion()
	_, textEnd := runtime.TextRegion()

	memoryRegion := memoryAttributes | TTE_PAGE
	deviceRegion := deviceAttributes | TTE_PAGE

	for i := uint64(entry); i < l3pageTableSize; i++ {
		page := base + 8*i
		addr := section + (i << n)

		switch {
		case addr >= ramStart && addr < textEnd:
			reg.Write64(page, addr|memoryRegion)
		case addr >= ramStart && addr < ramEnd:
			reg.Write64(page, addr|memoryRegion|TTE_EXECUTE_NEVER)
		default:
			reg.Write64(page, addr|deviceRegion|TTE_EXECUTE_NEVER)
		}
	}
}

// InitMMU initializes the first-level translation tables for all available
// memory with a flat mapping and privileged attribute flags.
//
// The first 4096 bytes (0x00000000 - 0x00001000) are flagged as invalid to
// trap null pointers.
//
// All available memory is marked as non-executable except for the range
// returned by runtime.TextRegion().
func (cpu *CPU) InitMMU() {
	ramStart, _ := runtime.MemRegion()

	l1pageTableStart := ramStart + l1pageTableOffset
	l2pageTableStart := ramStart + l2pageTableOffset
	l3pageTableStart := ramStart + l3pageTableOffset

	// Map the first L1 entry to an L2 table.
	tte := l2pageTableStart | TTE_TABLE
	reg.Write64(l1pageTableStart, tte)

	// Map the first L2 entry to an L3 table to trap null pointers within
	// the smallest possible section (4KB starting from 0x00000000).
	tte = l3pageTableStart | TTE_TABLE
	reg.Write64(l2pageTableStart, tte)

	// set first L3 entry as invalid
	reg.Write64(l3pageTableStart, 0)

	// set remaining entries with flat mapping
	cpu.initL1Table(1, l1pageTableStart, 0)
	cpu.initL2Table(1, l2pageTableStart, 0)
	cpu.initL3Table(1, l3pageTableStart, 0)

	// set memory region attributes
	//   * attr0: device
	//   * attr1: memory
	write_mair_el1(
		MemoryRegion<<(8*memoryAttributeIndex) |
			NormalNonCacheableRegion<<(8*normalNCAttributeIndex) |
			DeviceRegion<<(8*deviceAttributeIndex))

	// set translation control register
	write_tcr_el1(tcr)

	// enable MMU
	set_ttbr0_el1(l1pageTableStart)
}

// ConfigureMMU (re)configures the translation tables for the provided memory
// range with the argument attribute flags. An alias argument greater than zero
// specifies the physical address corresponding to the start argument in case
// virtual memory is required, otherwise a flat 1:1 mapping is set.
//
// Ranges below 1GB are configured with 2MB block mappings from the first L2
// table. Ranges at or above 1GB are configured with 1GB L1 block mappings.
func (cpu *CPU) ConfigureMMU(start, end, alias, flags uint64) {
	ramStart, _ := runtime.MemRegion()
	l1pageTableStart := ramStart + l1pageTableOffset

	if end <= start {
		return
	}

	cpu.configureL1(l1pageTableStart, start, end, alias, flags)
	cpu.FlushTLBs()
}

func (cpu *CPU) configureL1(table, start, end, alias, flags uint64) {
	startEntry := start >> 30
	endEntry := (end + (1 << 30) - 1) >> 30

	for i := startEntry; i < l1pageTableSize && i < endEntry; i++ {
		sectionStart := i << 30
		sectionEnd := sectionStart + (1 << 30)
		rangeStart := max(start, sectionStart)
		rangeEnd := min(end, sectionEnd)
		page := table + 8*i
		tte := reg.Read64(page)

		if tte&(0b11<<TTE_DESC) == TTE_TABLE {
			table := tte &^ 0xfff
			tableAlias := uint64(0)

			if alias > 0 {
				tableAlias = alias + rangeStart - start
			}

			cpu.configureL2(table, rangeStart, rangeEnd, tableAlias, flags)
			continue
		}

		pa := sectionStart
		if alias > 0 {
			pa = alias + sectionStart - start
		}

		reg.Write64(page, pa|flags)
	}
}

func (cpu *CPU) configureL2(table, start, end, alias, flags uint64) {
	section := start &^ ((1 << 30) - 1)
	startEntry := (start - section) >> 21
	endEntry := (end + (1 << 21) - 1 - section) >> 21

	for i := startEntry; i < l2pageTableSize && i < endEntry; i++ {
		pa := section + (i << 21)
		if alias > 0 {
			pa = alias + (i-startEntry)<<21
		}

		reg.Write64(table+8*i, pa|flags)
	}
}

func (cpu *CPU) updateMMU(start, end uint64, mask uint64, val uint64) {
	ramStart, _ := runtime.MemRegion()
	l1pageTableStart := ramStart + l1pageTableOffset
	l2pageTableStart := ramStart + l2pageTableOffset

	update := func(page uint64) {
		tte := reg.Read64(page)
		reg.Write64(page, (tte&^mask)|val)
	}

	if start < 1<<30 {
		l2End := end
		if l2End > 1<<30 {
			l2End = 1 << 30
		}

		for i, e := start>>21, (l2End+(1<<21)-1)>>21; i < l2pageTableSize && i < e; i++ {
			update(l2pageTableStart + 8*i)
		}

		start = 1 << 30
	}

	for i, e := start>>30, (end+(1<<30)-1)>>30; i < l1pageTableSize && i < e; i++ {
		update(l1pageTableStart + 8*i)
	}

	cpu.FlushTLBs()
}

// SetAttributes (re)configures translation table attributes for the provided
// memory range.
func (cpu *CPU) SetAttributes(start, end, flags uint64) {
	mask := TTE_EXECUTE_NEVER | (0b11 << TTE_SH) | (0b11 << TTE_AP) |
		(0b111 << TTE_ATTR) | TTE_DESC

	cpu.updateMMU(start, end, mask, flags)
}
