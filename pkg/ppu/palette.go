package ppu

// Color represents an RGB color
type Color struct {
	R, G, B uint8
}

// HardwarePalette is the NES hardware color palette (64 colors)
//
// These are the actual RGB colors the NES can display. The palette RAM
// contains indices (0x00-0x3F) that map to these colors.
//
// This is the standard NTSC palette.
var HardwarePalette = [64]Color{
	{84, 84, 84}, {0, 30, 116}, {8, 16, 144}, {48, 0, 136},
	{68, 0, 100}, {92, 0, 48}, {84, 4, 0}, {60, 24, 0},
	{32, 42, 0}, {8, 58, 0}, {0, 64, 0}, {0, 60, 0},
	{0, 50, 60}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0},

	{152, 150, 152}, {8, 76, 196}, {48, 50, 236}, {92, 30, 228},
	{136, 20, 176}, {160, 20, 100}, {152, 34, 32}, {120, 60, 0},
	{84, 90, 0}, {40, 114, 0}, {8, 124, 0}, {0, 118, 40},
	{0, 102, 120}, {0, 0, 0}, {0, 0, 0}, {0, 0, 0},

	{236, 238, 236}, {76, 154, 236}, {120, 124, 236}, {176, 98, 236},
	{228, 84, 236}, {236, 88, 180}, {236, 106, 100}, {212, 136, 32},
	{160, 170, 0}, {116, 196, 0}, {76, 208, 32}, {56, 204, 108},
	{56, 180, 204}, {60, 60, 60}, {0, 0, 0}, {0, 0, 0},

	{236, 238, 236}, {168, 204, 236}, {188, 188, 236}, {212, 178, 236},
	{236, 174, 236}, {236, 174, 212}, {236, 180, 176}, {228, 196, 144},
	{204, 210, 120}, {180, 222, 120}, {168, 226, 144}, {152, 226, 180},
	{160, 214, 228}, {160, 162, 160}, {0, 0, 0}, {0, 0, 0},
}

// GetColorFromPalette retrieves an RGB color from the palette system
//
// paletteIndex: Which palette (0-7: 0-3 background, 4-7 sprite)
// pixelValue: Which color within palette (0-3)
func (p *PPU) GetColorFromPalette(paletteIndex uint8, pixelValue uint8) Color {
	// Calculate palette RAM address
	address := uint16((paletteIndex << 2) | (pixelValue & 0x03))

	// Read palette index from palette RAM
	colorIndex := p.ppuRead(0x3F00+address) & 0x3F

	// Return RGB color from hardware palette
	return HardwarePalette[colorIndex]
}
