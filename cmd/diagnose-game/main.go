package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: diagnose-game <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	// Load ROM
	fmt.Printf("=== Diagnosing: %s ===\n\n", romPath)
	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Failed to load ROM: %v\n", err)
		os.Exit(1)
	}

	cart := emulator.GetCartridge()
	fmt.Printf("Mapper: %d\n", cart.GetMapperID())
	fmt.Printf("PRG Banks: %d x 16KB\n", cart.GetPRGBanks())
	fmt.Printf("CHR Banks: %d x 8KB\n\n", cart.GetCHRBanks())

	// Reset
	emulator.Reset()
	cpu := emulator.GetCPU()
	_ = emulator.GetPPU() // Get PPU but don't use yet

	fmt.Printf("Initial CPU State:\n")
	fmt.Printf("  PC: $%04X\n", cpu.PC)
	fmt.Printf("  SP: $%02X\n", cpu.SP)
	fmt.Printf("  A: $%02X  X: $%02X  Y: $%02X\n", cpu.A, cpu.X, cpu.Y)
	fmt.Printf("  Status: $%02X\n\n", cpu.Status)

	// Run for a few frames and check PPU state
	fmt.Println("Running 60 frames (1 second)...")
	for i := 0; i < 60; i++ {
		emulator.RunFrame()
	}

	fmt.Printf("\nAfter 60 frames:\n")
	fmt.Printf("  PC: $%04X\n", cpu.PC)
	fmt.Printf("  SP: $%02X\n", cpu.SP)
	fmt.Printf("  A: $%02X  X: $%02X  Y: $%02X\n", cpu.A, cpu.X, cpu.Y)
	fmt.Printf("  Status: $%02X\n", cpu.Status)
	fmt.Printf("  Total CPU Cycles: %d\n\n", emulator.GetCycles())

	// Check PPU registers
	fmt.Println("PPU State:")
	fmt.Printf("  PPUCTRL: Would need to expose this\n")
	fmt.Printf("  PPUMASK: Would need to expose this\n")
	fmt.Printf("  PPUSTATUS: Would need to expose this\n\n")

	// Check frame buffer
	frameBuffer := emulator.GetFrameBuffer()

	// Count unique colors
	colorCounts := make(map[uint8]int)
	for _, color := range frameBuffer {
		colorCounts[color]++
	}

	fmt.Printf("Frame Buffer Analysis:\n")
	fmt.Printf("  Unique colors: %d\n", len(colorCounts))

	// Find most common colors
	type colorCount struct {
		color uint8
		count int
	}
	var counts []colorCount
	for color, count := range colorCounts {
		counts = append(counts, colorCount{color, count})
	}

	// Simple bubble sort to get top 5
	for i := 0; i < len(counts); i++ {
		for j := i + 1; j < len(counts); j++ {
			if counts[j].count > counts[i].count {
				counts[i], counts[j] = counts[j], counts[i]
			}
		}
	}

	fmt.Println("  Top 5 colors:")
	for i := 0; i < 5 && i < len(counts); i++ {
		percentage := float64(counts[i].count) * 100.0 / float64(len(frameBuffer))
		fmt.Printf("    $%02X: %6d pixels (%.1f%%)\n", counts[i].color, counts[i].count, percentage)
	}

	// Sample some specific areas
	fmt.Printf("\nSample pixels (scanline 120, columns 0-15):\n  ")
	for x := 0; x < 16; x++ {
		fmt.Printf("%02X ", frameBuffer[120*256+x])
	}
	fmt.Println()

	fmt.Printf("\nSample pixels (scanline 120, columns 120-135):\n  ")
	for x := 120; x < 136; x++ {
		fmt.Printf("%02X ", frameBuffer[120*256+x])
	}
	fmt.Println()

	// Check if rendering is likely enabled
	if len(colorCounts) == 1 {
		onlyColor := uint8(0)
		for c := range colorCounts {
			onlyColor = c
		}
		fmt.Printf("\n⚠️  WARNING: Only one color ($%02X) in entire frame!\n", onlyColor)
		fmt.Println("   This suggests rendering may not be enabled.")
		fmt.Println("   The game might not have set PPUMASK yet.")
	} else if len(colorCounts) < 3 {
		fmt.Printf("\n⚠️  WARNING: Only %d colors in frame.\n", len(colorCounts))
		fmt.Println("   This may indicate limited or no rendering.")
	} else {
		fmt.Printf("\n✅ Frame has %d different colors - rendering appears active\n", len(colorCounts))
	}

	// Check for CPU stuck in a loop
	fmt.Println("\nRunning 10 more frames to check for progress...")
	oldPC := cpu.PC
	oldCycles := emulator.GetCycles()

	for i := 0; i < 10; i++ {
		emulator.RunFrame()
	}

	newPC := cpu.PC
	newCycles := emulator.GetCycles()
	cyclesPerFrame := (newCycles - oldCycles) / 10

	fmt.Printf("  PC changed: $%04X -> $%04X\n", oldPC, newPC)
	fmt.Printf("  Cycles per frame: ~%d (expected ~29781)\n", cyclesPerFrame)

	if cyclesPerFrame < 20000 {
		fmt.Println("\n⚠️  WARNING: Very few cycles per frame!")
		fmt.Println("   CPU may be stuck or running very short loops")
	} else if cyclesPerFrame > 40000 {
		fmt.Println("\n⚠️  WARNING: Too many cycles per frame!")
		fmt.Println("   Frame timing may be incorrect")
	} else {
		fmt.Println("\n✅ Cycle count looks reasonable")
	}

	fmt.Println("\n=== Diagnosis Complete ===")
}
