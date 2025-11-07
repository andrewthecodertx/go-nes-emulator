# NES Mapper Implementation Guide

This document describes the cartridge mappers implemented in this emulator.

## Overview

NES cartridges use memory mappers to extend the limited address space of the NES through bank switching. Different mapper chips provide different capabilities, from simple fixed memory layouts to complex bank switching with IRQ generation.

## Supported Mappers

The emulator currently supports 5 mappers, covering approximately **70% of all NES games**:

| Mapper | Name | Coverage | Implementation File |
|--------|------|----------|---------------------|
| 0 | NROM | ~10% | `pkg/cartridge/mapper0.go` |
| 1 | MMC1 | ~28% | `pkg/cartridge/mapper1.go` |
| 2 | UxROM | ~11% | `pkg/cartridge/mapper2.go` |
| 3 | CNROM | ~7% | `pkg/cartridge/mapper3.go` |
| 4 | MMC3 | ~23% | `pkg/cartridge/mapper4.go` |

## Mapper Details

### Mapper 0 (NROM)

**Complexity:** Simple
**Games:** Super Mario Bros., Donkey Kong, Ice Climber
**Features:**
- No bank switching
- 16KB or 32KB PRG-ROM
- 8KB CHR-ROM or CHR-RAM

**Memory Map:**
- `$8000-$BFFF`: First 16KB of PRG-ROM
- `$C000-$FFFF`: Last 16KB of PRG-ROM (mirrors first 16KB if only one bank)
- `$0000-$1FFF`: 8KB CHR-ROM/RAM

### Mapper 1 (MMC1)

**Complexity:** Medium
**Games:** The Legend of Zelda, Metroid, Mega Man 2, Kid Icarus, Tetris
**Features:**
- 5-bit shift register control interface
- Switchable 16KB or 32KB PRG banks
- Switchable 4KB or 8KB CHR banks
- 8KB PRG-RAM at $6000-$7FFF
- Dynamic mirroring control

**Control Mechanism:**
Write to $8000-$FFFF 5 times with bit 0 shifting into internal register:
1. Reset if bit 7 is set
2. After 5 writes, value is copied to selected register

**Registers:**
- `$8000-$9FFF`: Control (mirroring, PRG/CHR mode)
- `$A000-$BFFF`: CHR bank 0
- `$C000-$DFFF`: CHR bank 1
- `$E000-$FFFF`: PRG bank

**PRG Modes:**
- Mode 0/1: 32KB mode (switch 32KB at once)
- Mode 2: Fix first bank, switch $C000-$FFFF
- Mode 3: Switch $8000-$BFFF, fix last bank

**CHR Modes:**
- Mode 0: 8KB mode (switch 8KB at once)
- Mode 1: 4KB mode (switch two 4KB banks independently)

### Mapper 2 (UxROM)

**Complexity:** Simple
**Games:** Mega Man, Castlevania, Duck Tales, Contra
**Features:**
- Switchable 16KB PRG banks
- Fixed 8KB CHR-RAM (no CHR-ROM)
- Last PRG bank fixed at $C000

**Bank Switching:**
Write to $8000-$FFFF to select PRG bank at $8000-$BFFF

**Memory Map:**
- `$8000-$BFFF`: Switchable 16KB PRG bank
- `$C000-$FFFF`: Fixed 16KB PRG bank (last bank)
- `$0000-$1FFF`: 8KB CHR-RAM

### Mapper 3 (CNROM)

**Complexity:** Simple
**Games:** Arkanoid, Cybernoid, Solomon's Key
**Features:**
- Fixed PRG-ROM (no bank switching)
- Switchable 8KB CHR banks
- Up to 4 CHR banks (32KB total)

**Bank Switching:**
Write to $8000-$FFFF to select CHR bank

**Memory Map:**
- `$8000-$BFFF`: First 16KB of PRG-ROM
- `$C000-$FFFF`: Last 16KB of PRG-ROM (mirrors first if only one bank)
- `$0000-$1FFF`: Switchable 8KB CHR bank

### Mapper 4 (MMC3)

**Complexity:** Complex
**Games:** Super Mario Bros. 2, Super Mario Bros. 3, Mega Man 3-6, Kirby's Adventure
**Features:**
- 8 bank registers (R0-R7)
- 2x switchable 8KB PRG banks
- 2x 2KB + 4x 1KB CHR banks
- Scanline-based IRQ counter
- 8KB PRG-RAM at $6000 with write protection
- Dynamic mirroring control

**Bank Registers:**
- R0-R1: 2KB CHR banks
- R2-R5: 1KB CHR banks
- R6-R7: 8KB PRG banks

**Control Registers:**
- `$8000-$9FFE` (even): Bank select
- `$8001-$9FFF` (odd): Bank data
- `$A000-$BFFE` (even): Mirroring
- `$A001-$BFFF` (odd): PRG-RAM protect
- `$C000-$DFFE` (even): IRQ latch
- `$C001-$DFFF` (odd): IRQ reload
- `$E000-$FFFE` (even): IRQ disable
- `$E001-$FFFF` (odd): IRQ enable

**IRQ Counter:**
The MMC3 has a scanline counter used for split-screen effects:
- Counter decrements each scanline (when rendering is enabled)
- When counter reaches 0, an IRQ is triggered if enabled
- Games use this for status bars, split-screen effects, etc.

**PRG Modes:**
- Mode 0: R6 at $8000, fixed at $C000
- Mode 1: Fixed at $8000, R6 at $C000

**CHR Modes:**
- Mode 0: 2KB banks at $0000, 1KB banks at $1000
- Mode 1: 2KB banks at $1000, 1KB banks at $0000

## Testing Mappers

Use the `cmd/test-mappers` tool to verify all mappers load correctly:

```bash
go run cmd/test-mappers/main.go
```

To check which mapper a ROM uses:

```bash
go run cmd/rom-info/main.go roms/game.nes
```

## Adding a New Mapper

To implement a new mapper:

1. **Create mapper file**: `pkg/cartridge/mapperX.go`

```go
package cartridge

type MapperX struct {
    prgROM []uint8
    chrMem []uint8
    // ... state fields
}

func NewMapperX(prgROM, chrROM []uint8, mirroring uint8) *MapperX {
    // Initialize mapper
}

func (m *MapperX) ReadPRG(addr uint16) uint8 {
    // Implement PRG read with bank switching
}

func (m *MapperX) WritePRG(addr uint16, value uint8) {
    // Implement bank selection/control
}

func (m *MapperX) ReadCHR(addr uint16) uint8 {
    // Implement CHR read with bank switching
}

func (m *MapperX) WriteCHR(addr uint16, value uint8) {
    // Implement CHR-RAM writes if applicable
}

func (m *MapperX) Scanline() {
    // Implement IRQ counter if needed
}

func (m *MapperX) GetMirroring() uint8 {
    // Return current mirroring mode
}
```

2. **Register mapper** in `pkg/cartridge/cartridge.go`:

```go
func createMapper(mapperID uint8, prgROM, chrROM []byte, mirroring uint8) (Mapper, error) {
    switch mapperID {
    // ... existing cases
    case X:
        return NewMapperX(prgROM, chrROM, mirroring), nil
    }
}
```

3. **Test with real ROM** that uses the mapper

## Reference Documentation

- [NESDev Wiki - Mapper List](https://www.nesdev.org/wiki/Mapper)
- [NESDev Wiki - iNES Format](https://www.nesdev.org/wiki/INES)
- Individual mapper pages:
  - [NROM](https://www.nesdev.org/wiki/NROM)
  - [MMC1](https://www.nesdev.org/wiki/MMC1)
  - [UxROM](https://www.nesdev.org/wiki/UxROM)
  - [CNROM](https://www.nesdev.org/wiki/CNROM)
  - [MMC3](https://www.nesdev.org/wiki/MMC3)

## Common Mapper Issues

### MMC1 Shift Register

The MMC1 requires 5 consecutive writes to configure. Make sure:
- Write bit 7 = 1 resets the shift register
- After reset, PRG mode is set to 3 (fix last bank)
- Each write shifts bit 0 into the register

### MMC3 IRQ Timing

The MMC3 IRQ counter must be clocked on PPU scanlines:
- Counter decrements when PPU is rendering (PPUMASK has rendering enabled)
- Counter reloads when it reaches 0 OR when reload flag is set
- IRQ is triggered when counter reaches 0 AND IRQ is enabled

### Bank Masking

When selecting banks, always mask to valid bank numbers:
```go
bank := value & (m.numBanks - 1)
```

This prevents accessing memory outside the ROM data.

## Future Mappers

Priority mappers for future implementation:
- **Mapper 7 (AxROM)**: Battletoads, Wizards & Warriors
- **Mapper 11 (Color Dreams)**: Simple bank switching
- **Mapper 66**: Simple PRG+CHR switching
- **Mapper 71**: PRG switching only
- **Mapper 5 (MMC5)**: Complex, used by Castlevania III

---

Generated with assistance from Claude Code (Anthropic)
