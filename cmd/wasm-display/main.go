//go:build js && wasm

package main

import (
	"fmt"
	"syscall/js"

	"github.com/andrewthecodertx/go-nes-emulator/pkg/cartridge"
	"github.com/andrewthecodertx/go-nes-emulator/pkg/controller"
	"github.com/andrewthecodertx/go-nes-emulator/pkg/nes"
	"github.com/andrewthecodertx/go-nes-emulator/pkg/ppu"
)

const (
	screenWidth  = 256
	screenHeight = 240
)

var (
	emulator    *nes.NES
	ctrl        *controller.Controller
	canvas      js.Value
	ctx         js.Value
	imageData   js.Value
	pixelArray  js.Value
	running     bool
	paused      bool
	loopStarted bool

	pixels      []byte
	rgbaPalette [64][4]byte

	lastFrameTime float64
	frameInterval float64 = 1000.0 / 60.0
)

func init() {
	pixels = make([]byte, screenWidth*screenHeight*4)

	for i := 0; i < 64; i++ {
		color := ppu.HardwarePalette[i]
		rgbaPalette[i] = [4]byte{color.R, color.G, color.B, 255}
	}
}

func main() {
	fmt.Println("NES Emulator WASM initialized")

	js.Global().Set("nesLoadROM", js.FuncOf(loadROM))
	js.Global().Set("nesReset", js.FuncOf(reset))
	js.Global().Set("nesPause", js.FuncOf(pause))
	js.Global().Set("nesResume", js.FuncOf(resume))
	js.Global().Set("nesSetButton", js.FuncOf(setButton))
	js.Global().Set("nesStep", js.FuncOf(step))

	select {}
}

func loadROM(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return js.ValueOf("Error: no ROM data provided")
	}

	jsArray := args[0]
	length := jsArray.Length()
	romData := make([]byte, length)
	js.CopyBytesToGo(romData, jsArray)

	cart, err := cartridge.LoadFromBytes(romData)
	if err != nil {
		return js.ValueOf(fmt.Sprintf("Error loading ROM: %v", err))
	}

	emulator = nes.NewFromCartridge(cart)
	emulator.Reset()

	ctrl = emulator.GetBus().GetController(0)

	document := js.Global().Get("document")
	canvas = document.Call("getElementById", "nes-canvas")
	if canvas.IsNull() || canvas.IsUndefined() {
		return js.ValueOf("Error: canvas element 'nes-canvas' not found")
	}

	ctx = canvas.Call("getContext", "2d")
	imageData = ctx.Call("createImageData", screenWidth, screenHeight)
	pixelArray = imageData.Get("data")

	fmt.Println("Initializing emulator...")
	for i := 0; i < 120; i++ {
		emulator.RunFrame()
	}

	running = true
	paused = false
	if !loopStarted {
		loopStarted = true
		renderLoopFunc = js.FuncOf(renderLoop)
		js.Global().Call("requestAnimationFrame", renderLoopFunc)
	}

	info := fmt.Sprintf("ROM loaded - Mapper: %d, PRG: %dKB, CHR: %dKB",
		cart.GetMapperID(),
		cart.GetPRGBanks()*16,
		cart.GetCHRBanks()*8)
	fmt.Println(info)

	return js.ValueOf(info)
}

var renderLoopFunc js.Func

func renderLoop(this js.Value, args []js.Value) interface{} {
	js.Global().Call("requestAnimationFrame", renderLoopFunc)

	if !running || paused || emulator == nil {
		return nil
	}

	now := args[0].Float()

	elapsed := now - lastFrameTime
	if elapsed < frameInterval {
		return nil
	}
	lastFrameTime = now

	emulator.RunFrame()

	renderFrame()

	return nil
}

func renderFrame() {
	frameBuffer := emulator.GetFrameBuffer()

	for i := 0; i < screenWidth*screenHeight; i++ {
		paletteIndex := frameBuffer[i] & 0x3F
		rgba := rgbaPalette[paletteIndex]
		offset := i * 4
		pixels[offset+0] = rgba[0]
		pixels[offset+1] = rgba[1]
		pixels[offset+2] = rgba[2]
		pixels[offset+3] = rgba[3]
	}

	js.CopyBytesToJS(pixelArray, pixels)

	ctx.Call("putImageData", imageData, 0, 0)
}

func reset(this js.Value, args []js.Value) interface{} {
	if emulator != nil {
		emulator.Reset()
		fmt.Println("Emulator reset")
	}
	return nil
}

func pause(this js.Value, args []js.Value) interface{} {
	paused = true
	fmt.Println("Emulator paused")
	return nil
}

func resume(this js.Value, args []js.Value) interface{} {
	paused = false
	fmt.Println("Emulator resumed")
	return nil
}

func step(this js.Value, args []js.Value) interface{} {
	if emulator != nil && paused {
		emulator.RunFrame()
		renderFrame()
		fmt.Println("Stepped one frame")
	}
	return nil
}

func setButton(this js.Value, args []js.Value) interface{} {
	if ctrl == nil || len(args) < 2 {
		return nil
	}

	buttonName := args[0].String()
	pressed := args[1].Bool()

	var button controller.Button
	switch buttonName {
	case "a":
		button = controller.ButtonA
	case "b":
		button = controller.ButtonB
	case "select":
		button = controller.ButtonSelect
	case "start":
		button = controller.ButtonStart
	case "up":
		button = controller.ButtonUp
	case "down":
		button = controller.ButtonDown
	case "left":
		button = controller.ButtonLeft
	case "right":
		button = controller.ButtonRight
	default:
		return nil
	}

	ctrl.SetButton(button, pressed)
	return nil
}
