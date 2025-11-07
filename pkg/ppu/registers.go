package ppu

// PPUControl represents the PPUCTRL register ($2000) - Write Only
//
// Controls PPU behavior including nametable selection, address increment,
// sprite/background pattern table selection, and NMI enable.
//
// Bit layout (VPHB SINN):
//   7: V = NMI enable (0: off; 1: on)
//   6: P = PPU master/slave select (unused in NES)
//   5: H = Sprite height (0: 8x8 pixels; 1: 8x16 pixels)
//   4: B = Background pattern table address (0: $0000; 1: $1000)
//   3: S = Sprite pattern table address (0: $0000; 1: $1000)
//   2: I = VRAM address increment per CPU read/write (0: add 1, across; 1: add 32, down)
//   1-0: NN = Base nametable address (0 = $2000; 1 = $2400; 2 = $2800; 3 = $2C00)
type PPUControl struct {
	register uint8
}

// Set writes a value to the PPUCTRL register
func (c *PPUControl) Set(value uint8) {
	c.register = value
}

// Get returns the current PPUCTRL register value
func (c *PPUControl) Get() uint8 {
	return c.register
}

// NametableX returns the X component of base nametable address (bit 0)
func (c *PPUControl) NametableX() uint8 {
	return (c.register >> 0) & 0x01
}

// NametableY returns the Y component of base nametable address (bit 1)
func (c *PPUControl) NametableY() uint8 {
	return (c.register >> 1) & 0x01
}

// IncrementMode returns VRAM address increment per CPU read/write (1 or 32)
func (c *PPUControl) IncrementMode() uint16 {
	if (c.register>>2)&0x01 != 0 {
		return 32
	}
	return 1
}

// SpritePatternTable returns sprite pattern table base address (0x0000 or 0x1000)
func (c *PPUControl) SpritePatternTable() uint16 {
	if (c.register>>3)&0x01 != 0 {
		return 0x1000
	}
	return 0x0000
}

// BackgroundPatternTable returns background pattern table base address (0x0000 or 0x1000)
func (c *PPUControl) BackgroundPatternTable() uint16 {
	if (c.register>>4)&0x01 != 0 {
		return 0x1000
	}
	return 0x0000
}

// SpriteSize returns sprite size (0 = 8x8, 1 = 8x16)
func (c *PPUControl) SpriteSize() uint8 {
	return (c.register >> 5) & 0x01
}

// SlaveMode returns PPU master/slave select (unused in NES)
func (c *PPUControl) SlaveMode() bool {
	return (c.register>>6)&0x01 != 0
}

// EnableNMI returns whether to generate NMI at start of vblank
func (c *PPUControl) EnableNMI() bool {
	return (c.register>>7)&0x01 != 0
}

// PPUMask represents the PPUMASK register ($2001) - Write Only
//
// Controls rendering options including grayscale, color emphasis,
// and sprite/background enable.
//
// Bit layout (BGRs bMmG):
//   7: B = Emphasize blue
//   6: G = Emphasize green (NTSC) / red (PAL)
//   5: R = Emphasize red (NTSC) / green (PAL)
//   4: s = Show sprites
//   3: b = Show background
//   2: M = Show sprites in leftmost 8 pixels of screen
//   1: m = Show background in leftmost 8 pixels of screen
//   0: G = Grayscale (0: normal color, 1: produce a grayscale display)
type PPUMask struct {
	register uint8
}

// Set writes a value to the PPUMASK register
func (m *PPUMask) Set(value uint8) {
	m.register = value
}

// Get returns the current PPUMASK register value
func (m *PPUMask) Get() uint8 {
	return m.register
}

// Grayscale returns whether grayscale mode is enabled
func (m *PPUMask) Grayscale() bool {
	return (m.register>>0)&0x01 != 0
}

// RenderBackgroundLeft returns whether to show background in leftmost 8 pixels
func (m *PPUMask) RenderBackgroundLeft() bool {
	return (m.register>>1)&0x01 != 0
}

// RenderSpritesLeft returns whether to show sprites in leftmost 8 pixels
func (m *PPUMask) RenderSpritesLeft() bool {
	return (m.register>>2)&0x01 != 0
}

// RenderBackground returns whether background rendering is enabled
func (m *PPUMask) RenderBackground() bool {
	return (m.register>>3)&0x01 != 0
}

// RenderSprites returns whether sprite rendering is enabled
func (m *PPUMask) RenderSprites() bool {
	return (m.register>>4)&0x01 != 0
}

// EmphasizeRed returns whether to emphasize red (NTSC) / green (PAL)
func (m *PPUMask) EmphasizeRed() bool {
	return (m.register>>5)&0x01 != 0
}

// EmphasizeGreen returns whether to emphasize green (NTSC) / red (PAL)
func (m *PPUMask) EmphasizeGreen() bool {
	return (m.register>>6)&0x01 != 0
}

// EmphasizeBlue returns whether to emphasize blue
func (m *PPUMask) EmphasizeBlue() bool {
	return (m.register>>7)&0x01 != 0
}

// IsRenderingEnabled returns true if rendering is enabled (background OR sprites)
func (m *PPUMask) IsRenderingEnabled() bool {
	return m.RenderBackground() || m.RenderSprites()
}

// PPUStatus represents the PPUSTATUS register ($2002) - Read Only
//
// Reports PPU status including vblank, sprite 0 hit, and sprite overflow.
//
// Bit layout (VSO- ----):
//   7: V = Vertical blank has started (0: not in vblank; 1: in vblank)
//   6: S = Sprite 0 Hit (set when a nonzero pixel of sprite 0 overlaps a nonzero background pixel)
//   5: O = Sprite overflow (more than 8 sprites appear on a scanline)
//   4-0: Unused (returns PPU open bus values)
type PPUStatus struct {
	register uint8
}

// Set writes a value to the PPUSTATUS register (internal use only)
func (s *PPUStatus) Set(value uint8) {
	s.register = value
}

// Get returns the current PPUSTATUS register value
func (s *PPUStatus) Get() uint8 {
	return s.register
}

// SetVBlank sets or clears the vertical blank flag
func (s *PPUStatus) SetVBlank(value bool) {
	if value {
		s.register |= 0x80
	} else {
		s.register &= ^uint8(0x80)
	}
}

// VBlank returns whether vertical blank has started
func (s *PPUStatus) VBlank() bool {
	return (s.register>>7)&0x01 != 0
}

// SetSprite0Hit sets or clears the sprite 0 hit flag
func (s *PPUStatus) SetSprite0Hit(value bool) {
	if value {
		s.register |= 0x40
	} else {
		s.register &= ^uint8(0x40)
	}
}

// Sprite0Hit returns whether sprite 0 hit has occurred
func (s *PPUStatus) Sprite0Hit() bool {
	return (s.register>>6)&0x01 != 0
}

// SetSpriteOverflow sets or clears the sprite overflow flag
func (s *PPUStatus) SetSpriteOverflow(value bool) {
	if value {
		s.register |= 0x20
	} else {
		s.register &= ^uint8(0x20)
	}
}

// SpriteOverflow returns whether sprite overflow has occurred
func (s *PPUStatus) SpriteOverflow() bool {
	return (s.register>>5)&0x01 != 0
}

// LoopyRegister represents a PPU scrolling register (15-bit)
//
// Internal PPU registers used for scrolling and addressing.
// Named after Loopy's documentation of the PPU.
//
// Bit layout (yyy NN YYYYY XXXXX):
//   14-12: yyy = Fine Y scroll (0-7)
//   11-10: NN = Nametable select
//   9-5:   YYYYY = Coarse Y scroll (0-29)
//   4-0:   XXXXX = Coarse X scroll (0-31)
type LoopyRegister struct {
	register uint16
}

// Set writes a value to the Loopy register
func (l *LoopyRegister) Set(value uint16) {
	l.register = value & 0x7FFF // 15-bit register
}

// Get returns the current Loopy register value
func (l *LoopyRegister) Get() uint16 {
	return l.register
}

// CoarseX returns the coarse X scroll value (bits 0-4)
func (l *LoopyRegister) CoarseX() uint16 {
	return l.register & 0x001F
}

// SetCoarseX sets the coarse X scroll value (bits 0-4)
func (l *LoopyRegister) SetCoarseX(value uint16) {
	l.register = (l.register & 0x7FE0) | (value & 0x001F)
}

// CoarseY returns the coarse Y scroll value (bits 5-9)
func (l *LoopyRegister) CoarseY() uint16 {
	return (l.register & 0x03E0) >> 5
}

// SetCoarseY sets the coarse Y scroll value (bits 5-9)
func (l *LoopyRegister) SetCoarseY(value uint16) {
	l.register = (l.register & 0x7C1F) | ((value & 0x001F) << 5)
}

// NametableX returns the nametable X select (bit 10)
func (l *LoopyRegister) NametableX() uint16 {
	return (l.register & 0x0400) >> 10
}

// SetNametableX sets the nametable X select (bit 10)
func (l *LoopyRegister) SetNametableX(value uint16) {
	if value != 0 {
		l.register |= 0x0400
	} else {
		l.register &= ^uint16(0x0400)
	}
}

// NametableY returns the nametable Y select (bit 11)
func (l *LoopyRegister) NametableY() uint16 {
	return (l.register & 0x0800) >> 11
}

// SetNametableY sets the nametable Y select (bit 11)
func (l *LoopyRegister) SetNametableY(value uint16) {
	if value != 0 {
		l.register |= 0x0800
	} else {
		l.register &= ^uint16(0x0800)
	}
}

// FineY returns the fine Y scroll value (bits 12-14)
func (l *LoopyRegister) FineY() uint16 {
	return (l.register & 0x7000) >> 12
}

// SetFineY sets the fine Y scroll value (bits 12-14)
func (l *LoopyRegister) SetFineY(value uint16) {
	l.register = (l.register & 0x0FFF) | ((value & 0x0007) << 12)
}

// IncrementX increments horizontal scroll (move right by 1 tile)
//
// This is called during rendering to move to the next tile.
// When reaching the end of a nametable (32 tiles), it wraps
// to 0 and flips the horizontal nametable bit.
func (l *LoopyRegister) IncrementX() {
	if l.CoarseX() == 31 {
		// Wrap coarse X to 0
		l.SetCoarseX(0)
		// Flip horizontal nametable
		l.SetNametableX(l.NametableX() ^ 1)
	} else {
		// Increment coarse X
		l.SetCoarseX(l.CoarseX() + 1)
	}
}

// IncrementY increments vertical scroll (move down by 1 scanline)
//
// This is called at the end of each scanline during rendering.
// First increments fine Y (pixel offset within tile), then
// when that overflows, increments coarse Y (tile row).
//
// Hardware bug: Coarse Y can go up to 31, but nametables are only
// 30 tiles tall. Rows 30-31 actually access attribute table data.
func (l *LoopyRegister) IncrementY() {
	if l.FineY() < 7 {
		// Increment fine Y (still within same tile)
		l.SetFineY(l.FineY() + 1)
	} else {
		// Fine Y overflows, reset to 0
		l.SetFineY(0)

		// Now increment coarse Y
		y := l.CoarseY()

		if y == 29 {
			// Reached bottom of nametable (30 rows)
			y = 0
			// Flip vertical nametable
			l.SetNametableY(l.NametableY() ^ 1)
		} else if y == 31 {
			// Hardware bug: Row 31 wraps to 0 without flipping nametable
			y = 0
		} else {
			// Normal increment
			y++
		}

		l.SetCoarseY(y)
	}
}

// TransferX transfers horizontal bits from another register
//
// Copies coarse X and nametable X from source.
// This is called at the end of each scanline (cycle 257) to reset
// horizontal position for the next scanline.
func (l *LoopyRegister) TransferX(source *LoopyRegister) {
	// v: ....A.. ...BCDEF <- t: ....A.. ...BCDEF
	//   (Coarse X, Nametable X)
	l.register = (l.register & 0x7BE0) | (source.register & 0x041F)
}

// TransferY transfers vertical bits from another register
//
// Copies coarse Y, nametable Y, and fine Y from source.
// This is called during pre-render scanline (cycles 280-304) to
// reset vertical position for the next frame.
func (l *LoopyRegister) TransferY(source *LoopyRegister) {
	// v: GHI.J.KLM NOPQRST <- t: GHI.J.KLM NOPQRST
	//   (Fine Y, Nametable Y, Coarse Y)
	l.register = (l.register & 0x041F) | (source.register & 0x7BE0)
}
