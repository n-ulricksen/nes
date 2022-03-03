package nes

import (
	"github.com/faiface/pixel/pixelgl"
)

type Controller struct {
	buttonState []bool // Key press state: on/off
}

func NewController() *Controller {
	return &Controller{
		buttonState: make([]bool, len(controllerKeys)),
	}
}

// Available NES controller buttons and their keyboard binds
// Keyboard binds:
/*
	0: Right  ---> D
	1: Left   ---> A
	2: Down   ---> S
	3: Up     ---> W
	4: Start  ---> Enter
	5: Select ---> Right Shift
	6: B      ---> K
	7: A      ---> J
*/
const (
	keyRight int = iota
	keyLeft
	keyDown
	keyUp
	keyStart
	keySelect
	keyB
	keyA
)

var controllerKeys = map[int]pixelgl.Button{
	keyRight:  pixelgl.KeyD,
	keyLeft:   pixelgl.KeyA,
	keyDown:   pixelgl.KeyS,
	keyUp:     pixelgl.KeyW,
	keyStart:  pixelgl.KeyEnter,
	keySelect: pixelgl.KeyRightShift,
	keyB:      pixelgl.KeyK,
	keyA:      pixelgl.KeyJ,
}

// GetState returns a byte, with each bit representing the state of a button on
// the controller.
func (c *Controller) GetState() byte {
	var state byte

	for pos, s := range c.buttonState {
		if s {
			state |= (1 << pos)
		}
	}

	return state
}

func (c *Controller) updateControllerInput(win *pixelgl.Window) {
	// Key down
	for idx, key := range controllerKeys {
		if win.JustPressed(key) {
			c.buttonState[idx] = true
		}
	}
	// Key up
	for idx, key := range controllerKeys {
		if win.JustReleased(key) {
			c.buttonState[idx] = false
		}
	}
}
