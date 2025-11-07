package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
	"github.com/andrewthecodertx/nes-emulator/pkg/ppu"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: inspect-ppu <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	// Load ROM
	fmt.Printf("=== Inspecting PPU State: %s ===\n\n", romPath)
	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Failed to load ROM: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()

	// Run for 2 seconds to let game initialize
	fmt.Println("Running 120 frames (2 seconds) to let game initialize...")
	for i := 0; i < 120; i++ {
		emulator.RunFrame()
	}

	_ = emulator.GetPPU() // Get PPU reference (for future use)

	// We need to add methods to expose PPU internal state
	// For now, let's check what we can see from the frame buffer

	frameBuffer := emulator.GetFrameBuffer()

	fmt.Println("\n=== Frame Buffer Analysis ===")

	// Sample different parts of the screen
	regions := []struct {
		name string
		y    int
		x    int
	}{
		{"Top-left corner", 0, 0},
		{"Top-center", 0, 128},
		{"Top-right corner", 0, 240},
		{"Middle-left", 120, 0},
		{"Center", 120, 128},
		{"Middle-right", 120, 240},
		{"Bottom-left", 200, 0},
		{"Bottom-center", 200, 128},
		{"Bottom-right", 200, 240},
	}

	for _, region := range regions {
		fmt.Printf("\n%s (Y=%d, X=%d):\n  ", region.name, region.y, region.x)
		for dx := 0; dx < 16 && region.x+dx < 256; dx++ {
			idx := frameBuffer[region.y*256+region.x+dx]
			fmt.Printf("%02X ", idx)
		}
	}

	// Check if we have the pattern table data in CHR-ROM
	fmt.Println("\n\n=== CHR-ROM Check ===")
	cart := emulator.GetCartridge()
	mapper := cart.GetMapper()

	// Sample pattern table 0
	fmt.Println("Pattern Table 0 (first 32 bytes):")
	fmt.Print("  ")
	for addr := uint16(0x0000); addr < 0x0020; addr++ {
		fmt.Printf("%02X ", mapper.ReadCHR(addr))
	}
	fmt.Println()

	// Sample pattern table 1
	fmt.Println("Pattern Table 1 (first 32 bytes):")
	fmt.Print("  ")
	for addr := uint16(0x1000); addr < 0x1020; addr++ {
		fmt.Printf("%02X ", mapper.ReadCHR(addr))
	}
	fmt.Println()

	// Check nametable
	fmt.Println("\n=== Nametable Check ===")
	fmt.Println("Reading nametable directly from PPU memory...")
	// We need to add a method to read nametable
	// For now, show what we know

	fmt.Println("\n=== Palette RAM Check ===")
	// We need to expose palette RAM reading
	// Try to infer from frame buffer colors

	colorUsage := make(map[uint8]int)
	for _, color := range frameBuffer {
		colorUsage[color]++
	}

	fmt.Printf("\nColors used in frame (palette indices):\n")
	uniqueColors := []uint8{}
	for color := range colorUsage {
		uniqueColors = append(uniqueColors, color)
	}

	// Simple sort
	for i := 0; i < len(uniqueColors); i++ {
		for j := i + 1; j < len(uniqueColors); j++ {
			if uniqueColors[j] < uniqueColors[i] {
				uniqueColors[i], uniqueColors[j] = uniqueColors[j], uniqueColors[i]
			}
		}
	}

	for _, color := range uniqueColors {
		percentage := float64(colorUsage[color]) * 100.0 / float64(len(frameBuffer))
		rgb := ppu.HardwarePalette[color]
		fmt.Printf("  $%02X: RGB(%3d,%3d,%3d) - %6d pixels (%.1f%%)\n",
			color, rgb.R, rgb.G, rgb.B, colorUsage[color], percentage)
	}

	// Check for specific problem patterns
	fmt.Println("\n=== Problem Detection ===")

	// Check if all one color
	if len(colorUsage) == 1 {
		fmt.Println("⚠️  Entire screen is one color - rendering likely disabled")
	} else if len(colorUsage) < 3 {
		fmt.Println("⚠️  Very few colors - partial rendering or solid background")
	}

	// Check for repeated patterns that might indicate tile corruption
	// Sample a horizontal line and look for repeating 8-pixel patterns
	scanline := 100
	fmt.Printf("\nScanline %d pattern check:\n  ", scanline)

	patternCounts := make(map[string]int)
	for x := 0; x < 256; x += 8 {
		pattern := ""
		for dx := 0; dx < 8; dx++ {
			pattern += fmt.Sprintf("%02X", frameBuffer[scanline*256+x+dx])
		}
		patternCounts[pattern]++
	}

	if len(patternCounts) == 1 {
		fmt.Println("All 8-pixel tiles are identical - possible problem!")
	} else if len(patternCounts) < 5 {
		fmt.Printf("Only %d unique tile patterns detected (might be OK for some screens)\n", len(patternCounts))
	} else {
		fmt.Printf("Found %d different tile patterns - looks diverse\n", len(patternCounts))
	}

	fmt.Println("\n=== Inspection Complete ===")
	fmt.Println("\nTo see the actual display, run: ./nes-sdl", romPath)
}
