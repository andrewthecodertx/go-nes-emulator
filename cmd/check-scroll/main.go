package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: check-scroll <rom-file>")
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

	// We can't directly read scroll registers, but we can infer from the pattern
	// A checkerboard suggests fine X scroll might be odd instead of even

	frameBuffer := emulator.GetFrameBuffer()

	fmt.Println("Checking for checkerboard pattern on scanline 100:")
	fmt.Println("(A checkerboard would show alternating colors)")
	fmt.Println()

	for x := 0; x < 32; x++ {
		idx := frameBuffer[100*256+x]
		fmt.Printf("%02X ", idx)
	}
	fmt.Println()

	// Check if pixels alternate
	alternates := 0
	same := 0
	for x := 1; x < 256; x++ {
		curr := frameBuffer[100*256+x]
		prev := frameBuffer[100*256+x-1]
		if curr != prev {
			alternates++
		} else {
			same++
		}
	}

	fmt.Printf("\nPixels that differ from previous: %d\n", alternates)
	fmt.Printf("Pixels that match previous: %d\n", same)

	if alternates > 200 {
		fmt.Println("\n⚠ High alternation detected - suggests checkerboard pattern!")
	}
}
