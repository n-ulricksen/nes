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
	gameRgba  *image.RGBA // Rectangle of RGBA points, used to manipulate pixels on the screen.
	debugRgba *image.RGBA

	window      *pixelgl.Window
	gameMatrix  pixel.Matrix // Scale and position to render the running NES game.
	debugMatrix pixel.Matrix // Scale and position to render the running NES game.
}

const (
	// Main NES display settings
	nesResW    float64 = 256
	nesResH    float64 = 240
	scale      float64 = 3 // Scale at which to render NES display.
	screenW    float64 = nesResW * scale
	screenH    float64 = nesResH * scale
	screenPosX float64 = 600 // Where to render the display on the user's monitor.
	screenPosY float64 = 400

	// Debug display settings
	debugResW float64 = 512
	debugResH float64 = screenH
)

func NewDisplay() *Display {
	rect := image.Rect(0, 0, int(nesResW), int(nesResH))
	gameRgba := image.NewRGBA(rect)
	rect = image.Rect(0, 0, int(debugResW), int(debugResH))
	debugRgba := image.NewRGBA(rect)

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
	pic := pixel.PictureDataFromImage(gameRgba)
	gameMatrix := pixel.IM.Moved(pic.Bounds().Center().Scaled(scale))
	gameMatrix = gameMatrix.Scaled(pic.Bounds().Center().Scaled(scale), scale)

	// Calculate debug window matrix used to treat (0, 0) as top-left corner of
	// the debug panel.
	pic = pixel.PictureDataFromImage(debugRgba)
	debugMatrix := pixel.IM.Moved(pic.Bounds().Center().Add(pixel.V(screenW, 0)))

	return &Display{
		gameRgba,
		debugRgba,
		window,
		gameMatrix,
		debugMatrix,
	}
}

func (d *Display) DrawPixel(x, y int, c color.RGBA) {
	d.gameRgba.SetRGBA(x, y, c)
}

func (d *Display) DrawDebugPixel(x, y int, c color.RGBA) {
	d.debugRgba.SetRGBA(x, y, c)
}

// DrawDebugRGBA draws a given image to an (x, y) offset within the debug image.
func (d *Display) DrawDebugRGBA(x, y int, img *image.RGBA) {
	for imgY := 0; imgY < img.Rect.Dy(); imgY++ {
		for imgX := 0; imgX < img.Rect.Dx(); imgX++ {
			c := img.RGBAAt(imgX, imgY)
			d.DrawDebugPixel(x+imgX, y+imgY, c)
		}
	}
}

// UpdateScreen updates both the game display and the debug display using the
// display's current image.RGBA representation of each.
func (d *Display) UpdateScreen() {
	d.window.Clear(colornames.Black)

	d.updateGameDisplay()
	d.updateDebugDisplay()

	d.window.Update()
}

func (d *Display) updateGameDisplay() {
	sprite := getSpriteFromImage(d.gameRgba)
	sprite.Draw(d.window, d.gameMatrix)
}

func (d *Display) updateDebugDisplay() {
	sprite := getSpriteFromImage(d.debugRgba)
	sprite.Draw(d.window, d.debugMatrix)
}

// Convenience function to get a pixel sprite from an image RGBA.
func getSpriteFromImage(img *image.RGBA) *pixel.Sprite {
	pic := pixel.PictureDataFromImage(img)
	sprite := pixel.NewSprite(pic, pic.Bounds())

	return sprite
}
