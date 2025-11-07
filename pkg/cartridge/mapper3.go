package cartridge

// Mapper3 implements iNES Mapper 3 (CNROM)
//
// CNROM is used by games like Arkanoid, Cybernoid, Solomon's Key.
// It provides fixed PRG-ROM with switchable CHR-ROM banks.
//
// PRG-ROM: 16KB or 32KB (no bank switching)
// CHR-ROM: Up to 32KB (up to 4 banks of 8KB)
//
// CPU Memory Map:
//   $8000-$BFFF: First 16 KB of ROM
//   $C000-$FFFF: Last 16 KB of ROM (mirror of first 16KB if only one bank)
//
// PPU Memory Map:
//   $0000-$1FFF: Switchable 8 KB CHR-ROM bank
//
// Bank Switching:
//   Writing to $8000-$FFFF selects which 8KB CHR bank appears at $0000-$1FFF
//   Only the lower 2 bits are typically used (4 banks max)
type Mapper3 struct {
	prgROM []uint8 // PRG-ROM (16KB or 32KB)
	chrROM []uint8 // Full CHR-ROM (all banks)

	prgBanks  uint8 // Number of 16KB PRG banks (1 or 2)
	chrBanks  uint8 // Number of 8KB CHR banks
	chrBank   uint8 // Currently selected CHR bank
	mirroring uint8 // Nametable mirroring mode
}

// NewMapper3 creates a new CNROM mapper (Mapper 3)
func NewMapper3(prgROM, chrROM []uint8, mirroring uint8) *Mapper3 {
	m := &Mapper3{
		prgROM:    make([]uint8, len(prgROM)),
		chrROM:    make([]uint8, len(chrROM)),
		prgBanks:  uint8(len(prgROM) / 16384),
		chrBanks:  uint8(len(chrROM) / 8192),
		chrBank:   0, // Start with first bank
		mirroring: mirroring,
	}

	copy(m.prgROM, prgROM)
	copy(m.chrROM, chrROM)

	return m
}

// ReadPRG reads from PRG-ROM (CPU $8000-$FFFF)
func (m *Mapper3) ReadPRG(addr uint16) uint8 {
	// Map $8000-$FFFF to ROM
	addr -= 0x8000

	if m.prgBanks == 1 {
		// 16KB ROM: mirror $C000-$FFFF to $8000-$BFFF
		addr %= 0x4000
	}

	if int(addr) < len(m.prgROM) {
		return m.prgROM[addr]
	}

	return 0
}

// WritePRG handles writes to PRG space (CPU $8000-$FFFF)
// Writing to any address in $8000-$FFFF selects the CHR bank
func (m *Mapper3) WritePRG(addr uint16, value uint8) {
	if addr >= 0x8000 {
		// Select CHR bank (only lower 2 bits typically used)
		// Mask to valid bank number
		if m.chrBanks > 0 {
			m.chrBank = value & (m.chrBanks - 1)
		}
	}
}

// ReadCHR reads from CHR-ROM (PPU $0000-$1FFF)
func (m *Mapper3) ReadCHR(addr uint16) uint8 {
	// Calculate offset into CHR-ROM based on selected bank
	offset := uint32(m.chrBank)*0x2000 + uint32(addr)
	if int(offset) < len(m.chrROM) {
		return m.chrROM[offset]
	}
	return 0
}

// WriteCHR handles writes to CHR space (PPU $0000-$1FFF)
// CHR-ROM is read-only, writes are ignored
func (m *Mapper3) WriteCHR(addr uint16, value uint8) {
	// CHR-ROM writes are ignored
}

// Scanline is called by PPU on each scanline
// CNROM doesn't use scanline counting
func (m *Mapper3) Scanline() {
	// No-op for Mapper 3
}

// GetMirroring returns the nametable mirroring mode
func (m *Mapper3) GetMirroring() uint8 {
	return m.mirroring
}
