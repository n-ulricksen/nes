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

	window *pixelgl.Window
}

const (
	nesResW int     = 256
	nesResH int     = 240
	scale   float64 = 2 // Scale at which to render NES display.
)

func NewDisplay() *Display {
	rect := image.Rect(0, 0, nesResW, nesResH)
	rgba := image.NewRGBA(rect)

	config := pixelgl.WindowConfig{
		Title:  "NES Emulator",
		Bounds: pixel.R(0, 0, float64(nesResW)*scale, float64(nesResH)*scale),
		VSync:  true,
	}
	window, err := pixelgl.NewWindow(config)
	if err != nil {
		log.Fatal("Unable to create new PixelGl window...\n", err)
	}

	return &Display{
		rgba,
		window,
	}
}

func (d *Display) DrawPixel(x, y int, c color.RGBA) {
	d.rgba.SetRGBA(x, y, c)
}

func (d *Display) UpdateScreen() {
	d.window.Clear(colornames.Black)

	pic := pixel.PictureDataFromImage(d.rgba)

	matrix := pixel.IM.Moved(d.window.Bounds().Center())
	matrix = matrix.Scaled(d.window.Bounds().Center(), scale)

	sprite := pixel.NewSprite(pic, pic.Bounds())
	sprite.Draw(d.window, matrix)

	d.window.Update()
}
