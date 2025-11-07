package cartridge

// Mapper7 implements iNES Mapper 7 (AxROM)
//
// AxROM is used by games like Battletoads, Marble Madness, Wizards & Warriors.
// It provides switchable 32KB PRG-ROM banks with single-screen mirroring control.
//
// PRG-ROM: Up to 256KB (8 banks of 32KB)
// CHR-RAM: 8KB (not switchable)
//
// CPU Memory Map:
//   $8000-$FFFF: Switchable 32 KB PRG-ROM bank
//
// PPU Memory Map:
//   $0000-$1FFF: 8 KB CHR-RAM (not switchable)
//
// Bank Switching:
//   Writing to $8000-$FFFF selects PRG bank and mirroring:
//   - Bits 0-2: Select 32KB PRG-ROM bank
//   - Bit 4: Select nametable (0 = $2000, 1 = $2400)
//
// Features:
//   - Single-screen mirroring (switchable between two nametables)
//   - No IRQ counter
//   - No PRG-RAM
type Mapper7 struct {
	prgROM []uint8 // Full PRG-ROM (all banks)
	chrRAM []uint8 // 8KB CHR-RAM

	prgBanks  uint8 // Number of 32KB PRG banks
	prgBank   uint8 // Currently selected PRG bank (0-7)
	mirroring uint8 // Single-screen mirroring (2 or 3)
}

// NewMapper7 creates a new AxROM mapper (Mapper 7)
func NewMapper7(prgROM, chrROM []uint8, mirroring uint8) *Mapper7 {
	m := &Mapper7{
		prgROM:    make([]uint8, len(prgROM)),
		chrRAM:    make([]uint8, 8192), // Always 8KB CHR-RAM
		prgBanks:  uint8(len(prgROM) / 32768), // 32KB banks
		prgBank:   0, // Start with first bank
		mirroring: MirrorSingleLow, // Default to single-screen lower
	}

	copy(m.prgROM, prgROM)

	// AxROM always uses CHR-RAM, ignore any CHR-ROM data

	return m
}

// ReadPRG reads from PRG-ROM (CPU $8000-$FFFF)
func (m *Mapper7) ReadPRG(addr uint16) uint8 {
	if addr >= 0x8000 {
		// $8000-$FFFF: Switchable 32KB bank
		offset := uint32(m.prgBank)*0x8000 + uint32(addr-0x8000)
		if int(offset) < len(m.prgROM) {
			return m.prgROM[offset]
		}
	}
	return 0
}

// WritePRG handles writes to PRG space (CPU $8000-$FFFF)
// Writing to any address in $8000-$FFFF selects PRG bank and mirroring
func (m *Mapper7) WritePRG(addr uint16, value uint8) {
	if addr >= 0x8000 {
		// Bits 0-2: Select 32KB PRG bank
		m.prgBank = value & 0x07

		// Mask to valid bank number (in case ROM has fewer banks)
		if m.prgBanks > 0 {
			m.prgBank = m.prgBank & (m.prgBanks - 1)
		}

		// Bit 4: Select single-screen mirroring
		// 0 = use nametable at $2000 (lower)
		// 1 = use nametable at $2400 (upper)
		if (value & 0x10) != 0 {
			m.mirroring = MirrorSingleHigh // Single-screen upper bank
		} else {
			m.mirroring = MirrorSingleLow // Single-screen lower bank
		}
	}
}

// ReadCHR reads from CHR-RAM (PPU $0000-$1FFF)
func (m *Mapper7) ReadCHR(addr uint16) uint8 {
	if int(addr) < len(m.chrRAM) {
		return m.chrRAM[addr]
	}
	return 0
}

// WriteCHR writes to CHR-RAM (PPU $0000-$1FFF)
func (m *Mapper7) WriteCHR(addr uint16, value uint8) {
	if int(addr) < len(m.chrRAM) {
		m.chrRAM[addr] = value
	}
}

// Scanline is called by PPU on each scanline
// AxROM doesn't use scanline counting
func (m *Mapper7) Scanline() {
	// No-op for Mapper 7
}

// GetMirroring returns the current nametable mirroring mode
func (m *Mapper7) GetMirroring() uint8 {
	return m.mirroring
}
