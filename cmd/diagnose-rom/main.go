package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: diagnose-rom <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	fmt.Printf("Loading %s...\n", romPath)
	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Error loading ROM: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()
	cpu := emulator.GetCPU()

	fmt.Printf("\nInitial CPU state:\n")
	fmt.Printf("  PC: $%04X\n", cpu.PC)
	fmt.Printf("  A: $%02X\n", cpu.A)
	fmt.Printf("  X: $%02X\n", cpu.X)
	fmt.Printf("  Y: $%02X\n", cpu.Y)
	fmt.Printf("  SP: $%02X\n", cpu.SP)

	// Run for 10 frames to let the game initialize
	fmt.Println("\nRunning for 10 frames...")
	for i := 0; i < 10; i++ {
		emulator.RunFrame()
		if i == 0 || i == 4 || i == 9 {
			fmt.Printf("  Frame %d - PC: $%04X\n", i+1, cpu.PC)
		}
	}

	frameBuffer := emulator.GetFrameBuffer()

	// Analyze frame buffer
	fmt.Println("\nFrame buffer analysis:")

	// Count occurrences of each palette index
	paletteCounts := make(map[uint8]int)
	for _, idx := range frameBuffer {
		paletteCounts[idx]++
	}

	fmt.Printf("  Unique palette indices used: %d\n", len(paletteCounts))

	// Check if it's all one color (blank screen)
	if len(paletteCounts) == 1 {
		for idx := range paletteCounts {
			fmt.Printf("  WARNING: Entire screen is palette index %d (blank screen!)\n", idx)
		}
	}

	// Check for common issues
	zeroCount := paletteCounts[0]
	totalPixels := 256 * 240
	if zeroCount == totalPixels {
		fmt.Println("  WARNING: All pixels are palette index 0 - rendering may be disabled")
	} else if zeroCount > totalPixels*9/10 {
		fmt.Printf("  NOTE: %.1f%% of pixels are palette index 0 (background)\n",
			float64(zeroCount)*100.0/float64(totalPixels))
	}

	// Sample a few scanlines
	fmt.Println("\nSample scanline data (scanline 100, first 32 pixels):")
	scanline := 100
	fmt.Print("  ")
	for x := 0; x < 32; x++ {
		idx := frameBuffer[scanline*256+x]
		fmt.Printf("%X", idx&0x0F)
	}
	fmt.Println()

	// Check CPU memory for PPU register writes
	fmt.Println("\nCurrent CPU state:")
	fmt.Printf("  PC: $%04X\n", cpu.PC)
	fmt.Printf("  Total cycles: %d\n", emulator.GetCycles())

	// Try to read current PPUCTRL and PPUMASK values by triggering writes
	bus := emulator.GetBus()
	ppuStatus := bus.Read(0x2002)
	fmt.Printf("\nPPU Status ($2002): $%02X\n", ppuStatus)
	fmt.Printf("  VBlank: %v\n", (ppuStatus&0x80) != 0)
	fmt.Printf("  Sprite 0 Hit: %v\n", (ppuStatus&0x40) != 0)
	fmt.Printf("  Sprite Overflow: %v\n", (ppuStatus&0x20) != 0)

	fmt.Println("\nDiagnostics complete.")
}
