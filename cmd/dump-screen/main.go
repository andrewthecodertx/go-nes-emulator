package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dump-screen <rom-file>")
		os.Exit(1)
	}

	emulator, _ := nes.New(os.Args[1])
	emulator.Reset()

	// Run for 120 frames
	fmt.Println("Running game for 2 seconds...")
	for i := 0; i < 120; i++ {
		emulator.RunFrame()
	}

	frameBuffer := emulator.GetFrameBuffer()

	// Dump screen as ASCII art with color codes
	fmt.Println("\n=== Screen Dump (30x30 tiles) ===")
	fmt.Println("Each character represents one 8x8 tile")
	fmt.Println()

	for ty := 0; ty < 30; ty++ {
		for tx := 0; tx < 32; tx++ {
			// Sample the top-left pixel of each tile
			y := ty * 8
			x := tx * 8
			if x >= 256 || y >= 240 {
				continue
			}

			pixel := frameBuffer[y*256+x]

			// Convert to character based on brightness
			char := ' '
			if pixel == 0x0F || pixel == 0x00 || pixel == 0x0D {
				char = ' ' // Black/dark
			} else if pixel == 0x30 || pixel == 0x20 || pixel == 0x10 {
				char = '#' // White/bright
			} else {
				char = '.' // Mid-tone
			}

			fmt.Printf("%c", char)
		}
		fmt.Println()
	}

	fmt.Println("\n=== Color Distribution ===")
	colors := make(map[uint8]int)
	for _, c := range frameBuffer {
		colors[c]++
	}

	for c, count := range colors {
		pct := float64(count) * 100.0 / float64(len(frameBuffer))
		if pct > 1.0 {
			fmt.Printf("$%02X: %5.1f%%\n", c, pct)
		}
	}
}
