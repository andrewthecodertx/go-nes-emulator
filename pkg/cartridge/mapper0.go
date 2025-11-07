package cartridge

// Mapper0 implements iNES Mapper 0 (NROM)
//
// NROM is the simplest mapper with no bank switching.
// PRG-ROM: 16KB or 32KB
// CHR-ROM: 8KB (or CHR-RAM if no CHR-ROM present)
//
// CPU Memory Map:
//   $6000-$7FFF: Family Basic only (not implemented)
//   $8000-$BFFF: First 16 KB of ROM
//   $C000-$FFFF: Last 16 KB of ROM (mirror of first 16KB if only one bank)
//
// PPU Memory Map:
//   $0000-$1FFF: 8 KB CHR-ROM or CHR-RAM
type Mapper0 struct {
	prgROM []uint8 // PRG-ROM (16KB or 32KB)
	chrMem []uint8 // CHR-ROM or CHR-RAM (8KB)

	prgBanks    uint8 // Number of 16KB PRG banks (1 or 2)
	chrIsRAM    bool  // True if using CHR-RAM instead of CHR-ROM
	mirroring   uint8 // Nametable mirroring mode
}

// NewMapper0 creates a new NROM mapper (Mapper 0)
func NewMapper0(prgROM, chrROM []uint8, mirroring uint8) *Mapper0 {
	m := &Mapper0{
		prgROM:    make([]uint8, len(prgROM)),
		mirroring: mirroring,
	}

	copy(m.prgROM, prgROM)

	// Determine number of 16KB PRG banks
	m.prgBanks = uint8(len(prgROM) / 16384)

	// Setup CHR memory (ROM or RAM)
	if len(chrROM) > 0 {
		// CHR-ROM present
		m.chrMem = make([]uint8, len(chrROM))
		copy(m.chrMem, chrROM)
		m.chrIsRAM = false
	} else {
		// No CHR-ROM, use 8KB CHR-RAM
		m.chrMem = make([]uint8, 8192)
		m.chrIsRAM = true
	}

	return m
}

// ReadPRG reads from PRG-ROM (CPU $8000-$FFFF)
func (m *Mapper0) ReadPRG(addr uint16) uint8 {
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
// NROM has no PRG-RAM or mapper registers, so writes are ignored
func (m *Mapper0) WritePRG(addr uint16, value uint8) {
	// NROM has no writable PRG space
	// Writes are ignored
}

// ReadCHR reads from CHR-ROM/RAM (PPU $0000-$1FFF)
func (m *Mapper0) ReadCHR(addr uint16) uint8 {
	if int(addr) < len(m.chrMem) {
		return m.chrMem[addr]
	}
	return 0
}

// WriteCHR writes to CHR-RAM (PPU $0000-$1FFF)
func (m *Mapper0) WriteCHR(addr uint16, value uint8) {
	if m.chrIsRAM && int(addr) < len(m.chrMem) {
		m.chrMem[addr] = value
	}
	// CHR-ROM writes are ignored
}

// Scanline is called by PPU on each scanline
// NROM doesn't use scanline counting
func (m *Mapper0) Scanline() {
	// No-op for Mapper 0
}

// GetMirroring returns the nametable mirroring mode
func (m *Mapper0) GetMirroring() uint8 {
	return m.mirroring
}
