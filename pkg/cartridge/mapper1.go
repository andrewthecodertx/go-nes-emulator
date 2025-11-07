package cartridge

// Mapper1 implements iNES Mapper 1 (MMC1)
//
// MMC1 is used by games like The Legend of Zelda, Metroid, Mega Man 2, Kid Icarus.
// It's one of the most common mappers (~28% of games).
//
// Features:
// - Switchable PRG-ROM banks (16KB or 32KB mode)
// - Switchable CHR-ROM banks (4KB or 8KB mode)
// - Configurable mirroring
// - Optional PRG-RAM (8KB, may be battery-backed)
//
// PRG-ROM: Up to 512KB (32 banks of 16KB)
// CHR-ROM: Up to 128KB (32 banks of 4KB)
// PRG-RAM: 8KB at $6000-$7FFF (optional)
//
// CPU Memory Map:
//   $6000-$7FFF: 8KB PRG-RAM (optional, may be battery-backed)
//   $8000-$BFFF: 16 KB PRG-ROM bank (switchable or fixed depending on mode)
//   $C000-$FFFF: 16 KB PRG-ROM bank (switchable or fixed depending on mode)
//
// PPU Memory Map:
//   $0000-$0FFF: 4 KB CHR bank (switchable in 4KB mode, or part of 8KB bank)
//   $1000-$1FFF: 4 KB CHR bank (switchable in 4KB mode, or part of 8KB bank)
//
// Control via Shift Register:
//   MMC1 uses a 5-bit shift register for all control writes.
//   Write to $8000-$FFFF with bit 7 set: Reset shift register
//   Write to $8000-$FFFF with bit 7 clear: Shift bit 0 into register
//   After 5 writes, the shift register value is copied to internal register
//
// Registers:
//   $8000-$9FFF: Control (mirroring, PRG/CHR mode)
//   $A000-$BFFF: CHR bank 0
//   $C000-$DFFF: CHR bank 1
//   $E000-$FFFF: PRG bank
type Mapper1 struct {
	prgROM []uint8 // Full PRG-ROM (all banks)
	chrMem []uint8 // CHR-ROM or CHR-RAM
	prgRAM []uint8 // 8KB PRG-RAM at $6000-$7FFF

	prgBanks uint8 // Number of 16KB PRG banks
	chrBanks uint8 // Number of 4KB CHR banks
	chrIsRAM bool  // True if using CHR-RAM

	// Shift register state
	shiftRegister uint8 // 5-bit shift register
	shiftCount    uint8 // Number of writes to shift register (0-4)

	// Control register ($8000-$9FFF)
	mirroring  uint8 // 0=one-screen-lower, 1=one-screen-upper, 2=vertical, 3=horizontal
	prgMode    uint8 // 0/1=32KB mode, 2=fix first bank, 3=fix last bank
	chrMode    uint8 // 0=8KB mode, 1=4KB mode

	// CHR bank registers
	chrBank0 uint8 // $A000-$BFFF: CHR bank 0 (or full 8KB in 8KB mode)
	chrBank1 uint8 // $C000-$DFFF: CHR bank 1 (ignored in 8KB mode)

	// PRG bank register
	prgBank uint8 // $E000-$FFFF: PRG bank select

	// PRG-RAM control
	prgRAMEnabled bool // PRG-RAM chip enable (not always implemented)
}

// NewMapper1 creates a new MMC1 mapper (Mapper 1)
func NewMapper1(prgROM, chrROM []uint8, mirroring uint8) *Mapper1 {
	m := &Mapper1{
		prgROM:        make([]uint8, len(prgROM)),
		prgRAM:        make([]uint8, 8192), // 8KB PRG-RAM
		prgBanks:      uint8(len(prgROM) / 16384),
		shiftRegister: 0x10, // Reset state
		prgMode:       3,    // Default: fix last bank
		mirroring:     mirroring,
		prgRAMEnabled: true,
	}

	copy(m.prgROM, prgROM)

	// Setup CHR memory (ROM or RAM)
	if len(chrROM) > 0 {
		// CHR-ROM present
		m.chrMem = make([]uint8, len(chrROM))
		copy(m.chrMem, chrROM)
		m.chrBanks = uint8(len(chrROM) / 4096) // 4KB banks
		m.chrIsRAM = false
	} else {
		// No CHR-ROM, use 8KB CHR-RAM
		m.chrMem = make([]uint8, 8192)
		m.chrBanks = 2 // Two 4KB banks
		m.chrIsRAM = true
	}

	return m
}

// ReadPRG reads from PRG space (CPU $6000-$FFFF)
func (m *Mapper1) ReadPRG(addr uint16) uint8 {
	switch {
	case addr >= 0x6000 && addr < 0x8000:
		// $6000-$7FFF: PRG-RAM
		if m.prgRAMEnabled {
			return m.prgRAM[addr-0x6000]
		}
		return 0

	case addr >= 0x8000 && addr < 0xC000:
		// $8000-$BFFF: First PRG bank
		var bank uint8
		switch m.prgMode {
		case 0, 1:
			// 32KB mode: ignore bit 0 of prgBank
			bank = (m.prgBank & 0xFE)
		case 2:
			// Fix first bank at $8000
			bank = 0
		case 3:
			// Switch 16KB bank at $8000
			bank = m.prgBank
		}
		offset := uint32(bank)*0x4000 + uint32(addr-0x8000)
		if int(offset) < len(m.prgROM) {
			return m.prgROM[offset]
		}

	case addr >= 0xC000:
		// $C000-$FFFF: Second PRG bank
		var bank uint8
		switch m.prgMode {
		case 0, 1:
			// 32KB mode: use odd bank
			bank = (m.prgBank & 0xFE) | 1
		case 2:
			// Switch 16KB bank at $C000
			bank = m.prgBank
		case 3:
			// Fix last bank at $C000
			bank = m.prgBanks - 1
		}
		offset := uint32(bank)*0x4000 + uint32(addr-0xC000)
		if int(offset) < len(m.prgROM) {
			return m.prgROM[offset]
		}
	}

	return 0
}

// WritePRG handles writes to PRG space (CPU $6000-$FFFF)
func (m *Mapper1) WritePRG(addr uint16, value uint8) {
	switch {
	case addr >= 0x6000 && addr < 0x8000:
		// $6000-$7FFF: PRG-RAM
		if m.prgRAMEnabled {
			m.prgRAM[addr-0x6000] = value
		}

	case addr >= 0x8000:
		// $8000-$FFFF: Shift register / control registers
		if (value & 0x80) != 0 {
			// Bit 7 set: Reset shift register
			m.shiftRegister = 0x10
			m.shiftCount = 0
			// Also set control to mode 3 (fix last bank)
			m.prgMode = 3
		} else {
			// Bit 7 clear: Shift bit 0 into register
			m.shiftRegister >>= 1
			m.shiftRegister |= (value & 1) << 4
			m.shiftCount++

			if m.shiftCount == 5 {
				// 5 writes complete: update internal register
				m.writeRegister(addr, m.shiftRegister)
				// Reset shift register
				m.shiftRegister = 0x10
				m.shiftCount = 0
			}
		}
	}
}

// writeRegister writes to MMC1 internal registers after shift register fills
func (m *Mapper1) writeRegister(addr uint16, value uint8) {
	switch {
	case addr >= 0x8000 && addr < 0xA000:
		// $8000-$9FFF: Control register
		m.mirroring = value & 0x03
		m.prgMode = (value >> 2) & 0x03
		m.chrMode = (value >> 4) & 0x01

	case addr >= 0xA000 && addr < 0xC000:
		// $A000-$BFFF: CHR bank 0
		m.chrBank0 = value & 0x1F

	case addr >= 0xC000 && addr < 0xE000:
		// $C000-$DFFF: CHR bank 1
		m.chrBank1 = value & 0x1F

	case addr >= 0xE000:
		// $E000-$FFFF: PRG bank
		m.prgBank = value & 0x0F
		m.prgRAMEnabled = (value & 0x10) == 0 // Bit 4 disables PRG-RAM
	}
}

// ReadCHR reads from CHR-ROM/RAM (PPU $0000-$1FFF)
func (m *Mapper1) ReadCHR(addr uint16) uint8 {
	var bank uint8
	var offset uint32

	if m.chrMode == 0 {
		// 8KB mode: use chrBank0, ignore bit 0
		bank = m.chrBank0 & 0xFE
		if addr >= 0x1000 {
			bank |= 1
		}
		offset = uint32(bank)*0x1000 + uint32(addr&0x0FFF)
	} else {
		// 4KB mode: separate banks
		if addr < 0x1000 {
			bank = m.chrBank0
			offset = uint32(bank)*0x1000 + uint32(addr)
		} else {
			bank = m.chrBank1
			offset = uint32(bank)*0x1000 + uint32(addr-0x1000)
		}
	}

	if int(offset) < len(m.chrMem) {
		return m.chrMem[offset]
	}
	return 0
}

// WriteCHR writes to CHR-RAM (PPU $0000-$1FFF)
func (m *Mapper1) WriteCHR(addr uint16, value uint8) {
	if !m.chrIsRAM {
		return // CHR-ROM is read-only
	}

	var bank uint8
	var offset uint32

	if m.chrMode == 0 {
		// 8KB mode
		bank = m.chrBank0 & 0xFE
		if addr >= 0x1000 {
			bank |= 1
		}
		offset = uint32(bank)*0x1000 + uint32(addr&0x0FFF)
	} else {
		// 4KB mode
		if addr < 0x1000 {
			bank = m.chrBank0
			offset = uint32(bank)*0x1000 + uint32(addr)
		} else {
			bank = m.chrBank1
			offset = uint32(bank)*0x1000 + uint32(addr-0x1000)
		}
	}

	if int(offset) < len(m.chrMem) {
		m.chrMem[offset] = value
	}
}

// Scanline is called by PPU on each scanline
// MMC1 doesn't use scanline counting
func (m *Mapper1) Scanline() {
	// No-op for Mapper 1
}

// GetMirroring returns the current nametable mirroring mode
func (m *Mapper1) GetMirroring() uint8 {
	// MMC1 can change mirroring dynamically
	switch m.mirroring {
	case 0:
		return 2 // One-screen, lower bank (map to single-low)
	case 1:
		return 3 // One-screen, upper bank (map to single-high)
	case 2:
		return MirrorVertical
	case 3:
		return MirrorHorizontal
	}
	return MirrorHorizontal
}
