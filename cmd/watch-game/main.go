package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: watch-game <rom-file>")
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
	cpu := emulator.GetCPU()
	bus := emulator.GetBus()

	fmt.Println("\nWatching CPU and PPU state over time...")
	fmt.Println("Frame | PC     | A  | X  | Y  | PPUSTATUS | Unique Colors")
	fmt.Println("------|--------|----|----|----|-----------|--------------")

	for frame := 0; frame < 600; frame++ {
		emulator.RunFrame()

		if frame%30 == 0 || frame < 10 {
			// Read PPU status
			ppuStatus := bus.Read(0x2002)

			// Count unique palette indices
			frameBuffer := emulator.GetFrameBuffer()
			paletteCounts := make(map[uint8]bool)
			for _, idx := range frameBuffer {
				paletteCounts[idx] = true
			}

			fmt.Printf("%5d | $%04X | $%02X | $%02X | $%02X | $%02X        | %d\n",
				frame, cpu.PC, cpu.A, cpu.X, cpu.Y, ppuStatus, len(paletteCounts))
		}
	}

	fmt.Println("\nTest: Are graphics changing over time?")

	// Take snapshots at different times
	emulator.Reset()

	snapshots := []int{60, 120, 180, 300, 600}
	for _, targetFrame := range snapshots {
		for frame := 0; frame < targetFrame; frame++ {
			emulator.RunFrame()
		}

		frameBuffer := emulator.GetFrameBuffer()

		// Count pixel values in center region
		centerSum := 0
		for y := 100; y < 140; y++ {
			for x := 100; x < 156; x++ {
				centerSum += int(frameBuffer[y*256+x])
			}
		}

		fmt.Printf("Frame %3d: Center region sum = %d\n", targetFrame, centerSum)
	}
}
