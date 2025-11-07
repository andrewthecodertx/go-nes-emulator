// Package cartridge implements NES cartridge ROM loading and memory mappers.
//
// NES cartridges contain PRG-ROM (program code for CPU) and CHR-ROM/RAM
// (graphics data for PPU). Different cartridges use different mapper chips
// to extend the NES's memory space through bank switching.
package cartridge

// Mapper defines the interface for NES cartridge mappers
//
// Mappers handle the translation between CPU/PPU addresses and actual
// ROM/RAM locations. Different mapper numbers implement different
// bank switching schemes.
type Mapper interface {
	// ReadPRG reads a byte from PRG-ROM/RAM (CPU address space $8000-$FFFF)
	ReadPRG(addr uint16) uint8

	// WritePRG writes a byte to PRG-RAM or triggers mapper control (CPU address space $6000-$FFFF)
	WritePRG(addr uint16, value uint8)

	// ReadCHR reads a byte from CHR-ROM/RAM (PPU address space $0000-$1FFF)
	ReadCHR(addr uint16) uint8

	// WriteCHR writes a byte to CHR-RAM (PPU address space $0000-$1FFF)
	// CHR-ROM is read-only; writes may be ignored or used for mapper control
	WriteCHR(addr uint16, value uint8)

	// Scanline is called by the PPU on each scanline (for IRQ timing)
	Scanline()

	// GetMirroring returns the current nametable mirroring mode
	GetMirroring() uint8
}
