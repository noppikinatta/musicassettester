package widgets

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/guigui"
)

// ProgressBar is a custom widget for displaying progress
type ProgressBar struct {
	guigui.DefaultWidget

	value  float64
	width  int
	height int
}

// NewProgressBar creates a new progress bar
func NewProgressBar() *ProgressBar {
	return &ProgressBar{
		value:  0,
		width:  100,
		height: 20,
	}
}

// SetValue sets the progress value (0.0 to 1.0)
func (p *ProgressBar) SetValue(value float64) {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	p.value = value
}

// Value returns the progress value
func (p *ProgressBar) Value() float64 {
	return p.value
}

// SetSize sets the size of the progress bar
func (p *ProgressBar) SetSize(width, height int) {
	p.width = width
	p.height = height
}

// Size returns the size of the progress bar
func (p *ProgressBar) Size(context *guigui.Context) (int, int) {
	return p.width, p.height
}

// Draw draws the progress bar
func (p *ProgressBar) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Draw background
	x, y := guigui.Position(p).X, guigui.Position(p).Y
	w, h := p.Size(context)

	// Background (gray)
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{100, 100, 100, 255}, false)

	// Progress (green)
	progressWidth := float32(float64(w) * p.value)
	if progressWidth > 0 {
		vector.DrawFilledRect(dst, float32(x), float32(y), progressWidth, float32(h), color.RGBA{0, 200, 100, 255}, false)
	}

	// Border
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1, color.RGBA{150, 150, 150, 255}, false)
}
