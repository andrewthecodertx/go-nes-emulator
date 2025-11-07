package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
	"github.com/andrewthecodertx/nes-emulator/pkg/ppu"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dump-palette <rom-file> [frames]")
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

	// Run frames
	fmt.Printf("Running %d frames...\n\n", frames)
	for i := 0; i < frames; i++ {
		emulator.RunFrame()
	}

	// Read palette RAM directly through PPU
	bus := emulator.GetBus()

	fmt.Println("Reading Palette RAM via $2006/$2007:")
	fmt.Println()

	// Background palettes
	fmt.Println("Background Palettes:")
	for pal := 0; pal < 4; pal++ {
		fmt.Printf("  Palette %d: ", pal)
		for i := 0; i < 4; i++ {
			addr := uint16(0x3F00 + pal*4 + i)

			// Set PPU address
			bus.Write(0x2006, uint8(addr>>8))
			bus.Write(0x2006, uint8(addr&0xFF))

			// Read data
			value := bus.Read(0x2007)

			if value < uint8(len(ppu.HardwarePalette)) {
				color := ppu.HardwarePalette[value]
				fmt.Printf("$%02X(#%02X%02X%02X) ", value, color.R, color.G, color.B)
			} else {
				fmt.Printf("$%02X(????) ", value)
			}
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("Sprite Palettes:")
	for pal := 0; pal < 4; pal++ {
		fmt.Printf("  Palette %d: ", pal)
		for i := 0; i < 4; i++ {
			addr := uint16(0x3F10 + pal*4 + i)

			// Set PPU address
			bus.Write(0x2006, uint8(addr>>8))
			bus.Write(0x2006, uint8(addr&0xFF))

			// Read data
			value := bus.Read(0x2007)

			if value < uint8(len(ppu.HardwarePalette)) {
				color := ppu.HardwarePalette[value]
				fmt.Printf("$%02X(#%02X%02X%02X) ", value, color.R, color.G, color.B)
			} else {
				fmt.Printf("$%02X(????) ", value)
			}
		}
		fmt.Println()
	}

	// Also check what the frame buffer is outputting
	fmt.Println()
	fmt.Println("Frame buffer sample (scanline 60, pixels 0-31):")
	frameBuffer := emulator.GetFrameBuffer()
	for x := 0; x < 32; x++ {
		idx := frameBuffer[60*256+x]
		if idx < uint8(len(ppu.HardwarePalette)) {
			color := ppu.HardwarePalette[idx]
			fmt.Printf("$%02X(#%02X%02X%02X) ", idx, color.R, color.G, color.B)
		}
		if (x+1)%8 == 0 {
			fmt.Println()
		}
	}
}
