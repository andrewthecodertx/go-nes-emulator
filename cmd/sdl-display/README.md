# NES SDL2 Display Frontend

Simple SDL2-based graphical frontend for the NES emulator.

## Prerequisites

### Install SDL2

**Ubuntu/Debian:**
```bash
sudo apt-get install libsdl2-dev
```

**macOS:**
```bash
brew install sdl2
```

**Arch Linux:**
```bash
sudo pacman -S sdl2
```

## Building

```bash
cd /home/andrew/Projects/nes-emulator/cmd/sdl-display
go build -o nes-sdl
```

## Running

```bash
# From this directory
./nes-sdl ../../roms/donkeykong.nes

# Or with any ROM
./nes-sdl /path/to/your/rom.nes
```

## Controls

- **ESC**: Quit
- **P**: Pause/Resume emulation
- **SPACE**: Step one frame (when paused)
- **R**: Reset emulator
- **F**: Toggle forced rendering (debug mode)

## Features

- 256x240 display scaled 3x (768x720 window)
- ~60 FPS emulation
- Palette index → RGB conversion
- Background tile rendering

## Troubleshooting

**"Failed to initialize SDL":**
- Make sure SDL2 is installed (see prerequisites)

**Black screen:**
- ROM may not be enabling PPU rendering
- Try different ROM files
- Check console output for errors

**Build errors:**
- Ensure go-sdl2 is installed: `go get github.com/veandco/go-sdl2/sdl`
- Ensure SDL2 development libraries are installed

## Technical Details

The frontend:
1. Loads NES ROM using the nes-emulator library
2. Runs emulation at ~60 FPS
3. Converts PPU frame buffer (palette indices) to RGB
4. Uploads to SDL texture and displays

Frame buffer format:
- Input: 256x240 uint8 array (palette indices 0-63)
- Output: 256x240x3 RGB24 (using hardware palette lookup)

---

Generated with assistance from Claude Code (Anthropic)
