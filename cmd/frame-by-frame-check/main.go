package main

import (
	"fmt"
	"os"
	"time"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: frame-by-frame-check <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()

	fmt.Println("Running frame-by-frame, checking frame buffer after each frame")
	fmt.Println("Similar to how SDL reads it")
	fmt.Println()

	for frame := 0; frame < 305; frame++ {
		emulator.RunFrame()

		// Add a tiny delay to simulate real-time rendering
		time.Sleep(1 * time.Millisecond)

		if frame == 300 {
			frameBuffer := emulator.GetFrameBuffer()

			fmt.Printf("=== FRAME %d (frame-by-frame mode) ===\n", frame)
			fmt.Printf("Sampling scanline 100, x=64-79:\n")
			for x := 64; x < 80; x++ {
				idx := frameBuffer[100*256+x]
				fmt.Printf("$%02X ", idx)
			}
			fmt.Println()

			// Count $15 and $25
			count15 := 0
			count25 := 0
			for _, idx := range frameBuffer {
				if idx == 0x15 {
					count15++
				}
				if idx == 0x25 {
					count25++
				}
			}

			fmt.Printf("\n$15 (dark magenta) pixels: %d\n", count15)
			fmt.Printf("$25 (bright magenta) pixels: %d\n", count25)

			if count15 > 1000 {
				fmt.Println("\n⚠️ LOTS OF $15 (dark magenta) detected!")
			}
		}
	}
}
