package main

import (
	"fmt"
	"log"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
	"github.com/andrewthecodertx/nes-emulator/pkg/ppu"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: debug-frame <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	// Load NES ROM
	fmt.Printf("Loading ROM: %s\n", romPath)
	emulator, err := nes.New(romPath)
	if err != nil {
		log.Fatalf("Failed to load ROM: %v", err)
	}

	// Reset NES
	emulator.Reset()
	ppuUnit := emulator.GetPPU()

	// Check initial CPU state
	cpu := emulator.GetCPU()
	fmt.Printf("\nInitial CPU state:\n")
	fmt.Printf("  PC: $%04X\n", cpu.PC)
	fmt.Printf("  A: $%02X  X: $%02X  Y: $%02X\n", cpu.A, cpu.X, cpu.Y)
	fmt.Printf("  SP: $%02X  Status: $%02X\n", cpu.SP, cpu.Status)

	// Enable rendering
	ppuUnit.WriteCPURegister(0x2001, 0x1E) // Enable background + sprites, show left 8 pixels
	fmt.Printf("\nRendering enabled (PPUMASK = $1E)\n")

	// Check PPU state
	fmt.Printf("\nPPU Control: $%02X\n", ppuUnit.ReadCPURegister(0x2000))
	fmt.Printf("PPU Mask: $%02X\n", ppuUnit.ReadCPURegister(0x2001))

	// Run several frames
	fmt.Println("\nRunning 60 frames...")
	for i := 0; i < 60; i++ {
		emulator.RunFrame()
	}

	// Analyze frame buffer
	frameBuffer := emulator.GetFrameBuffer()

	fmt.Println("\nFrame buffer analysis:")

	// Count unique colors
	colorCount := make(map[uint8]int)
	for _, pixel := range frameBuffer {
		colorCount[pixel]++
	}

	fmt.Printf("  Unique palette indices: %d\n", len(colorCount))
	fmt.Printf("  Non-zero pixels: %d / %d\n",
		256*240-colorCount[0], 256*240)

	// Show most common colors
	fmt.Println("\n  Top 10 palette indices:")
	type colorFreq struct {
		idx   uint8
		count int
	}
	var colors []colorFreq
	for idx, count := range colorCount {
		colors = append(colors, colorFreq{idx, count})
	}
	// Simple bubble sort
	for i := 0; i < len(colors); i++ {
		for j := i + 1; j < len(colors); j++ {
			if colors[j].count > colors[i].count {
				colors[i], colors[j] = colors[j], colors[i]
			}
		}
	}
	for i := 0; i < 10 && i < len(colors); i++ {
		idx := colors[i].idx
		rgb := ppu.HardwarePalette[idx]
		fmt.Printf("    $%02X: %6d pixels  RGB(%3d,%3d,%3d)\n",
			idx, colors[i].count, rgb.R, rgb.G, rgb.B)
	}

	// Sample some pixels
	fmt.Println("\n  Sample pixels (top-left 16x16):")
	for y := 0; y < 16; y++ {
		fmt.Print("    ")
		for x := 0; x < 16; x++ {
			idx := frameBuffer[y*256+x]
			fmt.Printf("%02X ", idx)
		}
		fmt.Println()
	}

	// Check for patterns
	fmt.Println("\n  Visual representation (top 40 rows, every other pixel):")
	chars := " .:+=*#@"
	for y := 0; y < 40; y++ {
		fmt.Print("    ")
		for x := 0; x < 128; x += 2 {
			idx := frameBuffer[y*256+x*2]
			brightness := int(idx) * len(chars) / 64
			if brightness >= len(chars) {
				brightness = len(chars) - 1
			}
			fmt.Print(string(chars[brightness]))
		}
		fmt.Println()
	}
}
