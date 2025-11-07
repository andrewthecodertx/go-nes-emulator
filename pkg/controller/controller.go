// Package controller implements NES controller (gamepad) emulation.
//
// The NES controller has 8 buttons that are read serially through
// CPU registers $4016 (controller 1) and $4017 (controller 2).
package controller

// Button represents NES controller buttons
type Button uint8

const (
	ButtonA Button = iota
	ButtonB
	ButtonSelect
	ButtonStart
	ButtonUp
	ButtonDown
	ButtonLeft
	ButtonRight
)

// Controller represents an NES controller state
type Controller struct {
	// Current button states (true = pressed)
	buttons [8]bool

	// Strobe mode - when true, button states are latched
	strobe bool

	// Index for sequential button reads (0-7)
	index uint8
}

// NewController creates a new controller
func NewController() *Controller {
	return &Controller{}
}

// SetButton sets the state of a button
func (c *Controller) SetButton(button Button, pressed bool) {
	if button < 8 {
		c.buttons[button] = pressed
	}
}

// IsPressed returns whether a button is currently pressed
func (c *Controller) IsPressed(button Button) bool {
	if button < 8 {
		return c.buttons[button]
	}
	return false
}

// Write handles writes to controller register ($4016)
// Writing 1 then 0 latches the button states for reading
func (c *Controller) Write(value uint8) {
	wasStrobe := c.strobe
	c.strobe = (value & 0x01) != 0

	// On falling edge of strobe (1 -> 0), reset index
	if wasStrobe && !c.strobe {
		c.index = 0
	}
}

// Read returns the next button state in sequence
// Returns 0 or 1 for each of the 8 buttons, then returns 1 for all subsequent reads
func (c *Controller) Read() uint8 {
	// If strobe is on, always return A button state
	if c.strobe {
		if c.buttons[ButtonA] {
			return 0x01
		}
		return 0x00
	}

	// Return current button state
	var value uint8
	if c.index < 8 {
		// Return button state for first 8 reads
		if c.buttons[c.index] {
			value = 0x01
		} else {
			value = 0x00
		}
	} else {
		// After 8 reads, always return 1
		value = 0x01
	}

	// Increment index
	c.index++
	if c.index > 23 {
		// Cap at reasonable value to prevent overflow
		c.index = 8
	}

	return value
}

// Reset resets the controller state
func (c *Controller) Reset() {
	c.strobe = false
	c.index = 0
	// Don't reset button states - they persist
}
