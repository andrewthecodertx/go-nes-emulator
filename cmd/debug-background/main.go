package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: debug-background <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()

	// Run 300 frames
	for i := 0; i < 300; i++ {
		emulator.RunFrame()
	}

	frameBuffer := emulator.GetFrameBuffer()

	// Get the PPU to check what palette RAM contains
	ppu := emulator.GetPPU()

	fmt.Println("Background should show black and cyan/orange, not magenta")
	fmt.Println()

	// Check what colors we actually have
	fmt.Println("Checking scanline 100, x=64-79 (should be solid blue $12):")
	for x := 64; x < 80; x++ {
		idx := frameBuffer[100*256+x]
		if idx == 0x25 {
			fmt.Printf("x=%d: $%02X <- MAGENTA (WRONG!)\n", x, idx)
		} else if idx == 0x2C {
			fmt.Printf("x=%d: $%02X <- Cyan\n", x, idx)
		} else if idx == 0x12 {
			fmt.Printf("x=%d: $%02X <- Blue (correct)\n", x, idx)
		} else {
			fmt.Printf("x=%d: $%02X\n", x, idx)
		}
	}

	// Check if the pattern alternates
	fmt.Println("\nChecking for alternating pattern on scanline 120:")
	prevIdx := frameBuffer[120*256+0]
	alternates := 0
	for x := 1; x < 256; x++ {
		currIdx := frameBuffer[120*256+x]
		if currIdx != prevIdx && currIdx == 0x25 || prevIdx == 0x25 {
			alternates++
		}
		prevIdx = currIdx
	}

	fmt.Printf("Magenta alternations in scanline: %d\n", alternates)

	if alternates > 100 {
		fmt.Println("\n⚠️ CHECKERBOARD DETECTED IN FRAME BUFFER!")
		fmt.Println("This means the bug is in PPU rendering, not SDL display")
	}

	_ = ppu // Use ppu to avoid unused variable error
}
