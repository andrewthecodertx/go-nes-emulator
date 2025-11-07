package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: debug-shifters <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()
	ppuUnit := emulator.GetPPU()

	// Run 120 init frames
	for i := 0; i < 120; i++ {
		emulator.RunFrame()
	}

	// Check PPUMASK
	ppuMask := ppuUnit.ReadCPURegister(0x2001)
	fmt.Printf("PPUMASK: $%02X\n", ppuMask)
	fmt.Printf("Rendering enabled: %v\n", (ppuMask&0x18) != 0)

	// Run one more frame and try to inspect shift register state
	// We can't directly access shift registers, but we can check
	// what the frame buffer contains
	emulator.RunFrame()

	frameBuffer := emulator.GetFrameBuffer()

	// Sample middle of screen
	fmt.Println("\nSample pixels from scanline 120, x=100-131:")
	for x := 100; x < 132; x++ {
		idx := frameBuffer[120*256+x]
		fmt.Printf("%02X ", idx)
		if (x-100+1)%16 == 0 {
			fmt.Println()
		}
	}

	// Count colors
	colorCounts := make(map[uint8]int)
	for _, c := range frameBuffer {
		colorCounts[c]++
	}

	fmt.Printf("\nUnique palette indices in frame: %d\n", len(colorCounts))
	fmt.Println("Color distribution (top 10):")

	// Sort by count
	type kv struct {
		idx   uint8
		count int
	}
	var sorted []kv
	for k, v := range colorCounts {
		sorted = append(sorted, kv{k, v})
	}

	// Simple bubble sort
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].count > sorted[i].count {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	for i := 0; i < 10 && i < len(sorted); i++ {
		pct := float64(sorted[i].count) * 100.0 / float64(256*240)
		fmt.Printf("  $%02X: %6d pixels (%.1f%%)\n", sorted[i].idx, sorted[i].count, pct)
	}
}
