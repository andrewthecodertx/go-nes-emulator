package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dump-nametable <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	// Load NES ROM
	fmt.Printf("Loading ROM: %s\n", romPath)
	emulator, err := nes.New(romPath)
	if err != nil {
		log.Fatalf("Failed to load ROM: %v", err)
	}

	// Reset and run
	emulator.Reset()

	fmt.Println("Running 120 frames...")
	for i := 0; i < 120; i++ {
		emulator.RunFrame()
	}

	ppuUnit := emulator.GetPPU()

	// Try to force rendering
	ppuUnit.WriteCPURegister(0x2001, 0x1E)

	// Run a few more frames
	for i := 0; i < 10; i++ {
		emulator.RunFrame()
	}

	fmt.Printf("\nPPU Registers:\n")
	fmt.Printf("  PPUCTRL:   $%02X\n", ppuUnit.ReadCPURegister(0x2000))
	fmt.Printf("  PPUMASK:   $%02X\n", ppuUnit.ReadCPURegister(0x2001))
	fmt.Printf("  PPUSTATUS: $%02X\n", ppuUnit.ReadCPURegister(0x2002))

	// Read nametable by setting PPUADDR and reading PPUDATA
	// Nametable 0 is at $2000-$23FF
	fmt.Printf("\nFirst 32 bytes of Nametable 0 (tile IDs):\n")

	// Set address to $2000
	ppuUnit.WriteCPURegister(0x2006, 0x20)
	ppuUnit.WriteCPURegister(0x2006, 0x00)

	// Read 32 bytes
	fmt.Print("  ")
	uniqueTiles := make(map[uint8]bool)
	for i := 0; i < 32; i++ {
		value := ppuUnit.ReadCPURegister(0x2007)
		fmt.Printf("%02X ", value)
		uniqueTiles[value] = true
		if (i+1)%16 == 0 {
			fmt.Print("\n  ")
		}
	}
	fmt.Printf("\n  Unique tiles in first 32 bytes: %d\n", len(uniqueTiles))

	// Check attribute table (starts at $23C0)
	fmt.Printf("\nFirst 16 bytes of Attribute Table 0:\n")
	ppuUnit.WriteCPURegister(0x2006, 0x23)
	ppuUnit.WriteCPURegister(0x2006, 0xC0)

	fmt.Print("  ")
	for i := 0; i < 16; i++ {
		value := ppuUnit.ReadCPURegister(0x2007)
		fmt.Printf("%02X ", value)
		if (i+1)%8 == 0 {
			fmt.Print("\n  ")
		}
	}

	// Check palette RAM
	fmt.Printf("\nBackground Palettes ($3F00-$3F0F):\n")
	ppuUnit.WriteCPURegister(0x2006, 0x3F)
	ppuUnit.WriteCPURegister(0x2006, 0x00)

	for pal := 0; pal < 4; pal++ {
		fmt.Printf("  Palette %d: ", pal)
		for i := 0; i < 4; i++ {
			value := ppuUnit.ReadCPURegister(0x2007)
			fmt.Printf("$%02X ", value)
		}
		fmt.Println()
	}

	fmt.Printf("\nSprite Palettes ($3F10-$3F1F):\n")
	ppuUnit.WriteCPURegister(0x2006, 0x3F)
	ppuUnit.WriteCPURegister(0x2006, 0x10)

	for pal := 0; pal < 4; pal++ {
		fmt.Printf("  Palette %d: ", pal+4)
		for i := 0; i < 4; i++ {
			value := ppuUnit.ReadCPURegister(0x2007)
			fmt.Printf("$%02X ", value)
		}
		fmt.Println()
	}

	// Sample frame buffer
	frameBuffer := emulator.GetFrameBuffer()
	fmt.Printf("\nFrame buffer sample (row 10, columns 0-31):\n  ")
	for x := 0; x < 32; x++ {
		fmt.Printf("%02X ", frameBuffer[10*256+x])
	}
	fmt.Println()

	fmt.Printf("\nFrame buffer sample (row 100, columns 100-131):\n  ")
	for x := 100; x < 132; x++ {
		fmt.Printf("%02X ", frameBuffer[100*256+x])
	}
	fmt.Println()
}
