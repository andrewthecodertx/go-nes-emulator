package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/go-nes-emulator/pkg/nes"
	"github.com/andrewthecodertx/go-nes-emulator/pkg/ppu"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: detailed-render <rom-file> [frames]")
		os.Exit(1)
	}

	romPath := os.Args[1]
	frames := 300
	if len(os.Args) > 2 {
		fmt.Sscanf(os.Args[2], "%d", &frames)
	}

	fmt.Printf("Loading %s...\n", romPath)
	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()

	// Run frames
	fmt.Printf("Running %d frames...\n", frames)
	for i := 0; i < frames; i++ {
		emulator.RunFrame()
	}

	// Get frame buffer
	frameBuffer := emulator.GetFrameBuffer()

	// Show several scanlines in detail
	fmt.Println("\nDetailed scanline view (showing actual palette indices):")
	fmt.Println()

	scanlines := []int{0, 30, 60, 90, 120}
	for _, scanline := range scanlines {
		fmt.Printf("Scanline %3d (first 64 pixels):\n  ", scanline)
		for x := 0; x < 64; x++ {
			idx := frameBuffer[scanline*256+x]
			fmt.Printf("%02X ", idx)
			if (x+1)%16 == 0 {
				fmt.Print("\n  ")
			}
		}
		fmt.Println()
	}

	// Show what colors these palette indices represent
	fmt.Println("\nColor mapping (palette index -> RGB):")
	paletteCounts := make(map[uint8]int)
	for _, idx := range frameBuffer {
		paletteCounts[idx]++
	}

	type kv struct {
		k uint8
		v int
	}
	var sorted []kv
	for k, v := range paletteCounts {
		sorted = append(sorted, kv{k, v})
	}
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].v > sorted[i].v {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	fmt.Println("Index | RGB Color       | Usage")
	fmt.Println("------|-----------------|-------")
	for _, item := range sorted {
		if item.k < uint8(len(ppu.HardwarePalette)) {
			color := ppu.HardwarePalette[item.k]
			pct := float64(item.v) * 100.0 / float64(256*240)
			fmt.Printf("$%02X   | #%02X%02X%02X (%3d,%3d,%3d) | %5.1f%%\n",
				item.k, color.R, color.G, color.B, color.R, color.G, color.B, pct)
		}
	}

	// Check if this looks like valid graphics
	fmt.Println("\nAnalysis:")
	if len(paletteCounts) > 4 {
		fmt.Printf("✓ Using %d different colors (varied palette)\n", len(paletteCounts))
	} else {
		fmt.Printf("⚠ Only using %d colors (might be blank or simple)\n", len(paletteCounts))
	}

	// Check for patterns
	blackCount := paletteCounts[0x0F] + paletteCounts[0x1D] + paletteCounts[0x2D] + paletteCounts[0x3D]
	if blackCount > 256*240*8/10 {
		fmt.Println("⚠ Screen is mostly black")
	}

	totalNonZero := 0
	for idx, count := range paletteCounts {
		if idx != 0x0F { // 0x0F is black
			totalNonZero += count
		}
	}
	if totalNonZero > 256*240/10 {
		fmt.Printf("✓ %d pixels are non-black (graphics visible)\n", totalNonZero)
	}
}
