// Package ppu implements the NES Picture Processing Unit (2C02).
//
// The PPU is the graphics processor for the NES. It generates video signals
// at 256x240 resolution by rendering background tiles and sprites.
//
// Hardware Specifications:
//   - Clock speed: ~5.37 MHz (NTSC) / ~5.32 MHz (PAL)
//   - Runs 3x faster than CPU (~1.79 MHz)
//   - 341 PPU cycles per scanline
//   - 262 scanlines per frame (NTSC) / 312 (PAL)
//   - Output: 256 pixels wide x 240 pixels tall
//
// Memory Map:
//   - $0000-$0FFF: Pattern Table 0 (4KB, CHR-ROM/RAM)
//   - $1000-$1FFF: Pattern Table 1 (4KB, CHR-ROM/RAM)
//   - $2000-$23FF: Nametable 0 (1KB)
//   - $2400-$27FF: Nametable 1 (1KB)
//   - $2800-$2BFF: Nametable 2 (1KB)
//   - $2C00-$2FFF: Nametable 3 (1KB)
//   - $3000-$3EFF: Mirrors of $2000-$2EFF
//   - $3F00-$3F1F: Palette RAM (32 bytes)
//   - $3F20-$3FFF: Mirrors of $3F00-$3F1F
package ppu

import "github.com/andrewthecodertx/nes-emulator/pkg/cartridge"

// Mirroring modes for nametables
const (
	MirrorHorizontal = 0 // Vertical arrangement
	MirrorVertical   = 1 // Horizontal arrangement
	MirrorSingleLow  = 2 // All nametables map to lower bank
	MirrorSingleHigh = 3 // All nametables map to upper bank
	MirrorFourScreen = 4 // Four separate nametables (requires extra RAM on cartridge)
)

// Screen dimensions
const (
	ScreenWidth  = 256
	ScreenHeight = 240
)

// Timing constants (NTSC)
const (
	CyclesPerScanline = 341
	ScanlinesPerFrame = 262
	VisibleScanlines  = 240
)

// PPU represents the NES Picture Processing Unit (2C02)
type PPU struct {
	// ========================================================================
	// Memory Banks
	// ========================================================================

	// Nametable RAM (2KB internal)
	// The NES has 2KB of internal VRAM for nametables. The full 4KB nametable
	// space ($2000-$2FFF) is mapped to this 2KB using mirroring modes.
	nametable [2048]uint8

	// Palette RAM (32 bytes)
	// $3F00-$3F0F: Background palettes (4 palettes x 4 colors)
	// $3F10-$3F1F: Sprite palettes (4 palettes x 4 colors)
	// Note: $3F10, $3F14, $3F18, $3F1C are mirrored to $3F00, $3F04, $3F08, $3F0C
	paletteRAM [32]uint8

	// Object Attribute Memory (256 bytes)
	// Contains sprite data for 64 sprites (4 bytes each):
	//   Byte 0: Y position (top of sprite)
	//   Byte 1: Tile index
	//   Byte 2: Attributes (palette, priority, flip flags)
	//   Byte 3: X position (left of sprite)
	oam [256]uint8

	// OAM Address register ($2003)
	// Points to current position in OAM for CPU read/write
	oamAddress uint8

	// ========================================================================
	// PPU Registers (CPU-visible at $2000-$2007)
	// ========================================================================

	control  PPUControl  // PPUCTRL ($2000) - Control Register
	mask     PPUMask     // PPUMASK ($2001) - Mask Register
	status   PPUStatus   // PPUSTATUS ($2002) - Status Register
	oamData  uint8       // OAMDATA ($2004) - OAM Data Port
	ppuScroll uint8      // PPUSCROLL ($2005) - Scroll Position Register (write x2)
	ppuAddr  uint8       // PPUADDR ($2006) - PPU Address Register (write x2)
	ppuData  uint8       // PPUDATA ($2007) - PPU Data Port

	// ========================================================================
	// Internal Registers (Loopy Registers)
	// ========================================================================

	// VRAM Address Register (current address the PPU will read/write)
	// Also known as "v" in Loopy's documentation
	vramAddress LoopyRegister

	// Temporary VRAM Address Register
	// Also used for scroll position. Known as "t" in Loopy's documentation
	tempVRAMAddress LoopyRegister

	// Fine X scroll (3 bits: 0-7)
	fineX uint8

	// Write latch/toggle (first or second write to $2005/$2006)
	writeLatch bool

	// Internal read buffer for PPUDATA reads
	// Reads from PPUDATA are buffered (delayed by one read)
	readBuffer uint8

	// ========================================================================
	// Rendering State
	// ========================================================================

	// Current scanline (0-261)
	scanline int16

	// Current cycle within scanline (0-340)
	cycle uint16

	// Frame counter
	frame uint64

	// Odd/even frame (affects timing on odd frames)
	oddFrame bool

	// Frame complete flag
	frameComplete bool

	// ========================================================================
	// Background Rendering State
	// ========================================================================

	// Next background tile ID from nametable
	bgNextTileID uint8

	// Next background tile attribute (palette selection, 2 bits)
	bgNextTileAttrib uint8

	// Next background tile pattern low byte
	bgNextTileLSB uint8

	// Next background tile pattern high byte
	bgNextTileMSB uint8

	// Background pattern shifters (16-bit)
	// Top 8 bits = current 8 pixels, bottom 8 bits = next 8 pixels
	// Shifts left by 1 each cycle to output one pixel
	bgShifterPatternLo uint16
	bgShifterPatternHi uint16

	// Background attribute shifters (16-bit)
	// Holds palette selection for 16 pixels
	bgShifterAttribLo uint16
	bgShifterAttribHi uint16

	// ========================================================================
	// Sprite Rendering State
	// ========================================================================

	// Secondary OAM - holds sprites for current scanline (8 sprites max)
	// During sprite evaluation, the PPU scans primary OAM and copies
	// sprites that are visible on the next scanline to secondary OAM
	secondaryOAM [32]uint8 // 8 sprites * 4 bytes each

	// Sprite count for current scanline (0-8)
	spriteCount uint8

	// Sprite 0 present on current scanline (for sprite 0 hit detection)
	sprite0Present bool

	// Sprite shifters - hold pattern data for up to 8 sprites
	spriteShifterPatternLo [8]uint8
	spriteShifterPatternHi [8]uint8

	// Sprite attributes for current scanline
	spriteAttributes [8]uint8

	// Sprite X positions for current scanline
	spritePositions [8]uint8

	// ========================================================================
	// Cartridge Interface
	// ========================================================================

	// Cartridge mapper for CHR-ROM/CHR-RAM access
	mapper cartridge.Mapper

	// Nametable mirroring mode
	mirroringMode uint8

	// ========================================================================
	// Output
	// ========================================================================

	// Frame buffer (256x240 pixels, each pixel is a palette index 0-63)
	frameBuffer [ScreenWidth * ScreenHeight]uint8

	// NMI output signal (triggers CPU interrupt)
	nmiOutput bool
}

// NewPPU creates and initializes a new PPU
func NewPPU() *PPU {
	ppu := &PPU{
		scanline: 0,
		cycle:    0,
		frame:    0,
	}

	// Initialize palette RAM to default values
	for i := range ppu.paletteRAM {
		ppu.paletteRAM[i] = 0x00
	}

	return ppu
}

// SetMapper connects a cartridge mapper to the PPU for CHR-ROM/RAM access
func (p *PPU) SetMapper(mapper cartridge.Mapper) {
	p.mapper = mapper
}

// SetMirroring sets the nametable mirroring mode
func (p *PPU) SetMirroring(mode uint8) {
	p.mirroringMode = mode
}

// Clock advances the PPU by one cycle
// The PPU runs at 3x the CPU speed, so this should be called 3 times per CPU cycle
func (p *PPU) Clock() {
	// ====================================================================
	// Pixel Rendering - happens BEFORE shifter updates and fetching
	// ====================================================================
	if p.scanline >= 0 && p.scanline < 240 && p.cycle >= 1 && p.cycle <= 256 {
		p.renderPixel()
	}

	// ====================================================================
	// Pre-render and Visible Scanlines (-1, 0-239)
	// ====================================================================
	if p.scanline >= -1 && p.scanline < 240 {

		// Clear flags at start of pre-render scanline
		if p.scanline == -1 && p.cycle == 1 {
			p.status.SetVBlank(false)
			p.status.SetSprite0Hit(false)
			p.status.SetSpriteOverflow(false)
			p.frameComplete = false
		}

		// Background rendering cycles
		if (p.cycle >= 2 && p.cycle < 258) || (p.cycle >= 321 && p.cycle < 338) {

			// Update shifters every cycle
			p.updateShifters()

			// 8-cycle fetching pattern
			switch (p.cycle - 1) % 8 {
			case 0:
				// Load shifters with data from previous fetch
				p.loadBackgroundShifters()

				// Fetch next tile ID from nametable
				p.bgNextTileID = p.ppuRead(0x2000 | (p.vramAddress.Get() & 0x0FFF))

			case 2:
				// Fetch attribute byte
				address := uint16(0x23C0) |
					(p.vramAddress.NametableY() << 11) |
					(p.vramAddress.NametableX() << 10) |
					((p.vramAddress.CoarseY() >> 2) << 3) |
					(p.vramAddress.CoarseX() >> 2)

				p.bgNextTileAttrib = p.ppuRead(address)

				// Extract the 2 bits for this 2x2 tile quadrant
				if p.vramAddress.CoarseY()&0x02 != 0 {
					p.bgNextTileAttrib >>= 4
				}
				if p.vramAddress.CoarseX()&0x02 != 0 {
					p.bgNextTileAttrib >>= 2
				}
				p.bgNextTileAttrib &= 0x03

			case 4:
				// Fetch tile pattern low byte
				table := p.control.BackgroundPatternTable()
				tileID := uint16(p.bgNextTileID)
				fineY := p.vramAddress.FineY()
				address := table | (tileID << 4) | fineY
				p.bgNextTileLSB = p.ppuRead(address)

			case 6:
				// Fetch tile pattern high byte (same as low + 8)
				table := p.control.BackgroundPatternTable()
				tileID := uint16(p.bgNextTileID)
				fineY := p.vramAddress.FineY()
				address := table | (tileID << 4) | fineY
				p.bgNextTileMSB = p.ppuRead(address + 8)

			case 7:
				// Increment horizontal scroll
				if p.mask.IsRenderingEnabled() {
					p.vramAddress.IncrementX()
				}
			}
		}

		// End of visible scanline: increment vertical scroll
		if p.cycle == 256 {
			if p.mask.IsRenderingEnabled() {
				p.vramAddress.IncrementY()
			}
		}

		// Reset horizontal position and start sprite fetching
		if p.cycle == 257 {
			p.loadBackgroundShifters()
			if p.mask.IsRenderingEnabled() {
				p.vramAddress.TransferX(&p.tempVRAMAddress)
			}
			// Sprite evaluation for next scanline
			if p.scanline >= -1 && p.scanline < 240 {
				p.spriteEvaluation()
			}
		}

		// Sprite pattern fetching (cycles 257-320)
		if p.cycle == 320 {
			if p.scanline >= -1 && p.scanline < 240 {
				p.spriteFetching()
			}
		}

		// Superfluous nametable fetches at end of scanline
		if p.cycle == 338 || p.cycle == 340 {
			p.bgNextTileID = p.ppuRead(0x2000 | (p.vramAddress.Get() & 0x0FFF))
		}

		// Pre-render scanline: restore vertical position
		if p.scanline == -1 && p.cycle >= 280 && p.cycle < 305 {
			if p.mask.IsRenderingEnabled() {
				p.vramAddress.TransferY(&p.tempVRAMAddress)
			}
		}
	}

	// ====================================================================
	// Post-render Scanline (240)
	// ====================================================================
	// Idle - PPU does nothing

	// ====================================================================
	// VBlank Scanlines (241-260)
	// ====================================================================
	if p.scanline == 241 && p.cycle == 1 {
		// Set VBlank flag
		p.status.SetVBlank(true)

		// Trigger NMI if enabled
		if p.control.EnableNMI() {
			p.nmiOutput = true
		}
	}

	// ====================================================================
	// Advance Timing
	// ====================================================================
	p.cycle++

	// End of scanline
	if p.cycle >= CyclesPerScanline {
		p.cycle = 0
		p.scanline++

		// Odd frame skip: On odd frames, when rendering is enabled,
		// cycle 0 of scanline 0 is skipped
		if p.scanline == 0 && (p.frame&1) == 1 && p.mask.IsRenderingEnabled() {
			p.cycle = 1
		}

		// End of frame
		if p.scanline >= ScanlinesPerFrame {
			p.scanline = -1
			p.frameComplete = true
			p.frame++
			p.oddFrame = !p.oddFrame
		}
	}
}

// GetNMI returns and clears the NMI output signal
func (p *PPU) GetNMI() bool {
	nmi := p.nmiOutput
	p.nmiOutput = false
	return nmi
}

// GetFrameBuffer returns a pointer to the current frame buffer
func (p *PPU) GetFrameBuffer() *[ScreenWidth * ScreenHeight]uint8 {
	return &p.frameBuffer
}

// IsFrameComplete returns true if a frame has been fully rendered
func (p *PPU) IsFrameComplete() bool {
	return p.frameComplete
}

// ClearFrameComplete resets the frame complete flag
func (p *PPU) ClearFrameComplete() {
	p.frameComplete = false
}

// Reset initializes the PPU to power-on state
func (p *PPU) Reset() {
	p.control.Set(0)
	p.mask.Set(0)
	p.status.Set(0)
	p.oamAddress = 0
	p.writeLatch = false
	p.vramAddress.Set(0)
	p.tempVRAMAddress.Set(0)
	p.fineX = 0
	p.readBuffer = 0
	p.scanline = 0
	p.cycle = 0
	p.nmiOutput = false
}

// ========================================================================
// CPU Register Interface ($2000-$2007)
// ========================================================================

// WriteCPURegister handles writes from the CPU to PPU registers ($2000-$2007)
func (p *PPU) WriteCPURegister(addr uint16, value uint8) {
	switch addr {
	case 0x2000: // PPUCTRL
		p.control.Set(value)
		// t: ...GH.. ........ <- d: ......GH
		p.tempVRAMAddress.SetNametableX(uint16(p.control.NametableX()))
		p.tempVRAMAddress.SetNametableY(uint16(p.control.NametableY()))

	case 0x2001: // PPUMASK
		p.mask.Set(value)

	case 0x2003: // OAMADDR
		p.oamAddress = value

	case 0x2004: // OAMDATA
		p.oam[p.oamAddress] = value
		p.oamAddress++ // Wraps around

	case 0x2005: // PPUSCROLL
		if !p.writeLatch {
			// First write (X scroll)
			// t: ....... ...ABCDE <- d: ABCDE...
			// x:              FGH <- d: .....FGH
			p.tempVRAMAddress.SetCoarseX(uint16(value >> 3))
			p.fineX = value & 0x07
			p.writeLatch = true
		} else {
			// Second write (Y scroll)
			// t: FGH..AB CDE..... <- d: ABCDEFGH
			p.tempVRAMAddress.SetFineY(uint16(value & 0x07))
			p.tempVRAMAddress.SetCoarseY(uint16(value >> 3))
			p.writeLatch = false
		}

	case 0x2006: // PPUADDR
		if !p.writeLatch {
			// First write (high byte)
			// t: .CDEFGH ........ <- d: ..CDEFGH
			// t: X...... ........ <- 0
			p.tempVRAMAddress.Set((p.tempVRAMAddress.Get() & 0x00FF) | ((uint16(value) & 0x3F) << 8))
			p.writeLatch = true
		} else {
			// Second write (low byte)
			// t: ....... ABCDEFGH <- d: ABCDEFGH
			// v: <...all bits...> <- t: <...all bits...>
			p.tempVRAMAddress.Set((p.tempVRAMAddress.Get() & 0xFF00) | uint16(value))
			p.vramAddress.Set(p.tempVRAMAddress.Get())
			p.writeLatch = false
		}

	case 0x2007: // PPUDATA
		p.ppuWrite(p.vramAddress.Get(), value)
		p.vramAddress.Set(p.vramAddress.Get() + p.control.IncrementMode())
	}
}

// ReadCPURegister handles reads from the CPU to PPU registers ($2000-$2007)
func (p *PPU) ReadCPURegister(addr uint16) uint8 {
	var value uint8

	switch addr {
	case 0x2002: // PPUSTATUS
		value = p.status.Get()
		// Reading PPUSTATUS clears VBlank flag and write latch
		p.status.SetVBlank(false)
		p.writeLatch = false

	case 0x2004: // OAMDATA
		value = p.oam[p.oamAddress]

	case 0x2007: // PPUDATA
		value = p.readBuffer
		p.readBuffer = p.ppuRead(p.vramAddress.Get())

		// Palette reads are not buffered
		if p.vramAddress.Get() >= 0x3F00 {
			value = p.readBuffer
		}

		p.vramAddress.Set(p.vramAddress.Get() + p.control.IncrementMode())
	}

	return value
}

// ========================================================================
// Internal PPU Memory Access
// ========================================================================

// ppuRead reads from PPU memory space ($0000-$3FFF)
func (p *PPU) ppuRead(addr uint16) uint8 {
	addr &= 0x3FFF // 14-bit address space

	switch {
	case addr < 0x2000:
		// Pattern tables (CHR-ROM/RAM)
		if p.mapper != nil {
			return p.mapper.ReadCHR(addr)
		}
		return 0

	case addr < 0x3F00:
		// Nametables
		return p.nametable[p.mirrorNametableAddress(addr)]

	case addr < 0x4000:
		// Palette RAM
		addr = p.mirrorPaletteAddress(addr)
		return p.paletteRAM[addr]
	}

	return 0
}

// ppuWrite writes to PPU memory space ($0000-$3FFF)
func (p *PPU) ppuWrite(addr uint16, value uint8) {
	addr &= 0x3FFF // 14-bit address space

	switch {
	case addr < 0x2000:
		// Pattern tables (CHR-ROM/RAM)
		if p.mapper != nil {
			p.mapper.WriteCHR(addr, value)
		}

	case addr < 0x3F00:
		// Nametables
		p.nametable[p.mirrorNametableAddress(addr)] = value

	case addr < 0x4000:
		// Palette RAM
		addr = p.mirrorPaletteAddress(addr)
		p.paletteRAM[addr] = value
	}
}

// mirrorNametableAddress applies nametable mirroring to get actual RAM address
// Adapted from fogleman/nes for correctness
func (p *PPU) mirrorNametableAddress(addr uint16) uint16 {
	addr = (addr - 0x2000) % 0x1000
	table := addr / 0x0400
	offset := addr % 0x0400
	switch p.mirroringMode {
	case MirrorVertical:
		return addr % 0x0800
	case MirrorHorizontal:
		return (table/2)*0x0400 + offset
	case MirrorSingleLow:
		return offset
	case MirrorSingleHigh:
		return 0x0400 + offset
	case MirrorFourScreen:
		return addr
	}
	return 0
}

// mirrorPaletteAddress applies palette mirroring ($3F00-$3F1F)
func (p *PPU) mirrorPaletteAddress(addr uint16) uint16 {
	addr = (addr - 0x3F00) % 32

	// Mirror $3F10, $3F14, $3F18, $3F1C to $3F00, $3F04, $3F08, $3F0C
	if addr >= 16 && addr%4 == 0 {
		addr -= 16
	}

	return addr
}
