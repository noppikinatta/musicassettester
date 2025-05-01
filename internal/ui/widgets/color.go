package widgets

import (
	"image/color"
)

// Colors returns the default colors for the UI
func Colors() (text, background, highlight color.Color) {
	text = color.White
	background = color.Black
	highlight = color.RGBA{R: 0x33, G: 0x33, B: 0x33, A: 0xFF}
	return
}
