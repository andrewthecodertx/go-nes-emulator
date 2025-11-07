# SDL2 Display Frontend - Quick Start

An SDL2-based graphical frontend has been created for visualizing NES emulation!

## What Was Created

### Files
- **`cmd/sdl-display/main.go`** - Complete SDL2 application (190 lines)
- **`cmd/sdl-display/README.md`** - Frontend documentation

### Features
- ✅ 256x240 NES display scaled 3x (768x720 window)
- ✅ ~60 FPS emulation
- ✅ Palette index → RGB conversion using hardware palette
- ✅ Keyboard controls (ESC, P, SPACE, R)
- ✅ Frame counter and status output

## Building

The SDL2 bindings require CGO, so compilation takes a bit longer than normal Go programs.

```bash
# Navigate to the frontend directory
cd /home/andrew/Projects/nes-emulator/cmd/sdl-display

# Build (this may take 1-2 minutes due to CGO)
go build -o nes-sdl

# The binary will be: nes-sdl
```

## Running

```bash
# Run with Donkey Kong
./nes-sdl ../../roms/donkeykong.nes

# Run with Tetris
./nes-sdl ../../roms/tetris.nes

# Run with nestest
./nes-sdl ../../roms/nestest.nes
```

## Controls

### System Controls

| Key | Action |
|-----|--------|
| **ESC** | Quit emulator |
| **P** | Pause/Resume |
| **SPACE** | Step one frame (when paused) |
| **R** | Reset emulator |
| **F** | Toggle forced rendering (debug) |

### Game Controls (NES Controller)

| Key | NES Button |
|-----|------------|
| **Arrow Keys** | D-Pad (Up/Down/Left/Right) |
| **X** | A Button |
| **Z** | B Button |
| **Enter** | START |
| **Right Shift** | SELECT |

## What To Expect

When you run the emulator, you should see:

1. **Console output:**
   ```
   Loading ROM: ../../roms/donkeykong.nes
   NES initialized
   Rendering enabled
   Running initialization frames...
   Starting main loop (press ESC to exit, SPACE for next frame)
   ```

2. **Window:** 768x720 pixel window showing the NES display

3. **Graphics:** Background tiles rendered based on CHR-ROM data

### Note About ROM Content

- **nestest.nes**: CPU test ROM, may have minimal/test graphics
- **donkeykong.nes**: Should show Donkey Kong title screen graphics
- **tetris.nes**: Should show Tetris graphics

The emulator is currently rendering **backgrounds only** (no sprites yet), so you'll see:
- Background tiles from the pattern tables
- Nametable layout
- Palette colors

## Troubleshooting

### Build takes forever
- This is normal! CGO compilation of SDL2 bindings can take 1-2 minutes
- Be patient, it only needs to compile once

### "Failed to initialize SDL"
```bash
# Make sure SDL2 is installed
sudo apt-get install libsdl2-dev  # Ubuntu/Debian
```

### Black or garbled screen
- ROM may not be setting up PPU correctly
- Try pressing **F** to toggle forced rendering and see raw CHR-ROM data
- Try different ROMs
- Check console for errors

### Compilation errors
```bash
# Ensure dependencies are installed
cd /home/andrew/Projects/nes-emulator
go mod tidy
go get github.com/veandco/go-sdl2/sdl
```

## How It Works

```
1. Load ROM file → Parse iNES format
2. Initialize NES → Reset CPU & PPU
3. Enable rendering → Write to PPUMASK register
4. Main loop (60 FPS):
   a. Run one frame (~29,781 CPU cycles)
   b. Get PPU frame buffer (256x240 palette indices)
   c. Convert to RGB using hardware palette
   d. Upload to SDL texture
   e. Display to window
```

## Code Overview

### Main Components

**Initialization:**
```go
// Create SDL window
window := sdl.CreateWindow("NES Emulator", ...)

// Create texture for NES display
texture := renderer.CreateTexture(
    sdl.PIXELFORMAT_RGB24,
    256, 240)

// Load and reset NES
emulator := nes.New(romPath)
emulator.Reset()

// Enable background rendering
ppu.WriteCPURegister(0x2001, 0x08)
```

**Main Loop:**
```go
for running {
    // Run one frame
    emulator.RunFrame()

    // Convert palette indices to RGB
    frameBuffer := emulator.GetFrameBuffer()
    for i, paletteIdx := range frameBuffer {
        color := ppu.HardwarePalette[paletteIdx]
        pixels[i*3+0] = color.R
        pixels[i*3+1] = color.G
        pixels[i*3+2] = color.B
    }

    // Display
    texture.Update(nil, pixels, 256*3)
    renderer.Copy(texture, nil, nil)
    renderer.Present()
}
```

## Debug Feature: Forced Rendering

The **F** key toggles forced rendering mode:

- **OFF (default)**: The game's code controls the PPU. This is normal operation - you see what the game wants to display.
- **ON**: The emulator forces the PPU to render background and sprites regardless of what the game writes to PPUMASK. Useful for debugging:
  - See raw CHR-ROM tile data
  - Verify PPU is processing correctly
  - Debug games that don't initialize PPU properly

When forced rendering is ON, the emulator writes `$1E` to PPUMASK (enable background + sprites, show left 8 pixels). When OFF, it writes `$00` (rendering disabled, letting the game take over).

## Next Steps

Once you can see graphics, you might want to:

1. **Add sprite rendering** - See sprites on screen
2. **Controller input** - Map keyboard/gamepad to NES buttons
3. **Audio** - Add APU emulation for sound
4. **Save states** - Save/load emulator state
5. **Debugger** - Add pattern table viewer, nametable viewer, etc.

## Performance

The emulator should run at ~60 FPS on most systems. If it's too slow:
- Check CPU usage
- Disable the frame counter prints
- Try a different ROM

## Alternative: Quick Test Without Building

If you want to see if the NES core is working without SDL, you can use the basic example:

```bash
cd /home/andrew/Projects/nes-emulator/examples/basic
go build
./basic ../../roms/donkeykong.nes
```

This won't show graphics but will verify the ROM loads and CPU runs.

---

**Generated with assistance from Claude Code (Anthropic)**

Enjoy your NES emulator! 🎮
