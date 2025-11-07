# NES Emulator Integration Guide

This document explains how the NES emulator integrates with the go-6502-emulator library.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     NES Emulator                            │
│  (/home/andrew/Projects/nes-emulator)                       │
│                                                             │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  pkg/nes/nes.go                                      │   │
│  │  - Main NES struct                                   │   │
│  │  - Coordinates all components                        │   │
│  │  - Frame execution logic                             │   │
│  └─────────┬─────────────────────┬──────────────────────┘   │
│            │                     │                          │
│    ┌───────▼─────────┐   ┌──────▼────────┐                  │
│    │ pkg/ppu/        │   │ pkg/bus/      │                  │
│    │ - PPU           │   │ - NES Bus     │                  │
│    │ - Registers     │   │   (implements │                  │
│    │ - Frame buffer  │   │   core.Bus)   │                  │
│    └─────────────────┘   └───────┬───────┘                  │
│                                   │                         │
│                       ┌───────────▼──────────┐              │
│                       │ pkg/cartridge/       │              │
│                       │ - ROM loader         │              │
│                       │ - Mappers            │              │
│                       └──────────────────────┘              │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ uses (via go.mod replace)
                              │
┌─────────────────────────────▼─────────────────────────────────┐
│          go-6502-emulator                                     │
│  (/mnt/internalssd/Projects/go-6502)                          │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐     │
│  │  pkg/mos6502/cpu.go                                  │     │
│  │  - NMOS 6502 CPU emulation                           │     │
│  │  - Instruction execution                             │     │
│  │  - Interrupt handling                                │     │
│  └──────────────────────────────────────────────────────┘     │
│                                                               │
│  ┌──────────────────────────────────────────────────────┐     │
│  │  pkg/core/bus.go                                     │     │
│  │  - Bus interface definition                          │     │
│  │  - Read(addr uint16) uint8                           │     │
│  │  - Write(addr uint16, data uint8)                    │     │
│  └──────────────────────────────────────────────────────┘     │
└───────────────────────────────────────────────────────────────┘
```

## Module Dependency Setup

### Current Setup (Development)

In `nes-emulator/go.mod`:

```go
module github.com/andrewthecodertx/nes-emulator

go 1.25

require github.com/andrewthecodertx/go-6502-emulator v0.0.0

// Local development: use local copy
replace github.com/andrewthecodertx/go-6502-emulator => /mnt/internalssd/Projects/go-6502
```

### Future Setup (Production)

Once you tag and publish go-6502-emulator to GitHub:

```bash
# In go-6502 directory
cd /mnt/internalssd/Projects/go-6502
git tag v0.1.0
git push origin v0.1.0
```

Then in `nes-emulator/go.mod`, change to:

```go
module github.com/andrewthecodertx/nes-emulator

go 1.25

require github.com/andrewthecodertx/go-6502-emulator v0.1.0

// Remove the replace directive
```

## Key Integration Points

### 1. CPU Bus Interface

The NES bus (`pkg/bus/bus.go`) implements the `core.Bus` interface from go-6502-emulator:

```go
type NESBus struct {
    cpuRAM [2048]uint8
    ppu    *ppu.PPU
    mapper cartridge.Mapper
    // ...
}

// Implements core.Bus interface
func (b *NESBus) Read(addr uint16) uint8 { /* ... */ }
func (b *NESBus) Write(addr uint16, data uint8) { /* ... */ }
```

This allows the 6502 CPU to access:

- 2KB CPU RAM ($0000-$1FFF, mirrored)
- PPU registers ($2000-$3FFF, mirrored)
- Cartridge PRG-ROM ($8000-$FFFF)

### 2. CPU Creation

In `pkg/nes/nes.go`:

```go
import "github.com/andrewthecodertx/go-6502-emulator/pkg/mos6502"

// Create CPU with NES bus
cpu := mos6502.NewCPU(nesbus)
```

The CPU automatically uses the bus for all memory operations.

### 3. Execution Loop

```go
func (n *NES) Step() uint8 {
    // Execute one CPU instruction
    n.cpu.Step()

    // Clock PPU at 3x CPU speed
    n.bus.Clock() // Internally calls ppu.Clock() 3 times

    // Handle NMI from PPU
    if n.bus.IsNMI() {
        n.cpu.NMIPending = true
    }

    return n.cpu.GetCycles()
}
```

## Building and Testing

### Build Everything

```bash
# Build go-6502 library
cd /mnt/internalssd/Projects/go-6502
go build ./...
go test ./...

# Build NES emulator
cd /home/andrew/Projects/nes-emulator
go build ./...

# Build example
cd examples/basic
go build
```

### Using Go Workspace (Alternative)

Instead of `replace` directive, you can use Go workspaces:

```bash
cd /home/andrew/Projects
go work init
go work use ./go-6502 ./nes-emulator
```

This is cleaner for local development with multiple modules.

## Component Details

### PPU (pkg/ppu/)

**Files:**

- `registers.go`: PPU register types (PPUControl, PPUMask, PPUStatus, LoopyRegister)
- `ppu.go`: Main PPU implementation

**Key Features:**

- CPU register interface ($2000-$2007)
- 2KB nametable RAM
- 32 bytes palette RAM
- 256 bytes OAM
- Mirroring support
- VBlank NMI generation

**TODO:**

- Background tile rendering
- Sprite rendering
- Sprite 0 hit detection

### Cartridge (pkg/cartridge/)

**Files:**

- `cartridge.go`: iNES ROM loader
- `mapper.go`: Mapper interface
- `mapper0.go`: NROM implementation

**Key Features:**

- iNES format parsing
- PRG-ROM/CHR-ROM loading
- Mapper abstraction

**TODO:**

- More mappers (MMC1, MMC3, etc.)
- Save RAM support

### Bus (pkg/bus/)

**Files:**

- `bus.go`: System bus connecting all components

**Key Features:**

- CPU RAM with mirroring
- PPU register mapping
- Cartridge ROM mapping
- DMA transfer (for OAM)
- Controller input (basic)

## Example Usage

See `examples/basic/main.go` for a complete example:

```go
// Load ROM
emulator, err := nes.New("roms/game.nes")
if err != nil {
    log.Fatal(err)
}

// Reset
emulator.Reset()

// Run frames
for i := 0; i < 60; i++ { // 1 second at 60 FPS
    emulator.RunFrame()
}

// Get frame buffer for rendering
frameBuffer := emulator.GetFrameBuffer()
```

## Next Steps

1. **Publish go-6502-emulator**: Tag and push to GitHub
2. **Implement PPU Rendering**: Add background/sprite rendering to PPU
3. **Add More Mappers**: Implement MMC1, MMC3, etc.
4. **Create Frontend**: Add SDL2/OpenGL frontend for display
5. **Add APU**: Implement sound generation
6. **Testing**: Use nestest.nes and other test ROMs

## Troubleshooting

### Module not found errors

If you see errors about the go-6502-emulator module:

1. Check the `replace` directive points to correct path
2. Verify go.mod in go-6502 has correct module name
3. Try `go mod tidy` in nes-emulator directory

### Import errors

Make sure all imports use the correct module name:

```go
import "github.com/andrewthecodertx/go-6502-emulator/pkg/mos6502"
```

Not:

```go
import "github.com/andrewthecodertx/go-6502-emulator/pkg/mos6502"  // WRONG
```

---

Generated with assistance from Claude Code (Anthropic)
