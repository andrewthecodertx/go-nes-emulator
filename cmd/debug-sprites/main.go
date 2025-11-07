package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: debug-sprites <rom-file>")
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

	// Sample a specific 8x8 region that should be a solid color tile
	// Let's check around y=100, x=64 (middle-ish area)
	fmt.Println("8x8 tile starting at (64, 100):")
	fmt.Println("(If checkerboarded, will show alternating palette indices)")
	fmt.Println()

	for y := 100; y < 108; y++ {
		fmt.Printf("Row %d: ", y)
		for x := 64; x < 72; x++ {
			idx := frameBuffer[y*256+x]
			fmt.Printf("%02X ", idx)
		}
		fmt.Println()
	}

	// Check if alternating
	fmt.Println("\nChecking for alternation pattern:")
	alternateX := 0
	alternateY := 0

	for y := 100; y < 108; y++ {
		for x := 64; x < 71; x++ {
			curr := frameBuffer[y*256+x]
			right := frameBuffer[y*256+x+1]
			if curr != right {
				alternateX++
			}
		}
	}

	for y := 100; y < 107; y++ {
		for x := 64; x < 72; x++ {
			curr := frameBuffer[y*256+x]
			below := frameBuffer[(y+1)*256+x]
			if curr != below {
				alternateY++
			}
		}
	}

	fmt.Printf("Horizontal transitions: %d/56\n", alternateX)
	fmt.Printf("Vertical transitions: %d/56\n", alternateY)

	if alternateX > 28 {
		fmt.Println("\n⚠ Horizontal checkerboard detected!")
	}
	if alternateY > 28 {
		fmt.Println("\n⚠ Vertical checkerboard detected!")
	}
}
