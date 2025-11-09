package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/go-nes-emulator/pkg/cartridge"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: rom-info <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	// Read the ROM file
	data, err := os.ReadFile(romPath)
	if err != nil {
		fmt.Printf("Error reading ROM: %v\n", err)
		os.Exit(1)
	}

	if len(data) < 16 {
		fmt.Println("File too small to be a valid iNES ROM")
		os.Exit(1)
	}

	// Parse header
	fmt.Printf("ROM File: %s\n", romPath)
	fmt.Printf("File Size: %d bytes\n\n", len(data))

	// Check magic
	magic := string(data[0:4])
	fmt.Printf("Magic: %q (should be \"NES\\x1a\")\n", magic)

	prgBanks := data[4]
	chrBanks := data[5]
	flags6 := data[6]
	flags7 := data[7]

	fmt.Printf("PRG-ROM Banks: %d (= %d KB)\n", prgBanks, prgBanks*16)
	fmt.Printf("CHR-ROM Banks: %d (= %d KB)\n", chrBanks, chrBanks*8)

	// Parse flags
	mirroring := flags6 & 0x01
	hasSaveRAM := (flags6 & 0x02) != 0
	hasTrainer := (flags6 & 0x04) != 0
	fourScreen := (flags6 & 0x08) != 0

	mapperLow := (flags6 & 0xF0) >> 4
	mapperHigh := flags7 & 0xF0
	mapperID := mapperHigh | mapperLow

	fmt.Printf("\nFlags 6: 0x%02X\n", flags6)
	fmt.Printf("  Mirroring: %s (%d)\n", []string{"Horizontal", "Vertical"}[mirroring], mirroring)
	fmt.Printf("  Battery-backed RAM: %v\n", hasSaveRAM)
	fmt.Printf("  Trainer: %v\n", hasTrainer)
	fmt.Printf("  Four-screen VRAM: %v\n", fourScreen)
	fmt.Printf("  Mapper (low nibble): %d\n", mapperLow)

	fmt.Printf("\nFlags 7: 0x%02X\n", flags7)
	fmt.Printf("  Mapper (high nibble): %d\n", mapperHigh>>4)

	fmt.Printf("\nMapper ID: %d\n", mapperID)

	// Try to load with cartridge loader
	fmt.Println("\nAttempting to load with cartridge loader...")
	cart, err := cartridge.LoadFromFile(romPath)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: Loaded mapper %d\n", cart.GetMapperID())
	}
}
