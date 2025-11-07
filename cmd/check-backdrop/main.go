package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: check-backdrop <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()
	ppuUnit := emulator.GetPPU()

	// Run initialization frames
	for i := 0; i < 120; i++ {
		emulator.RunFrame()
	}

	// Try to read $3F00 (backdrop color)
	// We can't directly read PPU memory from here, but we can
	// check what color appears most in the frame buffer when rendering is disabled

	frameBuffer := emulator.GetFrameBuffer()

	colorCounts := make(map[uint8]int)
	for _, c := range frameBuffer {
		colorCounts[c]++
	}

	fmt.Printf("PPUMASK: $%02X\n", ppuUnit.ReadCPURegister(0x2001))
	fmt.Printf("Rendering enabled: %v\n\n", (ppuUnit.ReadCPURegister(0x2001)&0x18) != 0)

	fmt.Printf("Color distribution:\n")
	for idx, count := range colorCounts {
		pct := float64(count) * 100.0 / float64(256*240)
		fmt.Printf("  $%02X: %6d pixels (%.1f%%)\n", idx, count, pct)
	}

	fmt.Printf("\nTotal unique colors: %d\n", len(colorCounts))
	fmt.Println("\nIf rendering is disabled, there should be only 1 color (backdrop).")
}
