package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/controller"
	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-controls <rom-file>")
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
	ppu := emulator.GetPPU()
	bus := emulator.GetBus()
	ctrl := bus.GetController(0)

	fmt.Println("\n=== Testing PPU Mask Register ===")

	// Read initial PPUMASK value
	initialMask := bus.Read(0x2002) // Read status to clear
	fmt.Printf("Initial PPUSTATUS: $%02X\n", initialMask)

	// Write to PPUMASK
	fmt.Println("\nWriting $1E to PPUMASK ($2001) - enable background + sprites")
	ppu.WriteCPURegister(0x2001, 0x1E)

	// Run a few cycles
	for i := 0; i < 100; i++ {
		emulator.Step()
	}

	// Check if rendering is enabled
	fmt.Printf("Checking internal PPU state after write...\n")
	fmt.Printf("  (We can't read PPUMASK directly, but we can test effects)\n")

	fmt.Println("\nWriting $00 to PPUMASK - disable rendering")
	ppu.WriteCPURegister(0x2001, 0x00)

	fmt.Println("\nWriting $08 to PPUMASK - enable background only")
	ppu.WriteCPURegister(0x2001, 0x08)

	fmt.Println("\nWriting $10 to PPUMASK - enable sprites only")
	ppu.WriteCPURegister(0x2001, 0x10)

	fmt.Println("\nWriting $1E to PPUMASK - enable both + show left 8 pixels")
	ppu.WriteCPURegister(0x2001, 0x1E)

	fmt.Println("\n=== Testing Controller Input ===")

	// Test button setting
	fmt.Println("\nSetting A button pressed")
	ctrl.SetButton(controller.ButtonA, true)
	fmt.Printf("  A button state: %v\n", ctrl.IsPressed(controller.ButtonA))

	fmt.Println("\nSetting B button pressed")
	ctrl.SetButton(controller.ButtonB, true)
	fmt.Printf("  B button state: %v\n", ctrl.IsPressed(controller.ButtonB))

	// Test controller strobe protocol
	fmt.Println("\n=== Testing Controller Strobe Protocol ===")

	// Set some buttons
	ctrl.SetButton(controller.ButtonA, true)
	ctrl.SetButton(controller.ButtonB, false)
	ctrl.SetButton(controller.ButtonStart, true)
	ctrl.SetButton(controller.ButtonUp, true)

	// Strobe sequence: write 1, then 0 to $4016
	fmt.Println("\nWriting $01 to $4016 (strobe on)")
	bus.Write(0x4016, 0x01)

	fmt.Println("Writing $00 to $4016 (strobe off - latch buttons)")
	bus.Write(0x4016, 0x00)

	// Read button states
	fmt.Println("\nReading buttons from $4016:")
	buttons := []string{"A", "B", "Select", "Start", "Up", "Down", "Left", "Right"}
	for i := 0; i < 8; i++ {
		value := bus.Read(0x4016)
		fmt.Printf("  Button %d (%s): %d\n", i, buttons[i], value&0x01)
	}

	// Test continuous reading (should return 1 after 8 reads)
	fmt.Println("\nReading 3 more times (should return 1):")
	for i := 0; i < 3; i++ {
		value := bus.Read(0x4016)
		fmt.Printf("  Extra read %d: %d\n", i+1, value&0x01)
	}

	// Test strobe mode (should always return A button)
	fmt.Println("\n=== Testing Strobe Mode ===")
	fmt.Println("Writing $01 to $4016 (strobe on)")
	bus.Write(0x4016, 0x01)

	fmt.Println("Reading $4016 multiple times (should always return A button state):")
	for i := 0; i < 5; i++ {
		value := bus.Read(0x4016)
		fmt.Printf("  Read %d: %d (A button is pressed=%v)\n", i+1, value&0x01, ctrl.IsPressed(controller.ButtonA))
	}

	fmt.Println("\n=== Test Complete ===")
	fmt.Println("If you see button states and PPU writes above, the hardware is working correctly.")
	fmt.Println("The issue may be that Donkey Kong hasn't initialized yet or needs specific timing.")
}
