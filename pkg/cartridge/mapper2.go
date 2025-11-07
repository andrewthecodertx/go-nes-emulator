package cartridge

// Mapper2 implements iNES Mapper 2 (UxROM)
//
// UxROM is used by games like Mega Man, Castlevania, Duck Tales.
// It provides switchable PRG-ROM banks with fixed CHR-RAM.
//
// PRG-ROM: Up to 256KB (16 banks of 16KB)
// CHR-RAM: 8KB (not switchable)
//
// CPU Memory Map:
//   $6000-$7FFF: Family Basic only (not implemented)
//   $8000-$BFFF: Switchable 16 KB PRG-ROM bank
//   $C000-$FFFF: Fixed 16 KB PRG-ROM bank (last bank)
//
// PPU Memory Map:
//   $0000-$1FFF: 8 KB CHR-RAM (not switchable)
//
// Bank Switching:
//   Writing to $8000-$FFFF selects which 16KB PRG bank appears at $8000-$BFFF
//   Only the lower 3-4 bits are used (depending on ROM size)
type Mapper2 struct {
	prgROM []uint8 // Full PRG-ROM (all banks)
	chrRAM []uint8 // 8KB CHR-RAM

	prgBanks  uint8 // Number of 16KB PRG banks
	prgBank   uint8 // Currently selected PRG bank at $8000-$BFFF
	mirroring uint8 // Nametable mirroring mode
}

// NewMapper2 creates a new UxROM mapper (Mapper 2)
func NewMapper2(prgROM, chrROM []uint8, mirroring uint8) *Mapper2 {
	m := &Mapper2{
		prgROM:    make([]uint8, len(prgROM)),
		chrRAM:    make([]uint8, 8192), // Always 8KB CHR-RAM
		prgBanks:  uint8(len(prgROM) / 16384),
		prgBank:   0, // Start with first bank
		mirroring: mirroring,
	}

	copy(m.prgROM, prgROM)

	// UxROM uses CHR-RAM, ignore any CHR-ROM data
	// (Some ROMs may have CHR-ROM data but it's not used)

	return m
}

// ReadPRG reads from PRG-ROM (CPU $8000-$FFFF)
func (m *Mapper2) ReadPRG(addr uint16) uint8 {
	switch {
	case addr >= 0x8000 && addr < 0xC000:
		// $8000-$BFFF: Switchable bank
		offset := uint32(m.prgBank)*0x4000 + uint32(addr-0x8000)
		if int(offset) < len(m.prgROM) {
			return m.prgROM[offset]
		}

	case addr >= 0xC000:
		// $C000-$FFFF: Fixed to last bank
		lastBank := m.prgBanks - 1
		offset := uint32(lastBank)*0x4000 + uint32(addr-0xC000)
		if int(offset) < len(m.prgROM) {
			return m.prgROM[offset]
		}
	}

	return 0
}

// WritePRG handles writes to PRG space (CPU $8000-$FFFF)
// Writing to any address in $8000-$FFFF selects the PRG bank
func (m *Mapper2) WritePRG(addr uint16, value uint8) {
	if addr >= 0x8000 {
		// Select PRG bank (only lower bits used depending on ROM size)
		// Mask to valid bank number
		m.prgBank = value & (m.prgBanks - 1)
	}
}

// ReadCHR reads from CHR-RAM (PPU $0000-$1FFF)
func (m *Mapper2) ReadCHR(addr uint16) uint8 {
	if int(addr) < len(m.chrRAM) {
		return m.chrRAM[addr]
	}
	return 0
}

// WriteCHR writes to CHR-RAM (PPU $0000-$1FFF)
func (m *Mapper2) WriteCHR(addr uint16, value uint8) {
	if int(addr) < len(m.chrRAM) {
		m.chrRAM[addr] = value
	}
}

// Scanline is called by PPU on each scanline
// UxROM doesn't use scanline counting
func (m *Mapper2) Scanline() {
	// No-op for Mapper 2
}

// GetMirroring returns the nametable mirroring mode
func (m *Mapper2) GetMirroring() uint8 {
	return m.mirroring
}
