package main

import (
	"fmt"

	"github.com/andrewthecodertx/go-nes-emulator/pkg/ppu"
)

func main() {
	fmt.Println("Verifying Hardware Palette Colors")
	fmt.Println("==================================")
	fmt.Println()

	// Check key palette indices from Donkey Kong
	indices := []uint8{0x0F, 0x2C, 0x27, 0x12, 0x30, 0x38, 0x00, 0x25}
	names := []string{"Black", "Cyan", "Orange", "Blue", "White", "Yellow", "Gray", "Magenta"}

	for i, idx := range indices {
		if int(idx) < len(ppu.HardwarePalette) {
			color := ppu.HardwarePalette[idx]
			fmt.Printf("Index $%02X (%s):\n", idx, names[i])
			fmt.Printf("  RGB: #%02X%02X%02X (%d, %d, %d)\n",
				color.R, color.G, color.B, color.R, color.G, color.B)
			fmt.Println()
		}
	}

	// Check the entire palette for magenta colors
	fmt.Println("Scanning for Magenta-ish colors (R>200, B>150, G<150):")
	for i := 0; i < len(ppu.HardwarePalette); i++ {
		color := ppu.HardwarePalette[i]
		if color.R > 200 && color.B > 150 && color.G < 150 {
			fmt.Printf("  $%02X: #%02X%02X%02X\n", i, color.R, color.G, color.B)
		}
	}

	fmt.Println()
	fmt.Println("Scanning for Cyan-ish colors (G>150, B>150, R<100):")
	for i := 0; i < len(ppu.HardwarePalette); i++ {
		color := ppu.HardwarePalette[i]
		if color.G > 150 && color.B > 150 && color.R < 100 {
			fmt.Printf("  $%02X: #%02X%02X%02X\n", i, color.R, color.G, color.B)
		}
	}
}
