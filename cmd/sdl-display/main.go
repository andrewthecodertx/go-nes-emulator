package main

import (
	"fmt"
	"log"
	"os"
	"unsafe"

	"github.com/andrewthecodertx/go-nes-emulator/pkg/controller"
	"github.com/andrewthecodertx/go-nes-emulator/pkg/nes"
	"github.com/andrewthecodertx/go-nes-emulator/pkg/ppu"
	"github.com/veandco/go-sdl2/sdl"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 240
	WindowScale  = 3 // Scale factor for display
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: sdl-display <rom-file>")
		fmt.Println("Example: sdl-display ../../roms/donkeykong.nes")
		os.Exit(1)
	}

	romPath := os.Args[1]

	// Initialize SDL
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		log.Fatalf("Failed to initialize SDL: %v", err)
	}
	defer sdl.Quit()

	// Create window
	window, err := sdl.CreateWindow(
		"NES Emulator - "+romPath,
		sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		ScreenWidth*WindowScale,
		ScreenHeight*WindowScale,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}
	defer window.Destroy()

	// Create renderer
	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}
	defer renderer.Destroy()

	// Create texture for NES display (256x240)
	// Try RGB24 format
	texture, err := renderer.CreateTexture(
		sdl.PIXELFORMAT_RGB24,
		sdl.TEXTUREACCESS_STREAMING,
		ScreenWidth,
		ScreenHeight,
	)
	if err != nil {
		log.Fatalf("Failed to create texture: %v", err)
	}
	defer texture.Destroy()

	// Load NES ROM
	fmt.Printf("\n=== Loading ROM ===\n")
	fmt.Printf("File: %s\n", romPath)
	emulator, err := nes.New(romPath)
	if err != nil {
		log.Fatalf("Failed to load ROM: %v", err)
	}

	// Show cartridge info
	cart := emulator.GetCartridge()
	fmt.Printf("Mapper: %d\n", cart.GetMapperID())
	fmt.Printf("PRG Banks: %d x 16KB = %dKB\n", cart.GetPRGBanks(), cart.GetPRGBanks()*16)
	fmt.Printf("CHR Banks: %d x 8KB = %dKB\n", cart.GetCHRBanks(), cart.GetCHRBanks()*8)

	// Reset NES to power-on state
	emulator.Reset()

	// Buffer for RGB pixels (256x240x3 bytes)
	pixels := make([]byte, ScreenWidth*ScreenHeight*3)

	// Run many frames to let the game initialize
	fmt.Println("\nInitializing (2 seconds)...")
	for i := 0; i < 120; i++ { // ~2 seconds at 60 FPS
		emulator.RunFrame()
	}

	// Get PPU state and controller
	ppuUnit := emulator.GetPPU()
	ctrl := emulator.GetBus().GetController(0)

	fmt.Println("\n=== NES Emulator Ready ===")
	fmt.Println("System: ESC=quit | P=pause | SPACE=step | R=reset | F=force render | D=debug")
	fmt.Println("Game:   Arrows=D-pad | Z=B | X=A | Enter=Start | RShift=Select")
	fmt.Println("==========================")

	running := true
	paused := false
	frameCount := 0
	forceRendering := false
	debugFrame := false // Disabled by default - press D to enable

	for running {
		// Handle events
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false

			case *sdl.KeyboardEvent:
				pressed := e.Type == sdl.KEYDOWN

				// Handle system keys (only on key down)
				if pressed {
					switch e.Keysym.Sym {
					case sdl.K_ESCAPE:
						running = false
						continue
					case sdl.K_SPACE:
						// Step one frame when paused
						if paused {
							emulator.RunFrame()
							frameCount++
							fmt.Printf("Frame %d rendered\n", frameCount)
						}
						continue
					case sdl.K_p:
						// Toggle pause
						paused = !paused
						if paused {
							fmt.Println("Paused (press SPACE to step, P to resume)")
						} else {
							fmt.Println("Resumed")
						}
						continue
					case sdl.K_r:
						// Reset
						emulator.Reset()
						if forceRendering {
							ppuUnit.WriteCPURegister(0x2001, 0x1E)
						}
						frameCount = 0
						fmt.Println("Reset")
						continue
					case sdl.K_f:
						// Toggle forced rendering
						forceRendering = !forceRendering
						if forceRendering {
							ppuUnit.WriteCPURegister(0x2001, 0x1E)
							fmt.Println("Forced rendering ON (background+sprites enabled)")
						} else {
							ppuUnit.WriteCPURegister(0x2001, 0x00)
							fmt.Println("Forced rendering OFF (game controls PPU)")
						}
						continue
					case sdl.K_d:
						// Toggle debug output
						debugFrame = !debugFrame
						if debugFrame {
							fmt.Println("Debug output ON")
						} else {
							fmt.Println("Debug output OFF")
						}
						continue
					}
				}

				// Handle game controller keys (both down and up)
				switch e.Keysym.Sym {
				case sdl.K_x:
					ctrl.SetButton(controller.ButtonA, pressed)
				case sdl.K_z:
					ctrl.SetButton(controller.ButtonB, pressed)
				case sdl.K_RSHIFT:
					ctrl.SetButton(controller.ButtonSelect, pressed)
				case sdl.K_RETURN:
					ctrl.SetButton(controller.ButtonStart, pressed)
				case sdl.K_UP:
					ctrl.SetButton(controller.ButtonUp, pressed)
				case sdl.K_DOWN:
					ctrl.SetButton(controller.ButtonDown, pressed)
				case sdl.K_LEFT:
					ctrl.SetButton(controller.ButtonLeft, pressed)
				case sdl.K_RIGHT:
					ctrl.SetButton(controller.ButtonRight, pressed)
				}
			}
		}

		// Run emulation if not paused
		if !paused {
			emulator.RunFrame()
			frameCount++
		}

		// Convert frame buffer to RGB
		frameBuffer := emulator.GetFrameBuffer()

		// Track unique colors for debug info
		colorCounts := make(map[uint8]int)
		uniqueColors := 0

		for i := 0; i < ScreenWidth*ScreenHeight; i++ {
			paletteIndex := frameBuffer[i]

			// Track color usage
			if colorCounts[paletteIndex] == 0 {
				uniqueColors++
			}
			colorCounts[paletteIndex]++

			// Bounds check - palette indices should be 0-63
			if paletteIndex >= 64 {
				if debugFrame {
					fmt.Printf("ERROR: palette index %d out of bounds at pixel %d\n", paletteIndex, i)
				}
				paletteIndex = 0x0F // Black
			}

			color := ppu.HardwarePalette[paletteIndex]

			// Write pixels in RGB order for RGB24 format
			pixels[i*3+0] = color.R
			pixels[i*3+1] = color.G
			pixels[i*3+2] = color.B
		}

		// Show periodic status updates
		if frameCount%60 == 0 {
			// Find most common color
			maxCount := 0
			mostCommonColor := uint8(0)
			for color, count := range colorCounts {
				if count > maxCount {
					maxCount = count
					mostCommonColor = color
				}
			}

			if debugFrame {
				fmt.Printf("[Frame %4d] Colors: %d unique | Most common: $%02X (%d pixels)\n",
					frameCount, uniqueColors, mostCommonColor, maxCount)
			} else if frameCount%300 == 0 {
				// Less frequent updates when debug is off
				fmt.Printf("[Frame %d] Running... (press D for debug info)\n", frameCount)
			}
		}

		texture.Update(nil, unsafe.Pointer(&pixels[0]), ScreenWidth*3)

		renderer.Clear()
		renderer.Copy(texture, nil, nil)
		renderer.Present()

		// ~60 FPS
		if !paused {
			sdl.Delay(16)
		} else {
			sdl.Delay(100) // Slower refresh when paused
		}
	}

	fmt.Printf("\nTotal frames rendered: %d\n", frameCount)
}
