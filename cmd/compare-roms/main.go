package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: compare-roms <rom1> [rom2]")
		os.Exit(1)
	}

	// Load and run first ROM
	fmt.Printf("Loading %s...\n", os.Args[1])
	emulator1, err := nes.New(os.Args[1])
	if err != nil {
		fmt.Printf("Error loading ROM: %v\n", err)
		os.Exit(1)
	}

	emulator1.Reset()

	// Enable rendering
	ppu1 := emulator1.GetPPU()

	// Run for several frames to let the game initialize
	fmt.Println("Running for 5 frames...")
	for i := 0; i < 5; i++ {
		emulator1.RunFrame()
	}

	fmt.Printf("After 5 frames:\n")

	// Get PPU state
	fmt.Printf("  Scanline: %d\n", ppu1.GetScanline())
	fmt.Printf("  Cycle: %d\n", ppu1.GetCycle())

	// Sample some palette values
	fmt.Println("\nBackground Palette 0 (first 4 colors):")
	for i := 0; i < 4; i++ {
		val := ppu1.ReadPaletteRAM(uint16(i))
		fmt.Printf("  Palette[%d] = 0x%02X\n", i, val)
	}

	// Sample some nametable values
	fmt.Println("\nNametable sample (first 16 bytes at $2000):")
	for i := uint16(0); i < 16; i++ {
		// Read directly from nametable
		fmt.Printf("  [%04X] = 0x%02X\n", 0x2000+i, ppu1.ReadNametable(i))
	}

	// Check frame buffer
	frameBuffer := emulator1.GetFrameBuffer()

	// Sample some pixels from the frame buffer
	fmt.Println("\nFrame buffer samples:")
	fmt.Printf("  Top-left corner (0,0): palette index %d\n", frameBuffer[0])
	fmt.Printf("  (0,1): %d\n", frameBuffer[1])
	fmt.Printf("  (1,0): %d\n", frameBuffer[256])
	fmt.Printf("  Center (128,120): %d\n", frameBuffer[120*256+128])

	// Count unique palette indices used
	uniquePalettes := make(map[uint8]int)
	for _, val := range frameBuffer {
		uniquePalettes[val]++
	}

	fmt.Printf("\nUnique palette indices in frame: %d\n", len(uniquePalettes))
	fmt.Println("Top 10 most used palette indices:")

	// Find top 10
	type kv struct {
		Key   uint8
		Value int
	}
	var sorted []kv
	for k, v := range uniquePalettes {
		sorted = append(sorted, kv{k, v})
	}

	// Simple bubble sort for top 10
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Value > sorted[i].Value {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	for i := 0; i < 10 && i < len(sorted); i++ {
		fmt.Printf("  Index %d: used %d times (%.1f%%)\n",
			sorted[i].Key, sorted[i].Value,
			float64(sorted[i].Value)*100.0/float64(256*240))
	}

	// Check if rendering is enabled
	fmt.Println("\nPPU Control register:")
	fmt.Printf("  PPUCTRL value: 0x%02X\n", ppu1.GetPPUCTRL())
	fmt.Printf("  PPUMASK value: 0x%02X\n", ppu1.GetPPUMASK())
}
