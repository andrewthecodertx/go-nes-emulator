# Sprite Rendering Implementation

This document explains how sprite rendering is implemented in the NES emulator PPU.

## Overview

The NES can display up to 64 sprites (also called "objects") on screen. Each sprite is 8x8 or 8x16 pixels. However, only 8 sprites can be displayed on any single scanline due to hardware limitations.

Sprite data is stored in Object Attribute Memory (OAM), which is 256 bytes (64 sprites × 4 bytes each).

## OAM Format

Each sprite uses 4 bytes in OAM:

```
Byte 0: Y position (top of sprite, minus 1)
Byte 1: Tile index (which tile to use from pattern table)
Byte 2: Attributes
  Bit 0-1: Palette (0-3, selects from 4 sprite palettes)
  Bit 5:   Priority (0=in front of background, 1=behind background)
  Bit 6:   Flip horizontally
  Bit 7:   Flip vertically
Byte 3: X position (left of sprite)
```

## Rendering Pipeline

Sprite rendering happens in three phases each scanline:

### 1. Sprite Evaluation (Cycles 65-256)

During the visible portion of the scanline, the PPU evaluates which sprites will appear on the **next** scanline.

**Process:**
- Scan through all 64 sprites in primary OAM
- Check if sprite's Y position means it will be visible on next scanline
- Copy up to 8 sprites to secondary OAM (32 bytes)
- If more than 8 sprites found, set sprite overflow flag
- Remember if sprite 0 is present (for hit detection)

**Implementation:** `spriteEvaluation()` in `pkg/ppu/sprites.go`

### 2. Sprite Fetching (Cycles 257-320)

After sprite evaluation completes, the PPU fetches the pattern data for all sprites in secondary OAM.

**Process:**
- For each of the 8 sprites in secondary OAM:
  - Calculate which row of the sprite to display
  - Apply vertical flip if needed
  - Fetch pattern data (2 bytes: low and high bit planes)
  - Apply horizontal flip if needed
  - Store in sprite shifters

**For 8x8 sprites:**
- Pattern address = (PatternTable << 12) | (TileIndex << 4) | Row

**For 8x16 sprites:**
- Bit 0 of tile index selects pattern table
- Bits 1-7 select tile pair
- Top half uses tile N, bottom half uses tile N+1

**Implementation:** `spriteFetching()` in `pkg/ppu/sprites.go`

### 3. Sprite Rendering (Cycles 1-256, visible scanlines)

For each pixel during the visible portion of the scanline:

**Process:**
- Check each of the 8 sprites
- Determine if sprite covers this X position
- Extract pixel value from sprite shifter
- If pixel is transparent (0), skip sprite
- Return first non-transparent sprite pixel found

**Implementation:** `renderSprites()` in `pkg/ppu/sprites.go`

## Sprite Compositing

Once both background and sprite pixels are determined, they must be composited:

**Priority Rules:**
1. If both pixels are transparent (0) → Show backdrop color (palette 0, color 0)
2. If background transparent, sprite opaque → Show sprite
3. If sprite transparent, background opaque → Show background
4. If both opaque → Check sprite priority bit:
   - Priority=0 (front): Show sprite
   - Priority=1 (back): Show background

**Palette Selection:**
- Background uses palettes 0-3 ($3F00-$3F0F)
- Sprites use palettes 4-7 ($3F10-$3F1F)
- Transparency is always pixel value 0

**Implementation:** Pixel compositing in `pkg/ppu/ppu.go` Clock() method

## Sprite 0 Hit Detection

Sprite 0 (first sprite in OAM) is special. When an opaque pixel of sprite 0 overlaps with an opaque background pixel, the PPU sets the "sprite 0 hit" flag in PPUSTATUS.

**Rules:**
- Hit occurs when both pixels are non-zero
- Does not occur at X=255 (rightmost pixel)
- Does not occur if rendering disabled in leftmost 8 pixels (unless both enabled)
- Flag is cleared at dot 1 of pre-render scanline

**Use:** Games use this for timing - they can detect when the screen has been rendered to a certain point. Common use is for split-screen effects (status bar vs. scrolling playfield).

**Implementation:** In pixel compositing logic in `pkg/ppu/ppu.go`

## Sprite Overflow

If more than 8 sprites are visible on a single scanline, the sprite overflow flag is set in PPUSTATUS.

**Note:** The real NES has a bug in sprite overflow detection that causes false positives/negatives. This emulator implements the simplified correct behavior.

**Implementation:** Set in `spriteEvaluation()` when count exceeds 8

## Timing

```
Scanline Cycles:
  0: Idle
  1-256: Render pixels (using data from previous scanline's fetching)
  257-320: Fetch sprite data for next scanline
  321-340: Background fetching for next scanline
```

## Code Organization

**pkg/ppu/ppu.go:**
- Main Clock() method orchestrates timing
- Calls sprite evaluation at cycle 257
- Calls sprite fetching at cycle 320
- Calls renderSprites() for each pixel
- Composites sprites with background

**pkg/ppu/sprites.go:**
- `spriteEvaluation()` - Scans OAM, populates secondary OAM
- `spriteFetching()` - Loads pattern data into shifters
- `renderSprites()` - Returns sprite pixel for current position
- `reverseByte()` - Helper for horizontal flipping

**Sprite State (in PPU struct):**
```go
secondaryOAM [32]uint8           // 8 sprites for current scanline
spriteCount uint8                // How many sprites (0-8)
sprite0Present bool              // Is sprite 0 on this scanline?
spriteShifterPatternLo [8]uint8  // Pattern data for 8 sprites
spriteShifterPatternHi [8]uint8
spriteAttributes [8]uint8        // Palette, priority, flip flags
spritePositions [8]uint8         // X positions
```

## Limitations

Current implementation:

**✅ Implemented:**
- 8x8 sprite rendering
- 8x16 sprite rendering
- Horizontal and vertical flipping
- Sprite priority (front/back of background)
- Sprite 0 hit detection
- Sprite overflow detection
- Proper sprite/background compositing

**⚠️ Not Implemented:**
- Sprite overflow hardware bug emulation
- Precise cycle-accurate sprite evaluation

## Testing

To verify sprite rendering works:

```bash
cd /home/andrew/Projects/nes-emulator/cmd/sdl-display
./nes-sdl ../../roms/donkeykong.nes
```

Expected results:
- Mario character should be visible
- Barrels should be visible
- Donkey Kong character should be visible
- All sprites should appear with correct colors
- Sprites should appear in front of or behind background as appropriate

---

**Implementation by Claude Code (Anthropic)**
**Date: 2025-11-04**
