// Package nes implements the main NES emulator, coordinating CPU, PPU, and cartridge.
package nes

import (
	"fmt"

	"github.com/andrewthecodertx/go-65c02-emulator/pkg/mos6502"
	"github.com/andrewthecodertx/nes-emulator/pkg/bus"
	"github.com/andrewthecodertx/nes-emulator/pkg/cartridge"
	"github.com/andrewthecodertx/nes-emulator/pkg/ppu"
)

// NES represents the complete NES emulator system
type NES struct {
	cpu       *mos6502.CPU           // 6502 CPU
	bus       *bus.NESBus            // System bus
	ppu       *ppu.PPU               // Picture Processing Unit
	cartridge *cartridge.Cartridge   // Loaded cartridge
	cycles    uint64                 // Total CPU cycles executed
}

// New creates a new NES emulator from a ROM file
func New(romPath string) (*NES, error) {
	// Load cartridge from ROM file
	cart, err := cartridge.LoadFromFile(romPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load ROM: %w", err)
	}

	return NewFromCartridge(cart), nil
}

// NewFromCartridge creates a new NES emulator from a cartridge
func NewFromCartridge(cart *cartridge.Cartridge) *NES {
	// Create PPU
	ppuUnit := ppu.NewPPU()
	ppuUnit.SetMapper(cart.GetMapper())
	ppuUnit.SetMirroring(cart.GetMirroring())

	// Create system bus
	nesbus := bus.NewNESBus(ppuUnit, cart.GetMapper())

	// Create CPU with the bus
	cpu := mos6502.NewCPU(nesbus)

	nes := &NES{
		cpu:       cpu,
		bus:       nesbus,
		ppu:       ppuUnit,
		cartridge: cart,
		cycles:    0,
	}

	return nes
}

// Reset resets the NES to power-on state
func (n *NES) Reset() {
	n.cpu.Reset()
	n.ppu.Reset()
	n.cycles = 0
}

// Step executes one CPU cycle
// Returns 1 (always consumes 1 CPU cycle)
func (n *NES) Step() uint8 {
	// Execute one CPU cycle
	// The CPU's Step() method handles multi-cycle instructions internally
	n.cpu.Step()

	// Clock the bus once (which clocks PPU at 3x)
	n.bus.Clock()

	// Check for NMI from PPU
	if n.bus.IsNMI() {
		n.cpu.NMIPending = true
	}

	n.cycles++
	return 1
}

// RunFrame runs the emulator until a complete frame is rendered
// Returns when the PPU has finished rendering one frame (~29780 CPU cycles)
func (n *NES) RunFrame() {
	// Run until the PPU completes a frame
	// The PPU sets frameComplete=true at the end of scanline 261

	// First, clear the frame complete flag
	n.ppu.ClearFrameComplete()

	// Run until a frame is complete
	for !n.ppu.IsFrameComplete() {
		n.Step()
	}
}

// Clock executes one CPU cycle
func (n *NES) Clock() {
	n.Step()
}

// GetFrameBuffer returns the current PPU frame buffer
// The buffer contains 256x240 pixels, each byte is a palette index (0-63)
func (n *NES) GetFrameBuffer() *[ppu.ScreenWidth * ppu.ScreenHeight]uint8 {
	return n.ppu.GetFrameBuffer()
}

// GetPPU returns a pointer to the PPU for direct access
func (n *NES) GetPPU() *ppu.PPU {
	return n.ppu
}

// GetCPU returns a pointer to the CPU for direct access
func (n *NES) GetCPU() *mos6502.CPU {
	return n.cpu
}

// GetBus returns a pointer to the system bus for direct access
func (n *NES) GetBus() *bus.NESBus {
	return n.bus
}

// GetCycles returns the total number of CPU cycles executed
func (n *NES) GetCycles() uint64 {
	return n.cycles
}

// GetCartridge returns a pointer to the loaded cartridge
func (n *NES) GetCartridge() *cartridge.Cartridge {
	return n.cartridge
}
