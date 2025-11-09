package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/go-nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ascii-render <rom-file> [frames]")
		os.Exit(1)
	}

	romPath := os.Args[1]
	frames := 120
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

	// Run several frames
	fmt.Printf("Running %d frames...\n", frames)
	for i := 0; i < frames; i++ {
		emulator.RunFrame()
	}

	// Get frame buffer
	frameBuffer := emulator.GetFrameBuffer()

	// Render a portion of the screen as ASCII
	fmt.Println("\nFrame buffer visualization (top 32x24 section):")
	fmt.Println("  (Each character represents an 8x8 region)")
	fmt.Println()

	// Characters to use for different brightness levels
	chars := " .:-=+*#%@"

	for y := 0; y < 24; y++ {
		fmt.Print("  ")
		for x := 0; x < 32; x++ {
			// Sample 8x8 block and get average palette index
			sum := 0
			for dy := 0; dy < 8; dy++ {
				for dx := 0; dx < 8; dx++ {
					px := x*8 + dx
					py := y*8 + dy
					if py < 240 && px < 256 {
						idx := frameBuffer[py*256+px]
						sum += int(idx)
					}
				}
			}
			avg := sum / 64

			// Map to character
			charIndex := avg * len(chars) / 64
			if charIndex >= len(chars) {
				charIndex = len(chars) - 1
			}
			fmt.Printf("%c", chars[charIndex])
		}
		fmt.Println()
	}

	// Show palette information
	_ = emulator.GetPPU()
	fmt.Println("\nCurrent background palettes:")
	for pal := 0; pal < 4; pal++ {
		fmt.Printf("  Palette %d: ", pal)
		for i := 0; i < 4; i++ {
			_ = uint16(pal*4 + i)
			// This will work if we expose ReadPaletteRAM, otherwise just show frame stats
			fmt.Printf("-- ")
		}
		fmt.Println()
	}

	// Count palette usage
	paletteCounts := make(map[uint8]int)
	for _, idx := range frameBuffer {
		paletteCounts[idx]++
	}

	fmt.Printf("\nPalette usage summary (%d unique indices):\n", len(paletteCounts))
	type kv struct {
		k uint8
		v int
	}
	var sorted []kv
	for k, v := range paletteCounts {
		sorted = append(sorted, kv{k, v})
	}
	// Sort
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].v > sorted[i].v {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	for i := 0; i < 5 && i < len(sorted); i++ {
		pct := float64(sorted[i].v) * 100.0 / float64(256*240)
		fmt.Printf("  Index $%02X: %6d pixels (%.1f%%)\n", sorted[i].k, sorted[i].v, pct)
	}
}
