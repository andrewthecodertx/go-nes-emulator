package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: check-frame-300 <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()

	// Run 300 frames like SDL does
	for i := 0; i < 300; i++ {
		emulator.RunFrame()
	}

	frameBuffer := emulator.GetFrameBuffer()

	// Count colors
	colorCounts := make(map[uint8]int)
	for _, c := range frameBuffer {
		colorCounts[c]++
	}

	fmt.Printf("Frame 300 color distribution:\n")

	// Sort by count
	type kv struct {
		idx   uint8
		count int
	}
	var sorted []kv
	for k, v := range colorCounts {
		sorted = append(sorted, kv{k, v})
	}

	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	for i := 0; i < len(sorted); i++ {
		pct := float64(sorted[i].count) * 100.0 / float64(256*240)
		fmt.Printf("  $%02X: %6d pixels (%.1f%%)\n", sorted[i].idx, sorted[i].count, pct)
	}

	magentaCount := colorCounts[0x25]
	cyanCount := colorCounts[0x2C]

	fmt.Printf("\nMagenta ($25): %d pixels (%.1f%%)\n", magentaCount, float64(magentaCount)*100.0/float64(256*240))
	fmt.Printf("Cyan ($2C): %d pixels (%.1f%%)\n", cyanCount, float64(cyanCount)*100.0/float64(256*240))

	if magentaCount > 10000 {
		fmt.Println("\n⚠ LOTS OF MAGENTA - This matches the screenshot!")
	}
}
