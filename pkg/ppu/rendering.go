package ppu

// Background rendering helper functions

// loadBackgroundShifters loads shifters with next tile data
// Called every 8 cycles to prime shifters with next 8 pixels
func (p *PPU) loadBackgroundShifters() {
	// Load pattern shifters
	// Load new tile data into LOW bits (bits 7-0)
	// The HIGH bits (bits 15-8) contain data currently being rendered
	p.bgShifterPatternLo = (p.bgShifterPatternLo & 0xFF00) | uint16(p.bgNextTileLSB)
	p.bgShifterPatternHi = (p.bgShifterPatternHi & 0xFF00) | uint16(p.bgNextTileMSB)

	// Load attribute shifters
	// Attribute doesn't change per pixel, so "inflate" 2-bit palette to fill low 8 bits
	if p.bgNextTileAttrib&0x01 != 0 {
		p.bgShifterAttribLo = (p.bgShifterAttribLo & 0xFF00) | 0x00FF
	} else {
		p.bgShifterAttribLo = (p.bgShifterAttribLo & 0xFF00)
	}

	if p.bgNextTileAttrib&0x02 != 0 {
		p.bgShifterAttribHi = (p.bgShifterAttribHi & 0xFF00) | 0x00FF
	} else {
		p.bgShifterAttribHi = (p.bgShifterAttribHi & 0xFF00)
	}
}

// updateShifters shifts background shifters left by 1 bit
// Called every cycle during rendering to advance pixel output
func (p *PPU) updateShifters() {
	if p.mask.RenderBackground() {
		// Shift pattern shifters left by 1
		p.bgShifterPatternLo <<= 1
		p.bgShifterPatternHi <<= 1

		// Shift attribute shifters left by 1
		p.bgShifterAttribLo <<= 1
		p.bgShifterAttribHi <<= 1
	}
}

// renderPixel composes and outputs a single pixel to the frame buffer
// Called during visible scanlines (0-239) at cycles 1-256
func (p *PPU) renderPixel() {
	x := p.cycle - 1
	y := uint16(p.scanline)

	// Validate coordinates
	if x >= ScreenWidth || y >= ScreenHeight {
		return
	}

	// If rendering is completely disabled, output backdrop color only
	if !p.mask.IsRenderingEnabled() {
		// Rendering disabled - show backdrop color ($3F00)
		backdropColor := p.ppuRead(0x3F00) & 0x3F
		p.frameBuffer[y*ScreenWidth+x] = backdropColor
		return
	}

	// Rendering enabled - compose background and sprite pixels
	// Background pixel
	bgPixel := uint8(0)
	bgPalette := uint8(0)

	if p.mask.RenderBackground() {
		// Select bit based on fine X scroll
		bitMux := uint16(0x8000 >> p.fineX)

		// Extract pixel value (2 bits)
		p0 := uint8(0)
		if p.bgShifterPatternLo&bitMux != 0 {
			p0 = 1
		}
		p1 := uint8(0)
		if p.bgShifterPatternHi&bitMux != 0 {
			p1 = 1
		}
		bgPixel = (p1 << 1) | p0

		// Extract palette (2 bits)
		pal0 := uint8(0)
		if p.bgShifterAttribLo&bitMux != 0 {
			pal0 = 1
		}
		pal1 := uint8(0)
		if p.bgShifterAttribHi&bitMux != 0 {
			pal1 = 1
		}
		bgPalette = (pal1 << 1) | pal0
	}

	// Render sprites and get sprite pixel
	spritePixel, spritePalette, spritePriority, isSprite0 := p.renderSprites(x)

	// Composite background and sprite pixels
	finalPixel := uint8(0)
	finalPalette := uint8(0)

	// Determine which pixel to show based on priority
	if bgPixel == 0 && spritePixel == 0 {
		// Both transparent - use backdrop color
		finalPixel = 0
		finalPalette = 0
	} else if bgPixel == 0 && spritePixel > 0 {
		// Background transparent, sprite visible - use sprite
		finalPixel = spritePixel
		finalPalette = spritePalette + 4 // Sprite palettes are 4-7
	} else if bgPixel > 0 && spritePixel == 0 {
		// Sprite transparent, background visible - use background
		finalPixel = bgPixel
		finalPalette = bgPalette
	} else {
		// Both visible - check priority
		if spritePriority {
			// Sprite in front - use sprite
			finalPixel = spritePixel
			finalPalette = spritePalette + 4
		} else {
			// Background in front - use background
			finalPixel = bgPixel
			finalPalette = bgPalette
		}

		// Sprite 0 hit detection
		if isSprite0 && x < 255 && x >= 1 {
			// Sprite 0 hit occurs when both background and sprite 0 have
			// opaque pixels overlapping (not at x=255)
			if p.mask.RenderBackground() && p.mask.RenderSprites() {
				// Don't set hit if rendering is disabled in leftmost 8 pixels
				if p.mask.RenderBackgroundLeft() || x >= 8 {
					p.status.SetSprite0Hit(true)
				}
			}
		}
	}

	// Write to frame buffer
	address := uint16((finalPalette << 2) | (finalPixel & 0x03))
	colorIndex := p.ppuRead(0x3F00+address) & 0x3F
	p.frameBuffer[y*ScreenWidth+x] = colorIndex
}
