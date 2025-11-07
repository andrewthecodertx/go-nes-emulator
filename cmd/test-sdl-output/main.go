package main

import (
	"fmt"
	"os"

	"github.com/andrewthecodertx/nes-emulator/pkg/nes"
	"github.com/andrewthecodertx/nes-emulator/pkg/ppu"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test-sdl-output <rom-file>")
		os.Exit(1)
	}

	romPath := os.Args[1]

	emulator, err := nes.New(romPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	emulator.Reset()

	// Run 300 frames like SDL does (120 init + more)
	for i := 0; i < 300; i++ {
		emulator.RunFrame()
	}

	// Simulate SDL's conversion
	frameBuffer := emulator.GetFrameBuffer()
	pixels := make([]byte, 256*240*3)

	for i := 0; i < 256*240; i++ {
		paletteIndex := frameBuffer[i]

		// Check bounds
		if paletteIndex >= uint8(len(ppu.HardwarePalette)) {
			fmt.Printf("ERROR: Palette index %d out of bounds at pixel %d!\n", paletteIndex, i)
			paletteIndex = 0
		}

		color := ppu.HardwarePalette[paletteIndex]

		pixels[i*3+0] = color.R
		pixels[i*3+1] = color.G
		pixels[i*3+2] = color.B
	}

	// Sample the pixels that SDL would be rendering
	fmt.Println("Simulating SDL conversion - sampling scanline 60:")
	fmt.Println("Pixel | Pal Idx | RGB")
	fmt.Println("------|---------|----------")

	for x := 0; x < 32; x++ {
		pixelIdx := 60*256 + x
		paletteIdx := frameBuffer[pixelIdx]
		r := pixels[pixelIdx*3+0]
		g := pixels[pixelIdx*3+1]
		b := pixels[pixelIdx*3+2]

		fmt.Printf("%5d | $%02X     | %3d,%3d,%3d", x, paletteIdx, r, g, b)

		// Identify color
		if r > 200 && g < 100 && b > 150 {
			fmt.Print(" <- MAGENTA!")
		} else if r < 100 && g > 150 && b > 150 {
			fmt.Print(" <- CYAN")
		} else if r < 50 && g < 50 && b < 50 {
			fmt.Print(" <- BLACK")
		}
		fmt.Println()
	}

	// Count RGB color distribution
	colorCounts := make(map[[3]uint8]int)
	for i := 0; i < 256*240; i++ {
		r := pixels[i*3+0]
		g := pixels[i*3+1]
		b := pixels[i*3+2]
		colorCounts[[3]uint8{r, g, b}]++
	}

	fmt.Printf("\nTotal unique RGB colors: %d\n", len(colorCounts))

	// Count magenta vs cyan
	magentaCount := 0
	cyanCount := 0
	blackCount := 0

	for i := 0; i < 256*240; i++ {
		r := pixels[i*3+0]
		g := pixels[i*3+1]
		b := pixels[i*3+2]

		if r > 200 && g < 150 && b > 150 {
			magentaCount++
		} else if r < 100 && g > 150 && b > 150 {
			cyanCount++
		} else if r < 50 && g < 50 && b < 50 {
			blackCount++
		}
	}

	fmt.Printf("\nColor analysis:\n")
	fmt.Printf("  Black pixels:   %d (%.1f%%)\n", blackCount, float64(blackCount)*100.0/float64(256*240))
	fmt.Printf("  Cyan pixels:    %d (%.1f%%)\n", cyanCount, float64(cyanCount)*100.0/float64(256*240))
	fmt.Printf("  Magenta pixels: %d (%.1f%%)\n", magentaCount, float64(magentaCount)*100.0/float64(256*240))

	if magentaCount > 1000 {
		fmt.Println("\n⚠ WARNING: Lots of magenta detected! This would match your screenshot.")
	} else {
		fmt.Println("\n✓ Magenta count is low - SDL output should look correct.")
	}
}
