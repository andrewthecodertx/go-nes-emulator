package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: check-early-scanlines <rom-file>")
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

	// Check scanlines 0-176 (before rendering enabled at scanline 177)
	fmt.Println("Scanlines 0-10 (should be backdrop if rendering disabled):")
	for y := 0; y < 10; y++ {
		fmt.Printf("Scanline %3d: ", y)
		colorCounts := make(map[uint8]int)
		for x := 0; x < 256; x++ {
			idx := frameBuffer[y*256+x]
			colorCounts[idx]++
		}

		// Print color distribution
		for color, count := range colorCounts {
			if count > 10 {
				fmt.Printf("$%02X:%d ", color, count)
			}
		}
		fmt.Println()
	}

	fmt.Println("\nScanlines 177-187 (after rendering enabled):")
	for y := 177; y < 187; y++ {
		fmt.Printf("Scanline %3d: ", y)
		colorCounts := make(map[uint8]int)
		for x := 0; x < 256; x++ {
			idx := frameBuffer[y*256+x]
			colorCounts[idx]++
		}

		for color, count := range colorCounts {
			if count > 10 {
				fmt.Printf("$%02X:%d ", color, count)
			}
		}
		fmt.Println()
	}
}
