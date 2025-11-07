package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: check-frame-1 <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()

	// Check frame 1 (right after reset, before any RunFrame calls)
	fmt.Println("Frame 0 (immediately after reset):")
	checkFrame(emulator)

	// Run 1 frame
	emulator.RunFrame()
	fmt.Println("\nFrame 1:")
	checkFrame(emulator)

	// Run to frame 120
	for i := 1; i < 120; i++ {
		emulator.RunFrame()
	}
	fmt.Println("\nFrame 120:")
	checkFrame(emulator)

	// Run to frame 300
	for i := 120; i < 300; i++ {
		emulator.RunFrame()
	}
	fmt.Println("\nFrame 300:")
	checkFrame(emulator)
}

func checkFrame(emulator *nes.NES) {
	frameBuffer := emulator.GetFrameBuffer()

	colorCounts := make(map[uint8]int)
	for _, c := range frameBuffer {
		colorCounts[c]++
	}

	fmt.Printf("  Unique colors: %d\n", len(colorCounts))

	magentaCount := colorCounts[0x25]
	cyanCount := colorCounts[0x2C]

	fmt.Printf("  Magenta ($25): %d pixels (%.1f%%)\n", magentaCount, float64(magentaCount)*100.0/float64(256*240))
	fmt.Printf("  Cyan ($2C): %d pixels (%.1f%%)\n", cyanCount, float64(cyanCount)*100.0/float64(256*240))

	if magentaCount > 10000 {
		fmt.Println("  ⚠ LOTS OF MAGENTA!")
	}
}
