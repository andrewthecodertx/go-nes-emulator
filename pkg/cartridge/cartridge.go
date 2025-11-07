package cartridge

import (
	"fmt"
	"os"
)

const (
	// iNES file format constants
	inesHeaderSize = 16
	prgROMBankSize = 16384 // 16 KB
	chrROMBankSize = 8192  // 8 KB

	// iNES header magic number
	inesMagic = "NES\x1a"
)

// Mirroring modes
const (
	MirrorHorizontal = 0
	MirrorVertical   = 1
	MirrorSingleLow  = 2 // Single-screen, lower bank
	MirrorSingleHigh = 3 // Single-screen, upper bank
	MirrorFourScreen = 4
)

// Cartridge represents a loaded NES ROM cartridge
type Cartridge struct {
	mapper      Mapper
	mapperID    uint8
	prgBanks    uint8
	chrBanks    uint8
	mirroring   uint8
	hasSaveRAM  bool
	hasTrainer  bool
}

// LoadFromFile loads an iNES format ROM file (.nes)
func LoadFromFile(filename string) (*Cartridge, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read ROM file: %w", err)
	}

	return LoadFromBytes(data)
}

// LoadFromBytes parses an iNES format ROM from a byte slice
func LoadFromBytes(data []byte) (*Cartridge, error) {
	if len(data) < inesHeaderSize {
		return nil, fmt.Errorf("file too small to be a valid iNES ROM")
	}

	// Verify iNES header magic
	if string(data[0:4]) != inesMagic {
		return nil, fmt.Errorf("invalid iNES header magic: expected %q, got %q", inesMagic, string(data[0:4]))
	}

	// Parse iNES header
	header := parseINESHeader(data)

	// Calculate ROM offsets
	offset := inesHeaderSize
	if header.hasTrainer {
		offset += 512 // Skip trainer data
	}

	// Extract PRG-ROM
	prgSize := int(header.prgBanks) * prgROMBankSize
	if len(data) < offset+prgSize {
		return nil, fmt.Errorf("file too small for PRG-ROM data")
	}
	prgROM := data[offset : offset+prgSize]
	offset += prgSize

	// Extract CHR-ROM (if present)
	chrSize := int(header.chrBanks) * chrROMBankSize
	var chrROM []byte
	if chrSize > 0 {
		if len(data) < offset+chrSize {
			return nil, fmt.Errorf("file too small for CHR-ROM data")
		}
		chrROM = data[offset : offset+chrSize]
	} else {
		// No CHR-ROM means CHR-RAM will be used
		chrROM = nil
	}

	// Create appropriate mapper
	mapper, err := createMapper(header.mapperID, prgROM, chrROM, header.mirroring)
	if err != nil {
		return nil, err
	}

	return &Cartridge{
		mapper:      mapper,
		mapperID:    header.mapperID,
		prgBanks:    header.prgBanks,
		chrBanks:    header.chrBanks,
		mirroring:   header.mirroring,
		hasSaveRAM:  header.hasSaveRAM,
		hasTrainer:  header.hasTrainer,
	}, nil
}

// inesHeader represents the parsed iNES header
type inesHeader struct {
	prgBanks    uint8 // Number of 16KB PRG-ROM banks
	chrBanks    uint8 // Number of 8KB CHR-ROM banks
	mapperID    uint8 // Mapper number
	mirroring   uint8 // Nametable mirroring mode
	hasSaveRAM  bool  // Battery-backed PRG-RAM present
	hasTrainer  bool  // 512-byte trainer present
	fourScreen  bool  // Four-screen VRAM
}

// parseINESHeader extracts information from the 16-byte iNES header
func parseINESHeader(data []byte) inesHeader {
	header := inesHeader{}

	header.prgBanks = data[4]
	header.chrBanks = data[5]

	flags6 := data[6]
	flags7 := data[7]

	// Flags 6 (Mapper, mirroring, battery, trainer)
	header.mirroring = uint8(flags6 & 0x01) // 0 = horizontal, 1 = vertical
	header.hasSaveRAM = (flags6 & 0x02) != 0
	header.hasTrainer = (flags6 & 0x04) != 0
	header.fourScreen = (flags6 & 0x08) != 0

	if header.fourScreen {
		header.mirroring = MirrorFourScreen
	}

	// Mapper ID is split across flags 6 and 7
	mapperLow := (flags6 & 0xF0) >> 4
	mapperHigh := flags7 & 0xF0
	header.mapperID = mapperHigh | mapperLow

	return header
}

// createMapper instantiates the appropriate mapper for the given mapper ID
func createMapper(mapperID uint8, prgROM, chrROM []byte, mirroring uint8) (Mapper, error) {
	switch mapperID {
	case 0:
		// NROM (Mapper 0)
		// Games: Super Mario Bros., Donkey Kong, Ice Climber
		return NewMapper0(prgROM, chrROM, mirroring), nil

	case 1:
		// MMC1 (Mapper 1)
		// Games: The Legend of Zelda, Metroid, Mega Man 2, Kid Icarus
		return NewMapper1(prgROM, chrROM, mirroring), nil

	case 2:
		// UxROM (Mapper 2)
		// Games: Mega Man, Castlevania, Duck Tales, Contra
		return NewMapper2(prgROM, chrROM, mirroring), nil

	case 3:
		// CNROM (Mapper 3)
		// Games: Arkanoid, Cybernoid, Solomon's Key
		return NewMapper3(prgROM, chrROM, mirroring), nil

	case 4:
		// MMC3 (Mapper 4)
		// Games: Super Mario Bros. 2, Super Mario Bros. 3, Mega Man 3-6
		return NewMapper4(prgROM, chrROM, mirroring), nil

	case 7:
		// AxROM (Mapper 7)
		// Games: Battletoads, Marble Madness, Wizards & Warriors
		return NewMapper7(prgROM, chrROM, mirroring), nil

	default:
		return nil, fmt.Errorf("unsupported mapper: %d", mapperID)
	}
}

// GetMapper returns the cartridge's mapper
func (c *Cartridge) GetMapper() Mapper {
	return c.mapper
}

// GetMapperID returns the mapper number
func (c *Cartridge) GetMapperID() uint8 {
	return c.mapperID
}

// GetMirroring returns the nametable mirroring mode
func (c *Cartridge) GetMirroring() uint8 {
	return c.mirroring
}

// GetPRGBanks returns the number of 16KB PRG-ROM banks
func (c *Cartridge) GetPRGBanks() uint8 {
	return c.prgBanks
}

// GetCHRBanks returns the number of 8KB CHR-ROM banks
func (c *Cartridge) GetCHRBanks() uint8 {
	return c.chrBanks
}

// HasSaveRAM returns whether the cartridge has battery-backed save RAM
func (c *Cartridge) HasSaveRAM() bool {
	return c.hasSaveRAM
}
