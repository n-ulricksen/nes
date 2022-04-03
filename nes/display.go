package nes

import (
	"image"
	"image/color"
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

type Display struct {
	gameRgba  *image.RGBA // Rectangle of RGBA points, used to manipulate pixels on the screen.
	debugRgba *image.RGBA

	window      *pixelgl.Window
	gameMatrix  pixel.Matrix // Scale and position to render the running NES game.
	debugMatrix pixel.Matrix // Scale and position to render the running NES game.

	// Debug text stuff
	debugAtlas          *text.Atlas // Used to load the font
	debugRegText        *text.Text  // CPU register printout
	debugInstText       *text.Text  // CPU instruction disassembly
	debugControllerText *text.Text  // Controller input status

	isDebug bool // Debug mode enabled on the NES
}

const (
	// Main NES display settings
	nesResW    float64 = 256
	nesResH    float64 = 240
	scale      float64 = 3 // Scale at which to render NES display.
	gameW      float64 = nesResW * scale
	gameH      float64 = nesResH * scale
	screenPosX float64 = 600 // Where to render the display on the user's monitor.
	screenPosY float64 = 400

	// Debug display settings
	debugResW float64 = 512
	debugResH float64 = gameH
)

func NewDisplay(isDebug bool) *Display {
	rect := image.Rect(0, 0, int(nesResW), int(nesResH))
	gameRgba := image.NewRGBA(rect)

	rect = image.Rect(0, 0, int(debugResW), int(debugResH))
	debugRgba := image.NewRGBA(rect)

	screenW := gameW
	if isDebug {
		screenW += debugResW
	}

	config := pixelgl.WindowConfig{
		Title:    "NES Emulator",
		Bounds:   pixel.R(0, 0, screenW, gameH),
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
	debugMatrix := pixel.IM.Moved(pic.Bounds().Center().Add(pixel.V(gameW, 0)))

	// Debug text
	debugAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	debugRegText := text.New(pixel.V(gameW+8, gameH-40), debugAtlas)
	debugInstText := text.New(pixel.V(gameW+8, gameH-180), debugAtlas)
	debugControllerText := text.New(pixel.V(gameW+300, gameH-40), debugAtlas)

	return &Display{
		gameRgba,
		debugRgba,
		window,
		gameMatrix,
		debugMatrix,
		debugAtlas,
		debugRegText,
		debugInstText,
		debugControllerText,
		isDebug,
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

// Write a string of text to the CPU register section of the debug panel.
func (d *Display) WriteRegDebugString(t string) {
	d.debugRegText.Clear()
	d.debugRegText.WriteString(t)
}

// Write a string of text to the isntruction disassembly section of the debug panel.
func (d *Display) WriteInstDebugString(t string) {
	d.debugInstText.Clear()
	d.debugInstText.WriteString(t)
}

// Write a string of text to the controller status section of the debug panel.
func (d *Display) WriteControllerDebugString(t string) {
	d.debugControllerText.Clear()
	d.debugControllerText.WriteString(t)
}

// UpdateScreen updates both the game display and the debug display using the
// display's current image.RGBA representation of each.
func (d *Display) UpdateScreen() {
	d.window.Clear(colornames.Black)

	d.updateGameDisplay()

	// Update debug panel as well.
	if d.isDebug {
		d.updateDebugDisplay()
		d.debugRegText.Draw(d.window, pixel.IM)
		d.debugInstText.Draw(d.window, pixel.IM)
		d.debugControllerText.Draw(d.window, pixel.IM)
	}

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
