# NES Emulator

A Nintendo Entertainment System (NES) emulator written in Go, built on top of
the [go-6502-emulator](https://github.com/andrewthecodertx/go-6502-emulator) library.

## Features

### Currently Implemented

- **6502 CPU**: Full cycle-accurate MOS 6502 emulation via go-6502-emulator
- **PPU (Picture Processing Unit)**:
  - Complete PPU register interface ($2000-$2007)
  - Nametable memory with mirroring support (horizontal, vertical,
  single-screen, four-screen)
  - Palette RAM with 64-color hardware palette
  - OAM (Object Attribute Memory) for sprites
  - Internal scrolling registers (Loopy registers) with hardware-accurate scroll
  operations
  - VBlank NMI generation
  - **Cycle-accurate background rendering** (fully functional!)
    - 8-cycle tile fetching pipeline per tile
    - 16-bit shift registers for pixel output
    - Proper scrolling with coarse/fine X/Y
    - Attribute table support for palette selection
    - Pre-render scanline with vertical scroll restoration
    - See [RENDERING.md](RENDERING.md) for implementation details
  - **Sprite rendering** with proper priority and transparency
    - OAM evaluation (8 sprites per scanline limit)
    - Sprite 0 hit detection
    - 8x8 and 8x16 sprite modes
    - See [SPRITE-RENDERING.md](SPRITE-RENDERING.md) for details
- **Cartridge Support**:
  - iNES ROM format (.nes files)
  - **6 mappers implemented** (~72% of NES games):
    - Mapper 0 (NROM): ~10% of games
    - Mapper 1 (MMC1): ~28% of games (Zelda, Metroid, Tetris)
    - Mapper 2 (UxROM): ~11% of games (Mega Man, Castlevania, Contra)
    - Mapper 3 (CNROM): ~7% of games (Arkanoid, Cybernoid)
    - Mapper 4 (MMC3): ~23% of games (Super Mario Bros. 3, Mega Man 3-6)
    - Mapper 7 (AxROM): ~2% of games (Battletoads, Marble Madness)
  - Automatic mapper selection from ROM header
- **Controller Input**:
  - Full NES controller emulation (8 buttons)
  - Proper strobe/latch mechanism
  - Support for 2 controllers
- **SDL2 Display Frontend**:
  - Real-time rendering at 60 FPS
  - Keyboard controls for NES buttons
  - Debug output and frame stepping
  - Located in `cmd/sdl-display/`
- **Memory System**:
  - 2KB CPU RAM with mirroring
  - Complete NES memory map
  - PPU memory map with CHR-ROM/RAM access
  - OAM DMA support
- **System Bus**: Connects CPU, PPU, RAM, cartridge, and controllers

## Architecture

```
nes-emulator/
├── cmd/
│   └── nes-emulator/    # Main executable (TODO)
├── pkg/
│   ├── nes/             # Main NES emulator
│   ├── ppu/             # Picture Processing Unit
│   ├── cartridge/       # ROM loading and mappers
│   └── bus/             # System bus
├── roms/                # Place your .nes ROM files here
└── README.md
```

## Installation

### Prerequisites

- Go 1.25 or later
- SDL2 development libraries (for the display frontend)
  - Ubuntu/Debian: `sudo apt-get install libsdl2-dev`
  - macOS: `brew install sdl2`
  - Windows: Download from [libsdl.org](https://www.libsdl.org/)
- The go-6502-emulator library (pulled automatically via Go modules)

### Building

```bash
# Clone the repository
git clone https://github.com/andrewthecodertx/nes-emulator.git
cd nes-emulator

# Download dependencies
go mod download

# Build the SDL display frontend (takes 1-2 minutes due to CGO)
cd cmd/sdl-display
go build -o nes-sdl

# Run with a ROM file
./nes-sdl ../../roms/nestest.nes
```

### Controls

When running the SDL frontend:

**System Controls:**

- ESC: Quit
- P: Pause/Resume
- SPACE: Step one frame (when paused)
- R: Reset emulator
- D: Toggle debug output
- F: Toggle forced rendering

**NES Controller:**

- Arrow keys: D-Pad
- X: A button
- Z: B button
- Enter: START
- Right Shift: SELECT

## Usage

### As a Library

```go
package main

import (
    "log"
    "github.com/andrewthecodertx/nes-emulator/pkg/nes"
)

func main() {
    // Load a ROM file
    emulator, err := nes.New("roms/game.nes")
    if err != nil {
        log.Fatal(err)
    }

    // Reset to power-on state
    emulator.Reset()

    // Run one frame (approximately 29,781 CPU cycles)
    emulator.RunFrame()

    // Get frame buffer for rendering
    frameBuffer := emulator.GetFrameBuffer()
    // frameBuffer is 256x240 pixels, each byte is a palette index

    // Access CPU state directly
    cpu := emulator.GetCPU()
    println("Program Counter:", cpu.PC)

    // Access PPU directly
    ppu := emulator.GetPPU()
    // ...use PPU methods...
}
```

### Step-by-Step Execution

```go
// Execute one CPU instruction at a time
for i := 0; i < 1000; i++ {
    emulator.Step()
}

// Get total cycles executed
totalCycles := emulator.GetCycles()
```

### Controller Input

```go
// Controller button layout (bit positions):
//   Bit 7: A
//   Bit 6: B
//   Bit 5: Select
//   Bit 4: Start
//   Bit 3: Up
//   Bit 2: Down
//   Bit 1: Left
//   Bit 0: Right

const (
    ButtonA      = 0x80
    ButtonB      = 0x40
    ButtonSelect = 0x20
    ButtonStart  = 0x10
    ButtonUp     = 0x08
    ButtonDown   = 0x04
    ButtonLeft   = 0x02
    ButtonRight  = 0x01
)

// Press A and B buttons
emulator.SetControllerState(0, ButtonA | ButtonB)

// Press nothing
emulator.SetControllerState(0, 0x00)
```

## NES System Overview

### CPU (MOS 6502)

The NES uses an NMOS 6502 CPU running at approximately 1.79 MHz. This emulator
uses the go-6502-emulator library for accurate CPU emulation.

**Memory Map:**

- `$0000-$07FF`: 2KB internal RAM
- `$0800-$1FFF`: Mirrors of RAM
- `$2000-$2007`: PPU registers
- `$2008-$3FFF`: Mirrors of PPU registers
- `$4000-$4017`: APU and I/O registers
- `$4020-$FFFF`: Cartridge space (PRG-ROM, PRG-RAM, mapper registers)

### PPU (Picture Processing Unit)

The PPU generates the video signal, rendering 256x240 pixels at ~60 Hz (NTSC).
This implementation is **cycle-accurate**, emulating every single PPU cycle as
specified in the NES hardware.

**Rendering Pipeline:**

- **Pre-render scanline (-1):** Prepares for next frame, restores vertical scroll
- **Visible scanlines (0-239):** Outputs pixels, fetches tiles in 8-cycle pattern
- **Post-render scanline (240):** Idle
- **VBlank scanlines (241-260):** Game performs work, NMI triggered

**Features:**

- 2KB internal VRAM for nametables
- 256 bytes of OAM for 64 sprites (8 sprites per scanline limit)
- 32 bytes of palette RAM
- Hardware-accurate scrolling via Loopy registers (vramAddress, tempVRAMAddress,
fineX)
- Configurable nametable mirroring (horizontal, vertical, single-screen,
four-screen)
- Sprite 0 hit detection
- 8x8 and 8x16 sprite modes

**Memory Map:**

- `$0000-$0FFF`: Pattern Table 0 (4KB, CHR-ROM/RAM)
- `$1000-$1FFF`: Pattern Table 1 (4KB, CHR-ROM/RAM)
- `$2000-$2FFF`: Nametables (4KB, with mirroring)
- `$3F00-$3F1F`: Palette RAM (32 bytes)

**Registers (CPU-visible at $2000-$2007):**

- `$2000`: PPUCTRL - PPU control flags (NMI enable, sprite size, pattern table
select, etc.)
- `$2001`: PPUMASK - Rendering control (show background, show sprites, emphasis
bits)
- `$2002`: PPUSTATUS - PPU status (VBlank, Sprite0 hit, sprite overflow)
- `$2003`: OAMADDR - OAM address
- `$2004`: OAMDATA - OAM data port
- `$2005`: PPUSCROLL - Scroll position (write twice: X then Y)
- `$2006`: PPUADDR - VRAM address (write twice: high byte then low byte)
- `$2007`: PPUDATA - VRAM data port

**For detailed PPU architecture information, see:**

- [RENDERING.md](RENDERING.md) - Background rendering pipeline
- [SPRITE-RENDERING.md](SPRITE-RENDERING.md) - Sprite rendering details
- [NESDev PPU Reference](https://www.nesdev.org/wiki/PPU) - Official documentation

### Cartridge Mappers

NES games use memory mappers to extend the limited address space through bank
switching.

**Supported Mappers (~72% of NES games):**

- **Mapper 0 (NROM)**: No bank switching (~10% of games)
  - 16KB or 32KB PRG-ROM
  - 8KB CHR-ROM or CHR-RAM
  - Games: Super Mario Bros., Donkey Kong, Ice Climber

- **Mapper 1 (MMC1)**: Complex shift register control (~28% of games)
  - 4KB/8KB CHR banking, 16KB/32KB PRG banking
  - 8KB PRG-RAM
  - Games: Zelda, Metroid, Tetris

- **Mapper 2 (UxROM)**: Simple PRG bank switching (~11% of games)
  - Switchable 16KB PRG banks, fixed last bank
  - 8KB CHR-RAM
  - Games: Mega Man, Castlevania, Contra

- **Mapper 3 (CNROM)**: Simple CHR bank switching (~7% of games)
  - Fixed PRG, switchable CHR
  - Games: Arkanoid, Cybernoid

- **Mapper 4 (MMC3)**: Complex with IRQ counter (~23% of games)
  - Configurable PRG/CHR banking
  - Scanline IRQ for split-screen effects
  - 8KB PRG-RAM with write protection
  - Games: Super Mario Bros. 3, Mega Man 3-6

- **Mapper 7 (AxROM)**: 32KB PRG bank switching (~2% of games)
  - Single-screen mirroring control
  - 8KB CHR-RAM
  - Games: Battletoads, Marble Madness

## ROM Format

This emulator supports iNES format (.nes files).

**iNES Header (16 bytes):**

```
0-3: "NES" followed by MS-DOS EOF (0x1A)
4:   Number of 16KB PRG-ROM banks
5:   Number of 8KB CHR-ROM banks
6:   Flags 6 (Mapper low, mirroring, battery, trainer)
7:   Flags 7 (Mapper high, VS System, PlayChoice-10)
8-15: Unused (typically zero)
```

## Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/ppu
go test ./pkg/cartridge

# Run with verbose output
go test -v ./...
```

## Development

### Project Structure

- **pkg/nes**: Main emulator coordination
- **pkg/ppu**: PPU implementation
  - `ppu.go`: Core PPU logic
  - `registers.go`: PPU register types
- **pkg/cartridge**: ROM loading and mappers
  - `cartridge.go`: iNES format parser
  - `mapper.go`: Mapper interface
  - `mapper0.go`: NROM implementation
- **pkg/bus**: System bus connecting components

### Adding a New Mapper

1. Create `pkg/cartridge/mapperX.go`
2. Implement the `Mapper` interface
3. Add case to `createMapper()` in `cartridge.go`

Example:

```go
type Mapper1 struct {
    // ... mapper state ...
}

func NewMapper1(prgROM, chrROM []byte, mirroring uint8) *Mapper1 {
    // ... initialization ...
}

func (m *Mapper1) ReadPRG(addr uint16) uint8 { /* ... */ }
func (m *Mapper1) WritePRG(addr uint16, value uint8) { /* ... */ }
func (m *Mapper1) ReadCHR(addr uint16) uint8 { /* ... */ }
func (m *Mapper1) WriteCHR(addr uint16, value uint8) { /* ... */ }
func (m *Mapper1) Scanline() { /* ... */ }
func (m *Mapper1) GetMirroring() uint8 { /* ... */ }
```

## References

### NES Documentation

- [NESDev Wiki](https://www.nesdev.org/wiki/Nesdev_Wiki) - Comprehensive NES
documentation (primary reference)
- [6502 CPU Reference](http://www.6502.org/) - 6502 processor documentation
- [iNES Format](https://www.nesdev.org/wiki/INES) - ROM format specification

### PPU Architecture & Rendering

The PPU implementation in this emulator is based on extensive documentation from
the NES development community:

- The [NESDev Wiki](https://www.nesdev.org/) community's excellent documentation
- [PPU Reference](https://www.nesdev.org/wiki/PPU) - Complete PPU technical
details
- [PPU Rendering](https://www.nesdev.org/wiki/PPU_rendering) - Cycle-by-cycle
rendering timing
- [PPU Scrolling](https://www.nesdev.org/wiki/PPU_scrolling) - Loopy's
scrolling documentation
- [PPU Registers](https://www.nesdev.org/wiki/PPU_registers) - Register behavior
and side effects
- [PPU OAM](https://www.nesdev.org/wiki/PPU_OAM) - Sprite memory organization
- [PPU Palettes](https://www.nesdev.org/wiki/PPU_palettes) - Color palette
system
- [PPU Nametables](https://www.nesdev.org/wiki/PPU_nametables) - Nametable
addressing and mirroring
- [PPU Pattern Tables](https://www.nesdev.org/wiki/PPU_pattern_tables) - CHR-ROM
tile format

### Mappers

- [Mapper List](https://www.nesdev.org/wiki/Mapper) - Complete list of NES mappers
- [Mapper 0 (NROM)](https://www.nesdev.org/wiki/NROM)
- [Mapper 1 (MMC1)](https://www.nesdev.org/wiki/MMC1)
- [Mapper 2 (UxROM)](https://www.nesdev.org/wiki/UxROM)
- [Mapper 3 (CNROM)](https://www.nesdev.org/wiki/CNROM)
- [Mapper 4 (MMC3)](https://www.nesdev.org/wiki/MMC3)
- [Mapper 7 (AxROM)](https://www.nesdev.org/wiki/AxROM)

## Dependencies

- [go-6502-emulator](https://github.com/andrewthecodertx/go-6502-emulator) -
6502 CPU emulation

## License

This project is open source. See LICENSE file for details.

## Contributing

Contributions are welcome! Priority areas:

1. **APU (Audio Processing Unit)**: Sound generation and audio output
2. **Additional Mappers**: Mapper 5 (MMC5), Mapper 9/10 (MMC2/MMC4), Mapper 11, etc.
3. **Testing**: Integration with NES test ROMs (nestest, blargg's tests, etc.)
4. **Performance**: Optimization for lower-end hardware
5. **Tools & Debugging**:
   - Memory viewer/editor
   - CPU disassembler
   - Pattern table viewer
   - Nametable viewer
   - Audio waveform visualizer
6. **Platform Support**: Web frontend (WASM), mobile builds
7. **Features**: Save states, rewind, fast-forward, screenshots

---

**Note**: This is an educational emulator project. It aims for accuracy but may
not be suitable for all use cases. ROM files are not included - you must provide
your own legally obtained ROM files.
