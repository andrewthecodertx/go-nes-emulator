# SDL Display Debug Output Improvements

## Summary

Improved the SDL display frontend (`cmd/sdl-display/main.go`) to have cleaner, more informative output with less spam.

## Changes Made

### 1. Cleaner Startup Messages

**Before:**
```
Loading ROM: roms/tetris.nes
NES initialized
Running initialization frames (letting game start up)...
Starting main loop
System Controls: ESC=quit, P=pause, SPACE=step frame, R=reset, F=toggle forced rendering
Game Controls: Arrow keys=D-pad, Z=B, X=A, Enter=Start, RShift=Select
Note: PPU rendering must be enabled by the game (or press F to force)
```

**After:**
```
=== Loading ROM ===
File: roms/tetris.nes
Mapper: 1
PRG Banks: 2 x 16KB = 32KB
CHR Banks: 2 x 8KB = 16KB

Initializing (2 seconds)...

=== NES Emulator Ready ===
System: ESC=quit | P=pause | SPACE=step | R=reset | F=force render | D=debug
Game:   Arrows=D-pad | Z=B | X=A | Enter=Start | RShift=Select
==========================
```

**Improvements:**
- Shows cartridge information (mapper, PRG/CHR banks)
- More organized with clear sections
- Compact control listing (uses `|` separators)
- Added `D=debug` key to toggle debug output

### 2. Debug Output Control

**Before:**
- Debug output always on
- Verbose frame buffer dumps every 60 frames
- Confusing pixel analysis with ARGB checks
- Special frame 300 debugging

**After:**
- Debug output **off by default**
- Press `D` key to toggle debug on/off
- When off: Shows minimal status every 5 seconds
- When on: Shows useful per-frame statistics every second

**Debug Off (default):**
```
[Frame 300] Running... (press D for debug info)
[Frame 600] Running... (press D for debug info)
```

**Debug On (press D):**
```
[Frame  60] Colors: 3 unique | Most common: $0F (58234 pixels)
[Frame 120] Colors: 5 unique | Most common: $0F (52018 pixels)
[Frame 180] Colors: 8 unique | Most common: $21 (34567 pixels)
```

### 3. Informative Statistics

Instead of dumping raw palette indices and checking for specific colors (magenta/cyan), the new output shows:

- **Unique colors**: How many different palette indices are being used
- **Most common color**: Which palette index appears most frequently
- **Pixel count**: How many pixels use that color

This helps identify issues:
- All black screen: `Colors: 1 unique | Most common: $0F`
- Active rendering: `Colors: 8+ unique`
- Rendering enabled: Various colors, changing over time

### 4. Added GetCartridge() Method

Added to `pkg/nes/nes.go`:
```go
// GetCartridge returns a pointer to the loaded cartridge
func (n *NES) GetCartridge() *cartridge.Cartridge {
    return n.cartridge
}
```

This allows the frontend to display ROM information at startup.

## Usage

### Normal Usage (Minimal Output)
```bash
./nes-sdl roms/game.nes
```

You'll see:
- ROM info at startup
- Control guide
- Status update every 5 seconds

### Debug Mode
While running, press **D** to enable debug output.

You'll see every second:
- Frame number
- Number of unique colors being rendered
- Most common color and its usage

Press **D** again to disable.

### Other Controls
- **ESC**: Quit
- **P**: Pause/Resume
- **SPACE**: Step one frame (when paused)
- **R**: Reset emulator
- **F**: Force rendering on/off
- **D**: Toggle debug info

## Benefits

1. **Less Clutter**: No spam during normal operation
2. **More Informative**: Shows actual cartridge details
3. **User Control**: Debug info available on demand (D key)
4. **Better Diagnostics**: Color statistics help identify rendering issues
5. **Professional**: Clean, organized output

## Example Output

### Loading Donkey Kong (Mapper 0)
```
=== Loading ROM ===
File: roms/donkeykong.nes
Mapper: 0
PRG Banks: 1 x 16KB = 16KB
CHR Banks: 1 x 8KB = 8KB

Initializing (2 seconds)...

=== NES Emulator Ready ===
System: ESC=quit | P=pause | SPACE=step | R=reset | F=force render | D=debug
Game:   Arrows=D-pad | Z=B | X=A | Enter=Start | RShift=Select
==========================

[Frame 300] Running... (press D for debug info)
```

### Loading Tetris (Mapper 1)
```
=== Loading ROM ===
File: roms/tetris.nes
Mapper: 1
PRG Banks: 2 x 16KB = 32KB
CHR Banks: 2 x 8KB = 16KB

Initializing (2 seconds)...

=== NES Emulator Ready ===
System: ESC=quit | P=pause | SPACE=step | R=reset | F=force render | D=debug
Game:   Arrows=D-pad | Z=B | X=A | Enter=Start | RShift=Select
==========================

[Frame 300] Running... (press D for debug info)
```

---

Generated with assistance from Claude Code (Anthropic)
