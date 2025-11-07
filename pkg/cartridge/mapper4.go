package cartridge

// Mapper4 implements iNES Mapper 4 (MMC3)
//
// MMC3 is the most common mapper (~23% of games).
// Used by: Super Mario Bros. 2, Super Mario Bros. 3, Mega Man 3-6, etc.
//
// Features:
// - 2x 8KB switchable PRG-ROM banks + 1x 8KB fixed bank
// - 6x switchable CHR banks (2x 2KB + 4x 1KB) or CHR-RAM
// - Configurable PRG/CHR bank arrangement
// - Scanline counter with IRQ generation (for split-screen effects)
// - Optional PRG-RAM (8KB, may be battery-backed)
//
// PRG-ROM: Up to 512KB (64 banks of 8KB)
// CHR-ROM: Up to 256KB (256 banks of 1KB)
// PRG-RAM: 8KB at $6000-$7FFF (optional)
//
// CPU Memory Map:
//   $6000-$7FFF: 8KB PRG-RAM (optional, battery-backed save RAM)
//   $8000-$9FFF: 8 KB switchable PRG-ROM bank (or fixed to second-last bank)
//   $A000-$BFFF: 8 KB switchable PRG-ROM bank
//   $C000-$DFFF: 8 KB switchable PRG-ROM bank (or fixed to second-last bank)
//   $E000-$FFFF: 8 KB PRG-ROM bank (fixed to last bank)
//
// PPU Memory Map:
//   $0000-$07FF: 2KB switchable CHR bank
//   $0800-$0FFF: 2KB switchable CHR bank
//   $1000-$13FF: 1KB switchable CHR bank
//   $1400-$17FF: 1KB switchable CHR bank
//   $1800-$1BFF: 1KB switchable CHR bank
//   $1C00-$1FFF: 1KB switchable CHR bank
//
// Registers (all at $8000-$FFFF, even/odd addresses):
//   $8000-$9FFE (even): Bank select
//   $8001-$9FFF (odd):  Bank data
//   $A000-$BFFE (even): Mirroring
//   $A001-$BFFF (odd):  PRG-RAM protect
//   $C000-$DFFE (even): IRQ latch
//   $C001-$DFFF (odd):  IRQ reload
//   $E000-$FFFE (even): IRQ disable
//   $E001-$FFFF (odd):  IRQ enable
type Mapper4 struct {
	prgROM []uint8 // Full PRG-ROM
	chrMem []uint8 // CHR-ROM or CHR-RAM
	prgRAM []uint8 // 8KB PRG-RAM

	prgBanks uint8 // Number of 8KB PRG banks
	chrBanks uint8 // Number of 1KB CHR banks
	chrIsRAM bool  // True if using CHR-RAM

	// Bank select register
	bankSelect uint8 // Which bank register to update (0-7)
	prgMode    uint8 // PRG bank mode (0 or 1)
	chrMode    uint8 // CHR A12 inversion (0 or 1)

	// Bank registers (selected by bankSelect)
	registers [8]uint8 // R0-R7: bank numbers

	// Mirroring
	mirroring uint8 // 0=vertical, 1=horizontal

	// PRG-RAM protection
	prgRAMEnabled      bool // PRG-RAM chip enable
	prgRAMWriteProtect bool // PRG-RAM write protect

	// IRQ
	irqLatch       uint8 // IRQ counter reload value
	irqCounter     uint8 // IRQ counter (counts down)
	irqEnabled     bool  // IRQ enable flag
	irqPending     bool  // IRQ pending flag
	irqReloadFlag  bool  // IRQ reload flag (set when counter should reload)
}

// NewMapper4 creates a new MMC3 mapper (Mapper 4)
func NewMapper4(prgROM, chrROM []uint8, mirroring uint8) *Mapper4 {
	m := &Mapper4{
		prgROM:        make([]uint8, len(prgROM)),
		prgRAM:        make([]uint8, 8192),
		prgBanks:      uint8(len(prgROM) / 8192), // 8KB banks
		mirroring:     mirroring,
		prgRAMEnabled: true,
	}

	copy(m.prgROM, prgROM)

	// Setup CHR memory (ROM or RAM)
	if len(chrROM) > 0 {
		m.chrMem = make([]uint8, len(chrROM))
		copy(m.chrMem, chrROM)
		m.chrBanks = uint8(len(chrROM) / 1024) // 1KB banks
		m.chrIsRAM = false
	} else {
		// No CHR-ROM, use 8KB CHR-RAM
		m.chrMem = make([]uint8, 8192)
		m.chrBanks = 8 // Eight 1KB banks
		m.chrIsRAM = true
	}

	return m
}

// ReadPRG reads from PRG space (CPU $6000-$FFFF)
func (m *Mapper4) ReadPRG(addr uint16) uint8 {
	switch {
	case addr >= 0x6000 && addr < 0x8000:
		// $6000-$7FFF: PRG-RAM
		if m.prgRAMEnabled {
			return m.prgRAM[addr-0x6000]
		}
		return 0

	case addr >= 0x8000 && addr < 0xA000:
		// $8000-$9FFF
		var bank uint8
		if m.prgMode == 0 {
			bank = m.registers[6] // R6: swappable
		} else {
			bank = m.prgBanks - 2 // Fixed to second-last bank
		}
		offset := uint32(bank)*0x2000 + uint32(addr-0x8000)
		if int(offset) < len(m.prgROM) {
			return m.prgROM[offset]
		}

	case addr >= 0xA000 && addr < 0xC000:
		// $A000-$BFFF: R7 (always swappable)
		bank := m.registers[7]
		offset := uint32(bank)*0x2000 + uint32(addr-0xA000)
		if int(offset) < len(m.prgROM) {
			return m.prgROM[offset]
		}

	case addr >= 0xC000 && addr < 0xE000:
		// $C000-$DFFF
		var bank uint8
		if m.prgMode == 0 {
			bank = m.prgBanks - 2 // Fixed to second-last bank
		} else {
			bank = m.registers[6] // R6: swappable
		}
		offset := uint32(bank)*0x2000 + uint32(addr-0xC000)
		if int(offset) < len(m.prgROM) {
			return m.prgROM[offset]
		}

	case addr >= 0xE000:
		// $E000-$FFFF: Fixed to last bank
		bank := m.prgBanks - 1
		offset := uint32(bank)*0x2000 + uint32(addr-0xE000)
		if int(offset) < len(m.prgROM) {
			return m.prgROM[offset]
		}
	}

	return 0
}

// WritePRG handles writes to PRG space (CPU $6000-$FFFF)
func (m *Mapper4) WritePRG(addr uint16, value uint8) {
	switch {
	case addr >= 0x6000 && addr < 0x8000:
		// $6000-$7FFF: PRG-RAM
		if m.prgRAMEnabled && !m.prgRAMWriteProtect {
			m.prgRAM[addr-0x6000] = value
		}

	case addr >= 0x8000 && addr < 0xA000:
		if (addr & 1) == 0 {
			// $8000, $8002, ..., $9FFE: Bank select
			m.bankSelect = value & 0x07
			m.prgMode = (value >> 6) & 0x01
			m.chrMode = (value >> 7) & 0x01
		} else {
			// $8001, $8003, ..., $9FFF: Bank data
			m.registers[m.bankSelect] = value
		}

	case addr >= 0xA000 && addr < 0xC000:
		if (addr & 1) == 0 {
			// $A000, $A002, ..., $BFFE: Mirroring
			if (value & 1) == 0 {
				m.mirroring = MirrorVertical
			} else {
				m.mirroring = MirrorHorizontal
			}
		} else {
			// $A001, $A003, ..., $BFFF: PRG-RAM protect
			m.prgRAMWriteProtect = (value & 0x40) != 0
			m.prgRAMEnabled = (value & 0x80) != 0
		}

	case addr >= 0xC000 && addr < 0xE000:
		if (addr & 1) == 0 {
			// $C000, $C002, ..., $DFFE: IRQ latch
			m.irqLatch = value
		} else {
			// $C001, $C003, ..., $DFFF: IRQ reload
			m.irqCounter = 0
			m.irqReloadFlag = true
		}

	case addr >= 0xE000:
		if (addr & 1) == 0 {
			// $E000, $E002, ..., $FFFE: IRQ disable
			m.irqEnabled = false
			m.irqPending = false
		} else {
			// $E001, $E003, ..., $FFFF: IRQ enable
			m.irqEnabled = true
		}
	}
}

// ReadCHR reads from CHR-ROM/RAM (PPU $0000-$1FFF)
func (m *Mapper4) ReadCHR(addr uint16) uint8 {
	var bank uint8
	var offset uint32

	// CHR mode determines bank arrangement
	if m.chrMode == 0 {
		// Mode 0: 2KB banks at $0000, 1KB banks at $1000
		switch {
		case addr < 0x0800:
			// $0000-$07FF: R0 (2KB)
			bank = m.registers[0] & 0xFE
			offset = uint32(bank)*0x400 + uint32(addr)
		case addr < 0x1000:
			// $0800-$0FFF: R1 (2KB)
			bank = m.registers[1] & 0xFE
			offset = uint32(bank)*0x400 + uint32(addr-0x0800)
		case addr < 0x1400:
			// $1000-$13FF: R2 (1KB)
			bank = m.registers[2]
			offset = uint32(bank)*0x400 + uint32(addr-0x1000)
		case addr < 0x1800:
			// $1400-$17FF: R3 (1KB)
			bank = m.registers[3]
			offset = uint32(bank)*0x400 + uint32(addr-0x1400)
		case addr < 0x1C00:
			// $1800-$1BFF: R4 (1KB)
			bank = m.registers[4]
			offset = uint32(bank)*0x400 + uint32(addr-0x1800)
		default:
			// $1C00-$1FFF: R5 (1KB)
			bank = m.registers[5]
			offset = uint32(bank)*0x400 + uint32(addr-0x1C00)
		}
	} else {
		// Mode 1: 2KB banks at $1000, 1KB banks at $0000
		switch {
		case addr < 0x0400:
			// $0000-$03FF: R2 (1KB)
			bank = m.registers[2]
			offset = uint32(bank)*0x400 + uint32(addr)
		case addr < 0x0800:
			// $0400-$07FF: R3 (1KB)
			bank = m.registers[3]
			offset = uint32(bank)*0x400 + uint32(addr-0x0400)
		case addr < 0x0C00:
			// $0800-$0BFF: R4 (1KB)
			bank = m.registers[4]
			offset = uint32(bank)*0x400 + uint32(addr-0x0800)
		case addr < 0x1000:
			// $0C00-$0FFF: R5 (1KB)
			bank = m.registers[5]
			offset = uint32(bank)*0x400 + uint32(addr-0x0C00)
		case addr < 0x1800:
			// $1000-$17FF: R0 (2KB)
			bank = m.registers[0] & 0xFE
			offset = uint32(bank)*0x400 + uint32(addr-0x1000)
		default:
			// $1800-$1FFF: R1 (2KB)
			bank = m.registers[1] & 0xFE
			offset = uint32(bank)*0x400 + uint32(addr-0x1800)
		}
	}

	if int(offset) < len(m.chrMem) {
		return m.chrMem[offset]
	}
	return 0
}

// WriteCHR writes to CHR-RAM (PPU $0000-$1FFF)
func (m *Mapper4) WriteCHR(addr uint16, value uint8) {
	if !m.chrIsRAM {
		return // CHR-ROM is read-only
	}

	var bank uint8
	var offset uint32

	// Same bank calculation as ReadCHR
	if m.chrMode == 0 {
		switch {
		case addr < 0x0800:
			bank = m.registers[0] & 0xFE
			offset = uint32(bank)*0x400 + uint32(addr)
		case addr < 0x1000:
			bank = m.registers[1] & 0xFE
			offset = uint32(bank)*0x400 + uint32(addr-0x0800)
		case addr < 0x1400:
			bank = m.registers[2]
			offset = uint32(bank)*0x400 + uint32(addr-0x1000)
		case addr < 0x1800:
			bank = m.registers[3]
			offset = uint32(bank)*0x400 + uint32(addr-0x1400)
		case addr < 0x1C00:
			bank = m.registers[4]
			offset = uint32(bank)*0x400 + uint32(addr-0x1800)
		default:
			bank = m.registers[5]
			offset = uint32(bank)*0x400 + uint32(addr-0x1C00)
		}
	} else {
		switch {
		case addr < 0x0400:
			bank = m.registers[2]
			offset = uint32(bank)*0x400 + uint32(addr)
		case addr < 0x0800:
			bank = m.registers[3]
			offset = uint32(bank)*0x400 + uint32(addr-0x0400)
		case addr < 0x0C00:
			bank = m.registers[4]
			offset = uint32(bank)*0x400 + uint32(addr-0x0800)
		case addr < 0x1000:
			bank = m.registers[5]
			offset = uint32(bank)*0x400 + uint32(addr-0x0C00)
		case addr < 0x1800:
			bank = m.registers[0] & 0xFE
			offset = uint32(bank)*0x400 + uint32(addr-0x1000)
		default:
			bank = m.registers[1] & 0xFE
			offset = uint32(bank)*0x400 + uint32(addr-0x1800)
		}
	}

	if int(offset) < len(m.chrMem) {
		m.chrMem[offset] = value
	}
}

// Scanline is called by PPU on each scanline
// MMC3 uses this for IRQ generation
func (m *Mapper4) Scanline() {
	if m.irqCounter == 0 || m.irqReloadFlag {
		// Reload counter
		m.irqCounter = m.irqLatch
		m.irqReloadFlag = false
	} else {
		// Decrement counter
		m.irqCounter--
	}

	if m.irqCounter == 0 && m.irqEnabled {
		// Trigger IRQ
		m.irqPending = true
	}
}

// GetMirroring returns the current nametable mirroring mode
func (m *Mapper4) GetMirroring() uint8 {
	return m.mirroring
}

// IRQPending returns true if an IRQ is pending
// The emulator should check this and trigger a CPU IRQ
func (m *Mapper4) IRQPending() bool {
	return m.irqPending
}

// ClearIRQ clears the IRQ pending flag
func (m *Mapper4) ClearIRQ() {
	m.irqPending = false
}
