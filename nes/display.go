package nes

import (
	"image"
	"image/color"
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

type Display struct {
	rgba *image.RGBA // Rectangle of RGBA points, used to manipulate pixels on the screen.

	window     *pixelgl.Window
	gameMatrix pixel.Matrix // Scale and position to render the running NES game.
}

const (
	// Main NES display settings
	nesResW    float64 = 256
	nesResH    float64 = 240
	scale      float64 = 2 // Scale at which to render NES display.
	screenW    float64 = nesResW * scale
	screenH    float64 = nesResH * scale
	screenPosX float64 = 600 // Where to render the display on the user's monitor.
	screenPosY float64 = 400

	// Debug display settings
	debugResW float64 = 256

	debugScreenPosX float64 = screenPosX + scale*nesResW
	debugScreenPosY float64 = screenPosY
)

func NewDisplay() *Display {
	rect := image.Rect(0, 0, int(nesResW), int(nesResH))
	rgba := image.NewRGBA(rect)

	config := pixelgl.WindowConfig{
		Title:    "NES Emulator",
		Bounds:   pixel.R(0, 0, screenW+debugResW, screenH),
		Position: pixel.V(screenPosX, screenPosY),
		VSync:    true,
	}
	window, err := pixelgl.NewWindow(config)
	if err != nil {
		log.Fatal("Unable to create new PixelGl window...\n", err)
	}

	// Calculate matrix recquired to render game to display based on the set scale.
	pic := pixel.PictureDataFromImage(rgba)

	matrix := pixel.IM.Moved(pic.Bounds().Center().Scaled(scale))
	matrix = matrix.Scaled(pic.Bounds().Center().Scaled(scale), scale)

	return &Display{
		rgba,
		window,
		matrix,
	}
}

func (d *Display) DrawPixel(x, y int, c color.RGBA) {
	d.rgba.SetRGBA(x, y, c)
}

func (d *Display) UpdateScreen() {
	d.window.Clear(colornames.Black)

	pic := pixel.PictureDataFromImage(d.rgba)

	sprite := pixel.NewSprite(pic, pic.Bounds())
	sprite.Draw(d.window, d.gameMatrix)

	d.window.Update()
}
