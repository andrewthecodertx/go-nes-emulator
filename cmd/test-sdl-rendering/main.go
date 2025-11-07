package main

import (
	"fmt"
	"log"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	ScreenWidth  = 256
	ScreenHeight = 240
	WindowScale  = 3
)

func main() {
	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		log.Fatalf("Failed to initialize SDL: %v", err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(
		"SDL Rendering Test",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		ScreenWidth*WindowScale, ScreenHeight*WindowScale,
		sdl.WINDOW_SHOWN,
	)
	if err != nil {
		log.Fatalf("Failed to create window: %v", err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		log.Fatalf("Failed to create renderer: %v", err)
	}
	defer renderer.Destroy()

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

	// Create a test pattern: vertical stripes of red, green, blue, white
	pixels := make([]byte, ScreenWidth*ScreenHeight*3)
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			i := y*ScreenWidth + x
			if x < 64 {
				// Red
				pixels[i*3+0] = 255
				pixels[i*3+1] = 0
				pixels[i*3+2] = 0
			} else if x < 128 {
				// Green
				pixels[i*3+0] = 0
				pixels[i*3+1] = 255
				pixels[i*3+2] = 0
			} else if x < 192 {
				// Blue
				pixels[i*3+0] = 0
				pixels[i*3+1] = 0
				pixels[i*3+2] = 255
			} else {
				// White
				pixels[i*3+0] = 255
				pixels[i*3+1] = 255
				pixels[i*3+2] = 255
			}
		}
	}

	fmt.Println("Test pattern: Red | Green | Blue | White")
	fmt.Println("If you see this correctly, SDL is working fine")
	fmt.Println("Press ESC to quit")

	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if e.Type == sdl.KEYDOWN && e.Keysym.Sym == sdl.K_ESCAPE {
					running = false
				}
			}
		}

		// Upload texture
		texture.Update(nil, unsafe.Pointer(&pixels[0]), ScreenWidth*3)

		// Render
		renderer.Clear()
		renderer.Copy(texture, nil, nil)
		renderer.Present()

		sdl.Delay(16)
	}
}
