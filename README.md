# NES Emulator

A Nintendo Entertainment System (NES) emulator written in Go.

## Prerequisites

- Go 1.21 or later
- SDL2 development libraries:
  - Ubuntu/Debian: `sudo apt-get install libsdl2-dev`
  - macOS: `brew install sdl2`
  - Windows: Download from [libsdl.org](https://www.libsdl.org/)

## Building

```bash
# Clone and enter the repository
git clone https://github.com/andrewthecodertx/go-nes-emulator.git
cd go-nes-emulator

# Build everything
make
```

Or build manually:
```bash
go mod download
go build -o nes-sdl ./cmd/sdl-display
```

## Running

```bash
./nes-sdl path/to/game.nes
```

## Controls

| Key | Action |
|-----|--------|
| Arrow Keys | D-Pad |
| X | A Button |
| Z | B Button |
| Enter | Start |
| Right Shift | Select |
| ESC | Quit |
| P | Pause/Resume |
| R | Reset |

## Supported Mappers

The emulator supports ~72% of NES games through these mappers:

- Mapper 0 (NROM) - Super Mario Bros., Donkey Kong
- Mapper 1 (MMC1) - Zelda, Metroid
- Mapper 2 (UxROM) - Mega Man, Castlevania
- Mapper 3 (CNROM) - Arkanoid
- Mapper 4 (MMC3) - Super Mario Bros. 3, Mega Man 3-6
- Mapper 7 (AxROM) - Battletoads

## License

This project is licensed under the [MIT License](LICENSE).
