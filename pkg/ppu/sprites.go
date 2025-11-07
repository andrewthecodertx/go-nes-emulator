package ppu

// spriteEvaluation performs sprite evaluation for the next scanline.
// This happens during cycles 65-256 of the current scanline.
// The PPU examines all 64 sprites in OAM and determines which ones
// are visible on the next scanline (up to 8 sprites max).
func (p *PPU) spriteEvaluation() {
	// Clear secondary OAM
	for i := range p.secondaryOAM {
		p.secondaryOAM[i] = 0xFF
	}

	p.spriteCount = 0
	p.sprite0Present = false

	// Debug: Only evaluate if rendering is enabled
	if !p.mask.IsRenderingEnabled() {
		return
	}

	// Get sprite height (8x8 or 8x16)
	spriteHeight := uint16(8)
	if p.control.SpriteSize() != 0 {
		spriteHeight = 16
	}

	// Scan through all 64 sprites
	for i := uint8(0); i < 64; i++ {
		// Read sprite Y position (byte 0 of sprite data)
		oamIndex := uint16(i) * 4
		spriteY := uint16(p.oam[oamIndex])

		// Calculate the difference between current scanline and sprite Y
		// The sprite is visible if the scanline is within sprite height
		diff := uint16(p.scanline) - spriteY

		// Check if sprite is on the next scanline
		if diff < spriteHeight {
			// Check if we've already found 8 sprites
			if p.spriteCount >= 8 {
				// Set sprite overflow flag
				p.status.SetSpriteOverflow(true)
				break
			}

			// Copy sprite to secondary OAM
			secondaryIndex := uint16(p.spriteCount) * 4
			p.secondaryOAM[secondaryIndex+0] = p.oam[oamIndex+0] // Y position
			p.secondaryOAM[secondaryIndex+1] = p.oam[oamIndex+1] // Tile index
			p.secondaryOAM[secondaryIndex+2] = p.oam[oamIndex+2] // Attributes
			p.secondaryOAM[secondaryIndex+3] = p.oam[oamIndex+3] // X position

			// Check if this is sprite 0
			if i == 0 {
				p.sprite0Present = true
			}

			p.spriteCount++
		}
	}
}

// spriteFetching fetches pattern data for all sprites in secondary OAM.
// This happens during cycles 257-320 of the current scanline.
func (p *PPU) spriteFetching() {
	// Get sprite height and pattern table address
	spriteHeight := uint16(8)
	spritePatternTable := p.control.SpritePatternTable()

	// For 8x16 sprites, height is 16 and pattern table is determined per sprite
	if p.control.SpriteSize() != 0 {
		spriteHeight = 16
	}

	// Fetch pattern data for each sprite in secondary OAM
	for i := uint8(0); i < p.spriteCount; i++ {
		secondaryIndex := uint16(i) * 4

		// Read sprite data from secondary OAM
		spriteY := p.secondaryOAM[secondaryIndex+0]
		tileIndex := p.secondaryOAM[secondaryIndex+1]
		attributes := p.secondaryOAM[secondaryIndex+2]
		spriteX := p.secondaryOAM[secondaryIndex+3]

		// Store attributes and X position for rendering
		p.spriteAttributes[i] = attributes
		p.spritePositions[i] = spriteX

		// Calculate which row of the sprite we're on
		spriteRow := uint16(p.scanline) - uint16(spriteY)

		// Check vertical flip
		if attributes&0x80 != 0 {
			// Flip vertically
			spriteRow = spriteHeight - 1 - spriteRow
		}

		// Calculate pattern address
		var patternAddress uint16

		if spriteHeight == 16 {
			// 8x16 sprites
			// Bit 0 of tile index selects pattern table
			// Bits 1-7 select tile pair
			if spriteRow < 8 {
				// Top half
				patternAddress = (uint16(tileIndex&0x01) << 12) |
					(uint16(tileIndex&0xFE) << 4) |
					(spriteRow & 0x07)
			} else {
				// Bottom half
				patternAddress = (uint16(tileIndex&0x01) << 12) |
					((uint16(tileIndex&0xFE) + 1) << 4) |
					((spriteRow - 8) & 0x07)
			}
		} else {
			// 8x8 sprites
			patternAddress = (spritePatternTable << 12) |
				(uint16(tileIndex) << 4) |
				(spriteRow & 0x07)
		}

		// Fetch pattern data (low and high bytes)
		patternLow := p.ppuRead(patternAddress)
		patternHigh := p.ppuRead(patternAddress + 8)

		// Check horizontal flip
		if attributes&0x40 != 0 {
			// Flip horizontally by reversing bits
			patternLow = reverseByte(patternLow)
			patternHigh = reverseByte(patternHigh)
		}

		// Store in sprite shifters
		p.spriteShifterPatternLo[i] = patternLow
		p.spriteShifterPatternHi[i] = patternHigh
	}
}

// reverseByte reverses the bits in a byte (used for horizontal sprite flipping)
func reverseByte(b uint8) uint8 {
	b = (b&0xF0)>>4 | (b&0x0F)<<4
	b = (b&0xCC)>>2 | (b&0x33)<<2
	b = (b&0xAA)>>1 | (b&0x55)<<1
	return b
}

// renderSprites renders sprites for the current pixel.
// Returns the sprite pixel value (0-3), palette index (0-3), and priority flag.
// If no sprite pixel is active, returns (0, 0, false).
func (p *PPU) renderSprites(x uint16) (pixel uint8, palette uint8, priority bool, isSprite0 bool) {
	// Sprite rendering must be enabled
	if !p.mask.RenderSprites() {
		return 0, 0, false, false
	}

	// Check left 8 pixels masking
	if x < 8 && !p.mask.RenderSpritesLeft() {
		return 0, 0, false, false
	}

	// Check each sprite to see if it's being rendered at this X position
	for i := uint8(0); i < p.spriteCount; i++ {
		// Calculate offset from sprite's X position
		offset := int16(x) - int16(p.spritePositions[i])

		// Check if we're within the sprite's 8-pixel width
		if offset >= 0 && offset < 8 {
			// Get the pixel from the sprite pattern shifters
			// Shift amount: we want bit 7 when offset=0, bit 6 when offset=1, etc.
			shift := uint8(7 - offset)

			// Extract pixel value (2 bits: one from low byte, one from high byte)
			pixelLow := (p.spriteShifterPatternLo[i] >> shift) & 0x01
			pixelHigh := (p.spriteShifterPatternHi[i] >> shift) & 0x01
			pixelValue := (pixelHigh << 1) | pixelLow

			// If pixel is transparent (0), skip this sprite
			if pixelValue == 0 {
				continue
			}

			// Extract palette and priority from attributes
			spritePalette := p.spriteAttributes[i] & 0x03
			spritePriority := (p.spriteAttributes[i] & 0x20) == 0 // 0 = in front, 1 = behind

			// Check if this is sprite 0 (for sprite 0 hit detection)
			sprite0 := (i == 0) && p.sprite0Present

			return pixelValue, spritePalette, spritePriority, sprite0
		}
	}

	// No sprite pixel at this position
	return 0, 0, false, false
}
