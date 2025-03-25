package widgets

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/guigui"
)

// Text is a widget that displays text
type Text struct {
	guigui.DefaultWidget

	text  string
	bold  bool
	scale float64
	size  image.Point
}

// NewText creates a new text widget
func NewText(text string) *Text {
	return &Text{
		text:  text,
		scale: 1.0,
		size:  image.Point{X: 100, Y: 20},
	}
}

// SetText sets the text content
func (t *Text) SetText(text string) {
	t.text = text
}

// SetBold sets whether the text is bold
func (t *Text) SetBold(bold bool) {
	t.bold = bold
}

// SetScale sets the text scale
func (t *Text) SetScale(scale float64) {
	t.scale = scale
}

// SetSize sets the size of the text widget
func (t *Text) SetSize(size image.Point) {
	t.size = size
}

// Draw draws the text widget
func (t *Text) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Get the widget position
	pos := guigui.Position(t)

	// Create options for drawing
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(pos.X), float64(pos.Y))
	opts.GeoM.Scale(t.scale, t.scale)

	// TODO: Draw text using ebiten.Font
	// For now, just draw a placeholder rectangle
	rect := ebiten.NewImage(t.size.X, t.size.Y)
	textColor, _, _ := Colors()
	rect.Fill(textColor)
	dst.DrawImage(rect, opts)
}

// Size returns the size of the text widget
func (t *Text) Size(context *guigui.Context) (int, int) {
	return t.size.X, t.size.Y
}

// Layout lays out the text widget
func (t *Text) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	// Text widget has no child widgets
}

// Update updates the text widget
func (t *Text) Update(context *guigui.Context) error {
	return nil
}

// CursorShape returns the cursor shape for this widget
func (t *Text) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	return ebiten.CursorShapeDefault, true
} 