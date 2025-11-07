package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/cartridge"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: dump-tile <rom-file> <tile-number-hex>")
		fmt.Println("Example: dump-tile roms/donkeykong.nes 24")
		os.Exit(1)
	}

	romPath := os.Args[1]
	var tileNum uint16
	fmt.Sscanf(os.Args[2], "%x", &tileNum)

	cart, err := cartridge.LoadFromFile(romPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	mapper := cart.GetMapper()

	// Each tile is 16 bytes: 8 bytes for low bit plane, 8 bytes for high bit plane
	baseAddr := tileNum * 16

	fmt.Printf("Tile $%02X (address $%04X in CHR-ROM):\n\n", tileNum, baseAddr)

	// Read the tile data
	lowPlane := make([]uint8, 8)
	highPlane := make([]uint8, 8)

	for row := 0; row < 8; row++ {
		lowPlane[row] = mapper.ReadCHR(baseAddr + uint16(row))
		highPlane[row] = mapper.ReadCHR(baseAddr + uint16(row) + 8)
	}

	// Print hex dump
	fmt.Println("Low bit plane:")
	for i := 0; i < 8; i++ {
		fmt.Printf("  Row %d: $%02X  (%08b)\n", i, lowPlane[i], lowPlane[i])
	}

	fmt.Println("\nHigh bit plane:")
	for i := 0; i < 8; i++ {
		fmt.Printf("  Row %d: $%02X  (%08b)\n", i, highPlane[i], highPlane[i])
	}

	// Render the tile visually
	fmt.Println("\nVisual representation (using palette 0):")
	fmt.Println("  0=black, 1=color1, 2=color2, 3=color3")
	fmt.Println()

	for row := 0; row < 8; row++ {
		fmt.Print("  ")
		for col := 0; col < 8; col++ {
			// Extract bit from each plane
			bit := 7 - col
			lowBit := (lowPlane[row] >> bit) & 1
			highBit := (highPlane[row] >> bit) & 1
			pixelValue := (highBit << 1) | lowBit

			// Print pixel value
			fmt.Printf("%d", pixelValue)
		}
		fmt.Println()
	}

	// Also show as ASCII art
	chars := " .+#"
	fmt.Println("\nASCII art:")
	for row := 0; row < 8; row++ {
		fmt.Print("  ")
		for col := 0; col < 8; col++ {
			bit := 7 - col
			lowBit := (lowPlane[row] >> bit) & 1
			highBit := (highPlane[row] >> bit) & 1
			pixelValue := (highBit << 1) | lowBit
			fmt.Printf("%c", chars[pixelValue])
		}
		fmt.Println()
	}
}
