// Package bus implements the NES system bus connecting CPU, RAM, PPU, and cartridge.
package bus

import (
	"github.com/andrewthecodertx/go-6502-emulator/pkg/core"
	"github.com/andrewthecodertx/go-nes-emulator/pkg/cartridge"
	"github.com/andrewthecodertx/go-nes-emulator/pkg/controller"
	"github.com/andrewthecodertx/go-nes-emulator/pkg/ppu"
)

// NESBus implements the core.Bus interface for the NES system
//
// CPU Memory Map:
//
//	$0000-$07FF: 2KB internal RAM
//	$0800-$1FFF: Mirrors of $0000-$07FF
//	$2000-$2007: PPU registers
//	$2008-$3FFF: Mirrors of $2000-$2007
//	$4000-$4017: APU and I/O registers
//	$4018-$401F: APU and I/O functionality (rarely used)
//	$4020-$FFFF: Cartridge space (PRG-ROM, PRG-RAM, mapper registers)
type NESBus struct {
	// 2KB CPU RAM (mirrored to fill $0000-$1FFF)
	cpuRAM [2048]uint8

	// PPU (Picture Processing Unit)
	ppu *ppu.PPU

	// Cartridge mapper
	mapper cartridge.Mapper

	// Controllers
	controller1 *controller.Controller
	controller2 *controller.Controller

	// DMA transfer state
	dmaPage     uint8
	dmaTransfer bool
}

// Ensure NESBus implements core.Bus
var _ core.Bus = (*NESBus)(nil)

// NewNESBus creates a new NES system bus
func NewNESBus(ppuUnit *ppu.PPU, mapper cartridge.Mapper) *NESBus {
	return &NESBus{
		ppu:         ppuUnit,
		mapper:      mapper,
		controller1: controller.NewController(),
		controller2: controller.NewController(),
	}
}

// Read implements core.Bus.Read for the CPU
func (b *NESBus) Read(addr uint16) uint8 {
	switch {
	case addr < 0x2000:
		// CPU RAM (with mirroring)
		return b.cpuRAM[addr&0x07FF]

	case addr < 0x4000:
		// PPU registers (mirrored every 8 bytes)
		return b.ppu.ReadCPURegister(0x2000 + (addr & 0x0007))

	case addr == 0x4015:
		// APU Status register (stub - APU not implemented)
		// Return 0 to indicate no sound channels active
		return 0

	case addr == 0x4016:
		// Controller 1
		return b.controller1.Read()

	case addr == 0x4017:
		// Controller 2
		return b.controller2.Read()

	case addr >= 0x4000 && addr < 0x4020:
		// Other APU/IO registers - return 0 (open bus)
		return 0

	case addr >= 0x4020:
		// Cartridge space
		return b.mapper.ReadPRG(addr)
	}

	return 0
}

// Write implements core.Bus.Write for the CPU
func (b *NESBus) Write(addr uint16, data uint8) {
	switch {
	case addr < 0x2000:
		// CPU RAM (with mirroring)
		b.cpuRAM[addr&0x07FF] = data

	case addr < 0x4000:
		// PPU registers (mirrored every 8 bytes)
		b.ppu.WriteCPURegister(0x2000+(addr&0x0007), data)

	case addr == 0x4014:
		// OAMDMA: DMA transfer of 256 bytes from CPU memory to OAM
		b.dmaPage = data
		b.dmaTransfer = true

	case addr == 0x4016:
		// Controller strobe
		// Writing 1 then 0 latches controller button states
		b.controller1.Write(data)
		b.controller2.Write(data)

	case addr >= 0x4020:
		// Cartridge space
		b.mapper.WritePRG(addr, data)
	}
}

// Clock advances the bus by one CPU cycle
// This runs the PPU at 3x CPU speed and handles DMA transfers
func (b *NESBus) Clock() {
	// PPU runs at 3x CPU speed
	b.ppu.Clock()
	b.ppu.Clock()
	b.ppu.Clock()

	// Handle DMA transfer if active
	if b.dmaTransfer {
		// DMA transfers 256 bytes from CPU memory page to OAM
		// Simplified implementation: copy all bytes immediately
		// (Real hardware takes 513-514 cycles, but this works for most games)
		for i := 0; i < 256; i++ {
			addr := uint16(b.dmaPage)<<8 | uint16(i)
			b.ppu.WriteOAM(uint8(i), b.Read(addr))
		}
		b.dmaTransfer = false
	}
}

// IsNMI returns true if the PPU is requesting an NMI
func (b *NESBus) IsNMI() bool {
	return b.ppu.GetNMI()
}

// GetPPU returns a pointer to the PPU
func (b *NESBus) GetPPU() *ppu.PPU {
	return b.ppu
}

// GetController returns a pointer to the specified controller (0 or 1)
func (b *NESBus) GetController(num int) *controller.Controller {
	if num == 0 {
		return b.controller1
	}
	return b.controller2
}
