package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: trace-palette-writes <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	fmt.Printf("Loading %s...\n", romPath)
	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()
	bus := emulator.GetBus()

	fmt.Println("\nMonitoring palette evolution over time...")
	fmt.Println()

	checkFrames := []int{1, 5, 10, 30, 60, 120, 180, 300, 600}

	for _, targetFrame := range checkFrames {
		// Run up to target frame
		for emulator.GetCycles() < uint64(targetFrame)*29781 {
			emulator.Step()
		}

		// Read palette
		fmt.Printf("Frame %3d - Background Palette 0: ", targetFrame)
		for i := 0; i < 4; i++ {
			addr := uint16(0x3F00 + i)
			bus.Write(0x2006, uint8(addr>>8))
			bus.Write(0x2006, uint8(addr&0xFF))
			value := bus.Read(0x2007)
			fmt.Printf("$%02X ", value)
		}
		fmt.Println()
	}

	// Check when palette first gets non-zero values
	fmt.Println("\nScanning for first palette write...")
	emulator.Reset()

	lastPalette := [4]uint8{0, 0, 0, 0}

	for frame := 0; frame < 600; frame++ {
		emulator.RunFrame()

		// Check palette every frame
		currentPalette := [4]uint8{}
		for i := 0; i < 4; i++ {
			addr := uint16(0x3F00 + i)
			bus.Write(0x2006, uint8(addr>>8))
			bus.Write(0x2006, uint8(addr&0xFF))
			currentPalette[i] = bus.Read(0x2007)
		}

		// Check if palette changed
		if currentPalette != lastPalette {
			fmt.Printf("Frame %3d: Palette changed to [$%02X $%02X $%02X $%02X]\n",
				frame, currentPalette[0], currentPalette[1], currentPalette[2], currentPalette[3])
			lastPalette = currentPalette
		}

		// Stop after first few changes
		if frame > 10 && currentPalette[1] != 0 {
			break
		}
	}
}
