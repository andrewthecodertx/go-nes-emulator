package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/cartridge"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dump-chr <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	fmt.Printf("Loading %s...\n", romPath)
	cart, err := cartridge.LoadFromFile(romPath)
	if err != nil {
		fmt.Printf("Error loading ROM: %v\n", err)
		os.Exit(1)
	}

	mapper := cart.GetMapper()

	fmt.Println("\nSampling CHR-ROM/RAM data:")

	// Check pattern table 0 - first few tiles
	fmt.Println("\nPattern Table 0, Tile 0 (first 16 bytes):")
	for i := uint16(0); i < 16; i++ {
		val := mapper.ReadCHR(i)
		fmt.Printf("  [$%04X] = $%02X\n", i, val)
	}

	// Check if CHR data has any non-zero bytes
	fmt.Println("\nScanning for non-zero CHR data...")
	nonZeroCount := 0
	firstNonZero := uint16(0xFFFF)
	for i := uint16(0); i < 0x2000; i++ {
		val := mapper.ReadCHR(i)
		if val != 0 {
			nonZeroCount++
			if firstNonZero == 0xFFFF {
				firstNonZero = i
				fmt.Printf("  First non-zero byte at $%04X = $%02X\n", i, val)
			}
		}
	}

	fmt.Printf("  Total non-zero bytes in CHR: %d out of 8192 (%.1f%%)\n",
		nonZeroCount, float64(nonZeroCount)*100.0/8192.0)

	if nonZeroCount == 0 {
		fmt.Println("  WARNING: CHR-ROM/RAM appears to be all zeros!")
		fmt.Println("  This means no tile patterns are defined yet.")
	}

	// Sample a few PRG-ROM bytes
	fmt.Println("\nSampling PRG-ROM data:")
	fmt.Println("  Reset vector area ($FFFC-$FFFD):")
	vectorLow := mapper.ReadPRG(0xFFFC)
	vectorHigh := mapper.ReadPRG(0xFFFD)
	fmt.Printf("    Reset vector: $%04X\n", uint16(vectorHigh)<<8|uint16(vectorLow))

	// Sample some code
	fmt.Println("  First 16 bytes at $8000:")
	for i := uint16(0x8000); i < 0x8010; i++ {
		val := mapper.ReadPRG(i)
		fmt.Printf("    [$%04X] = $%02X\n", i, val)
	}
}
