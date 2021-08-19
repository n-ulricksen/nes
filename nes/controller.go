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
	0: A      ---> J
	1: B      ---> K
	2: Select ---> Right Shift
	3: Start  ---> Enter
	4: Up     ---> W
	5: Down   ---> S
	6: Left   ---> A
	7: Right  ---> D
*/
const (
	keyA int = iota
	keyB
	keySelect
	keyStart
	keyUp
	keyDown
	keyLeft
	keyRight
)

var controllerKeys = map[int]pixelgl.Button{
	keyA:      pixelgl.KeyJ,
	keyB:      pixelgl.KeyK,
	keySelect: pixelgl.KeyRightShift,
	keyStart:  pixelgl.KeyEnter,
	keyUp:     pixelgl.KeyW,
	keyDown:   pixelgl.KeyS,
	keyLeft:   pixelgl.KeyA,
	keyRight:  pixelgl.KeyD,
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
