package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: trace-ppumask <rom-file>")
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

	lastMask := ppuUnit.ReadCPURegister(0x2001)
	fmt.Printf("Frame   0: PPUMASK=$%02X\n", lastMask)

	// Run 150 frames and track PPUMASK changes
	for frame := 1; frame <= 150; frame++ {
		emulator.RunFrame()

		currentMask := ppuUnit.ReadCPURegister(0x2001)
		if currentMask != lastMask {
			renderBG := (currentMask>>3)&1 != 0
			renderSpr := (currentMask>>4)&1 != 0
			fmt.Printf("Frame %3d: PPUMASK=$%02X (BG=%v, Spr=%v)\n",
				frame, currentMask, renderBG, renderSpr)
			lastMask = currentMask
		}
	}

	fmt.Printf("\nFinal PPUMASK: $%02X\n", lastMask)
}
