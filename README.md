# NES Emulator

A Nintendo Entertainment System (NES) emulator written in Go.

## Installation

Download the latest release for your platform from the [Releases](https://github.com/andrewthecodertx/go-nes-emulator/releases) page.

**Linux:**
```bash
tar -xzf nes-emulator-linux-amd64.tar.gz
./nes-emulator path/to/game.nes
```

**macOS:**
```bash
tar -xzf nes-emulator-macos-arm64.tar.gz
./nes-emulator path/to/game.nes
```

**Arch Linux (AUR):**
```bash
yay -S nes-emulator-git
```

## Building from Source

### Prerequisites

- Go 1.21 or later
- SDL2 development libraries:
  - Ubuntu/Debian: `sudo apt-get install libsdl2-dev`
  - Arch Linux: `sudo pacman -S sdl2`
  - macOS: `brew install sdl2`

### Build

```bash
git clone https://github.com/andrewthecodertx/go-nes-emulator.git
cd go-nes-emulator
make
```

## ROMs

This emulator does not include any games. You must provide your own ROM files (`.nes` format).

Due to copyright restrictions, ROMs cannot be distributed with this software. You can find NES ROMs at [Vimm's Lair](https://vimm.net/vault/NES).

## Running

```bash
./nes-emulator path/to/game.nes
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

## Limitations

This emulator is a work in progress. Current limitations include:

- **No audio** - APU (Audio Processing Unit) is not implemented
- **Single player only** - No support for a second controller
- **Limited mapper support** - Only 6 of 200+ mappers are implemented; games using unsupported mappers will not load
- **No save states** - Cannot save or load emulator state
- **No battery-backed saves** - Games with save functionality (Zelda, Final Fantasy) will not persist saves between sessions

## License

MIT, see [LICENSE](LICENSE).

## Contributing

PRs welcome. Please open an issue first for major changes.
