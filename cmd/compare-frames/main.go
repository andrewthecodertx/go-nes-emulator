package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
	"github.com/andrewthecodertx/nes-emulator/pkg/ppu"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: compare-frames <rom-file>")
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

	// Save frame 300
	frame300 := make([]uint8, ppu.ScreenWidth*ppu.ScreenHeight)
	copy(frame300, emulator.GetFrameBuffer()[:])

	// Run one more frame
	emulator.RunFrame()

	// Compare frame 301
	frame301 := emulator.GetFrameBuffer()

	differences := 0
	for i := 0; i < len(frame300); i++ {
		if frame300[i] != frame301[i] {
			differences++
		}
	}

	fmt.Printf("Differences between frame 300 and 301: %d pixels (%.1f%%)\n",
		differences, float64(differences)*100.0/float64(len(frame300)))

	if differences > 30000 {
		fmt.Println("\n⚠ Significant differences - frames alternate!")
		fmt.Println("Sample differences on scanline 100:")
		for x := 0; x < 32; x++ {
			idx := 100*256 + x
			if frame300[idx] != frame301[idx] {
				fmt.Printf("x=%2d: $%02X -> $%02X\n", x, frame300[idx], frame301[idx])
			}
		}
	} else if differences > 0 {
		fmt.Println("\nMinor differences (sprites/animation)")
	} else {
		fmt.Println("\nFrames are identical (game paused or static screen)")
	}
}
