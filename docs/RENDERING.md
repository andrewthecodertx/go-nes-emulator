# NES PPU Rendering Implementation

This document explains the PPU rendering system that was just implemented.

## Overview

The PPU (Picture Processing Unit) now includes **full background tile rendering** using cycle-accurate emulation based on the reference PHP implementation at `/home/andrew/Projects/NES`.

## What Was Implemented

### 1. Loopy Register Scroll Methods (`pkg/ppu/registers.go`)

Added hardware-accurate scrolling operations to the `LoopyRegister`:

- **`IncrementX()`**: Moves right by 1 tile, wraps and flips nametable
- **`IncrementY()`**: Moves down by 1 scanline, handles fine Y overflow
- **`TransferX(source)`**: Copies horizontal position (end of scanline)
- **`TransferY(source)`**: Copies vertical position (pre-render scanline)

These implement the "Loopy" scrolling algorithm discovered through reverse-engineering.

### 2. Hardware Color Palette (`pkg/ppu/palette.go`)

- **64-color NTSC palette**: The standard NES hardware palette as RGB values
- **`GetColorFromPalette()`**: Translates palette index + pixel value → RGB color

### 3. Background Rendering State (`pkg/ppu/ppu.go`)

Added to PPU struct:

```go
// Tile fetching state
bgNextTileID     uint8  // Next tile from nametable
bgNextTileAttrib uint8  // Palette selection (2 bits)
bgNextTileLSB    uint8  // Pattern low byte
bgNextTileMSB    uint8  // Pattern high byte

// 16-bit shift registers
bgShifterPatternLo uint16  // Pixel data low bit
bgShifterPatternHi uint16  // Pixel data high bit
bgShifterAttribLo  uint16  // Palette data low bit
bgShifterAttribHi  uint16  // Palette data high bit
```

### 4. Rendering Helper Functions (`pkg/ppu/rendering.go`)

- **`loadBackgroundShifters()`**: Loads next 8 pixels into shift registers
- **`updateShifters()`**: Shifts registers left by 1 bit each cycle

### 5. Full Clock Cycle Rendering (`pkg/ppu/ppu.go` - Clock method)

Implemented cycle-accurate PPU timing:

#### **Pre-render Scanline (-1)**
- Cycle 1: Clear VBlank, Sprite 0 Hit, Sprite Overflow flags
- Cycles 280-304: Restore vertical scroll position
- Fetches background data (preparing for next frame)

#### **Visible Scanlines (0-239)**
- **Cycles 2-257**: Active rendering
  - Every 8 cycles: Fetch tile data in 4 steps
    1. Cycle 0: Load shifters, fetch tile ID
    2. Cycle 2: Fetch attribute byte
    3. Cycle 4: Fetch pattern low byte
    4. Cycle 6: Fetch pattern high byte
    5. Cycle 7: Increment horizontal scroll
  - Every cycle: Shift registers, output pixel
- **Cycle 256**: Increment vertical scroll
- **Cycle 257**: Reset horizontal scroll position
- **Cycles 321-337**: Prefetch first 2 tiles of next scanline
- **Cycles 338, 340**: Dummy nametable reads

#### **Post-render Scanline (240)**
- Idle (no rendering)

#### **VBlank Scanlines (241-260)**
- Cycle 1 of scanline 241: Set VBlank flag, trigger NMI

#### **Pixel Output**
During cycles 1-256 of visible scanlines:
1. Extract 2-bit pixel value from shifters using fine X scroll
2. Extract 2-bit palette selection from attribute shifters
3. Look up color in palette RAM
4. Write palette index to frame buffer

#### **Odd Frame Behavior**
On odd frames with rendering enabled, cycle 0 of scanline 0 is skipped (jumps to cycle 1).

## How Background Rendering Works

### The 8-Cycle Pattern

The PPU fetches one 8x8 tile every 8 cycles:

```
Cycle 0: Load previous tile into shifters, fetch next tile ID
Cycle 2: Fetch attribute byte (palette selection)
Cycle 4: Fetch pattern table low byte (pixel data bit 0)
Cycle 6: Fetch pattern table high byte (pixel data bit 1)
Cycle 7: Increment coarse X (move to next tile)
```

### Shift Registers

The PPU uses 16-bit shift registers:
- **High 8 bits**: Currently rendering (output to screen)
- **Low 8 bits**: Next tile (loaded every 8 cycles)

Every cycle, registers shift left by 1:
```
Cycle N:   [AAAAAAAA BBBBBBBB] → shift left
Cycle N+1: [AAAAAAA BBBBBBBB_]
```

After 8 shifts, the "B" pixels move into the "A" position, and new pixels are loaded into "B".

### Pixel Composition

For each visible pixel:
1. Use fine X (0-7) to select which bit of the high byte
2. Combine bit from pattern low + pattern high = 2-bit pixel (0-3)
3. Combine bit from attrib low + attrib high = 2-bit palette (0-3)
4. Lookup: `palette_ram[(palette << 2) | pixel]` → color index (0-63)
5. Lookup: `hardware_palette[color_index]` → RGB color

### Scrolling

The VRAM address register tracks current position:
- **Coarse X/Y**: Which tile (0-31 horizontal, 0-29 vertical)
- **Fine Y**: Which row within tile (0-7)
- **Fine X**: Which pixel within tile (0-7, separate 3-bit register)
- **Nametable**: Which of 4 nametables

As rendering progresses:
- After each tile: IncrementX() - move right
- End of scanline (cycle 256): IncrementY() - move down
- End of scanline (cycle 257): TransferX() - reset to left edge
- During pre-render: TransferY() - reset to top

## Frame Buffer Format

Currently, the frame buffer stores **palette indices** (0-63):

```go
frameBuffer [256 * 240]uint8
```

To render to screen, you'd:
1. Read palette index from frame buffer
2. Look up RGB color: `palette.HardwarePalette[index]`
3. Write to display buffer

## What's NOT Implemented Yet

- ❌ **Sprite rendering** (OAM evaluation, sprite fetching, sprite output)
- ❌ **Sprite 0 hit detection**
- ❌ **Sprite overflow detection**
- ❌ **Background clipping** (leftmost 8 pixels)
- ❌ **Sprite clipping** (leftmost 8 pixels)
- ❌ **Grayscale mode**
- ❌ **Color emphasis bits**

## Testing

The implementation compiles and runs, but you'll need:

1. **A ROM file** with valid CHR-ROM data
2. **A display frontend** to visualize the frame buffer

Example test (with a real ROM):

```go
emulator, _ := nes.New("roms/game.nes")
emulator.Reset()

// Enable rendering
// (Normally set by the game, but you can force it for testing)
ppu := emulator.GetPPU()
ppu.WriteCPURegister(0x2001, 0x08) // Enable background rendering

// Run one frame
emulator.RunFrame()

// Get frame buffer
frameBuffer := emulator.GetFrameBuffer()

// Render to screen (pseudo-code)
for y := 0; y < 240; y++ {
    for x := 0; x < 256; x++ {
        paletteIndex := frameBuffer[y*256 + x]
        color := ppu.HardwarePalette[paletteIndex]
        setPixel(x, y, color.R, color.G, color.B)
    }
}
```

## Next Steps

To make rendering visible, you need to:

1. **Create a display frontend** (SDL2, OpenGL, web canvas, etc.)
2. **Load a test ROM** (like nestest.nes or a simple homebrew)
3. **Convert frame buffer** from palette indices to RGB
4. **Display the image** at 60 FPS

## Performance

The implementation is cycle-accurate, meaning:
- One `Clock()` call = one PPU cycle
- One frame = 89,341.5 PPU cycles = 29,780.5 CPU cycles
- Target: 60 FPS = 1,789,773 CPU cycles per second

For real-time emulation, you'd run the main loop at ~60 Hz and execute enough cycles per frame.

## References

- [NESDev Wiki - PPU Rendering](https://www.nesdev.org/wiki/PPU_rendering)
- [NESDev Wiki - PPU Scrolling](https://www.nesdev.org/wiki/PPU_scrolling)
- [Loopy's PPU Scrolling Document](https://www.nesdev.org/wiki/PPU_scrolling#Tile_and_attribute_fetching)

---

Generated with assistance from Claude Code (Anthropic), based on the PHP NES emulator at `/home/andrew/Projects/NES`
