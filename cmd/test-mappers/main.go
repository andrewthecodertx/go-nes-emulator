package main

import (
	"fmt"
	"log"

	"github.com/andrewthecodertx/nes-emulator/pkg/cartridge"
)

func main() {
	fmt.Println("Testing NES Mapper Support")
	fmt.Println("===========================\n")

	// Test each mapper with minimal ROM data
	testMappers := []struct {
		id          uint8
		name        string
		description string
	}{
		{0, "NROM", "Super Mario Bros., Donkey Kong"},
		{1, "MMC1", "The Legend of Zelda, Metroid, Tetris"},
		{2, "UxROM", "Mega Man, Castlevania, Contra"},
		{3, "CNROM", "Arkanoid, Cybernoid"},
		{4, "MMC3", "Super Mario Bros. 3, Mega Man 3-6"},
		{7, "AxROM", "Battletoads, Marble Madness, Wizards & Warriors"},
	}

	for _, tm := range testMappers {
		// Create minimal iNES ROM with this mapper
		header := []byte{
			'N', 'E', 'S', 0x1A, // Magic
			1,           // 1 PRG bank (16KB)
			1,           // 1 CHR bank (8KB)
			tm.id << 4,  // Mapper ID in upper nibble of byte 6
			0,           // Flags 7
			0, 0, 0, 0,  // Padding
			0, 0, 0, 0,  // Padding
		}

		// Add minimal PRG-ROM (16KB)
		prgROM := make([]byte, 16384)
		// Add reset vector at $FFFC-$FFFD pointing to $8000
		prgROM[0x3FFC] = 0x00
		prgROM[0x3FFD] = 0x80

		// Add minimal CHR-ROM (8KB)
		chrROM := make([]byte, 8192)

		// Combine header + PRG + CHR
		romData := append(header, prgROM...)
		romData = append(romData, chrROM...)

		// Try to load the ROM
		cart, err := cartridge.LoadFromBytes(romData)
		if err != nil {
			log.Printf("❌ Mapper %d (%s): FAILED - %v\n", tm.id, tm.name, err)
			continue
		}

		// Verify mapper ID
		if cart.GetMapperID() != tm.id {
			log.Printf("❌ Mapper %d (%s): FAILED - Got mapper %d\n", tm.id, tm.name, cart.GetMapperID())
			continue
		}

		fmt.Printf("✅ Mapper %d (%s): PASSED\n", tm.id, tm.name)
		fmt.Printf("   Games: %s\n\n", tm.description)
	}

	// Test unsupported mapper
	fmt.Println("Testing unsupported mapper:")
	unsupportedHeader := []byte{
		'N', 'E', 'S', 0x1A, // Magic
		1,                  // 1 PRG bank
		1,                  // 1 CHR bank
		(9 << 4),           // Mapper 9 (not implemented)
		0,                  // Flags 7
		0, 0, 0, 0, 0, 0, 0, 0, // Padding
	}
	prgROM := make([]byte, 16384)
	chrROM := make([]byte, 8192)
	unsupportedROM := append(unsupportedHeader, prgROM...)
	unsupportedROM = append(unsupportedROM, chrROM...)

	_, err := cartridge.LoadFromBytes(unsupportedROM)
	if err != nil {
		fmt.Printf("✅ Mapper 9: Correctly rejected - %v\n", err)
	} else {
		fmt.Printf("❌ Mapper 9: Should have been rejected but loaded\n")
	}

	fmt.Println("\n===========================")
	fmt.Println("Mapper test complete!")
	fmt.Println("\nSupported mappers cover ~72% of NES games:")
	fmt.Println("  Mapper 0 (NROM):  ~10% of games")
	fmt.Println("  Mapper 1 (MMC1):  ~28% of games")
	fmt.Println("  Mapper 2 (UxROM): ~11% of games")
	fmt.Println("  Mapper 3 (CNROM):  ~7% of games")
	fmt.Println("  Mapper 4 (MMC3):  ~23% of games")
	fmt.Println("  Mapper 7 (AxROM):  ~2% of games")
}
