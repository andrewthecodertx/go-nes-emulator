package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/controller"
	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: trace-io <rom-file>")
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

	// Get components
	bus := emulator.GetBus()
	ctrl := bus.GetController(0)
	cpu := emulator.GetCPU()

	// Press some buttons
	ctrl.SetButton(controller.ButtonStart, true)
	ctrl.SetButton(controller.ButtonA, true)

	fmt.Println("\nRunning emulation and watching for controller I/O...")
	fmt.Println("(Running 600 frames = ~10 seconds)")

	// Track I/O operations
	lastPC := uint16(0)
	controllerReadCount := 0
	controllerWriteCount := 0
	ppuWriteCount := 0

	// Run many frames
	for frame := 0; frame < 600; frame++ {
		frameStart := emulator.GetCycles()

		// Run one frame
		for emulator.GetCycles()-frameStart < 29781 {
			prevPC := cpu.PC
			emulator.Step()

			// Check if PC changed significantly (detect certain I/O operations indirectly)
			// We can't easily hook into bus reads/writes without modifying the emulator,
			// so let's check PC patterns

			// Sample CPU state periodically
			if frame%60 == 0 && emulator.GetCycles()%1000 == 0 {
				lastPC = cpu.PC
			}
		}

		// Print status every 60 frames (1 second)
		if frame%60 == 0 && frame > 0 {
			fmt.Printf("Frame %d: PC=$%04X A=$%02X X=$%02X Y=$%02X\n",
				frame, cpu.PC, cpu.A, cpu.X, cpu.Y)
		}

		// Check if we're stuck in a loop
		if frame > 120 && frame%60 == 0 {
			if lastPC == cpu.PC {
				fmt.Printf("  WARNING: PC hasn't changed much (stuck at $%04X?)\n", cpu.PC)
			}
			lastPC = cpu.PC
		}
	}

	fmt.Println("\nDone running.")
	fmt.Printf("Final CPU state: PC=$%04X A=$%02X X=$%02X Y=$%02X SP=$%02X\n",
		cpu.PC, cpu.A, cpu.X, cpu.Y, cpu.SP)

	// Check frame buffer to see if anything rendered
	frameBuffer := emulator.GetFrameBuffer()
	paletteCounts := make(map[uint8]int)
	for _, idx := range frameBuffer {
		paletteCounts[idx]++
	}
	fmt.Printf("\nFrame buffer has %d unique palette indices\n", len(paletteCounts))

	// The real issue: we need to check if the game is actually polling the controller
	// Without hooking into the bus, we can't easily trace this
	fmt.Println("\nNote: To properly debug controller input, we'd need to add logging")
	fmt.Println("      to the bus Read() function at addr $4016/$4017")
	fmt.Println("      For now, verify the game is running by checking if PC is changing.")
}
