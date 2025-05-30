package widgets

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/guigui"
)

// Slider is a widget for selecting a value within a range.
type Slider struct {
	guigui.DefaultWidget

	value      float64
	minimum    float64
	maximum    float64
	width      int
	height     int
	onChange   func(float64)
	isDragging bool
}

// NewSlider creates a new slider with default values.
func NewSlider() *Slider {
	return &Slider{
		value:   0,
		minimum: 0,
		maximum: 100,
		width:   200,
		height:  20,
	}
}

// SetValue sets the current value of the slider.
func (s *Slider) SetValue(value float64) {
	// Clamp value between minimum and maximum
	if value < s.minimum {
		value = s.minimum
	}
	if value > s.maximum {
		value = s.maximum
	}

	if s.value != value {
		s.value = value
		guigui.RequestRedraw(s)
		if s.onChange != nil {
			s.onChange(value)
		}
	}
}

// SetMinimum sets the minimum value of the slider.
func (s *Slider) SetMinimum(min float64) {
	s.minimum = min
	guigui.RequestRedraw(s)
	if s.value < min {
		s.SetValue(min)
	}
}

// SetMaximum sets the maximum value of the slider.
func (s *Slider) SetMaximum(max float64) {
	s.maximum = max
	guigui.RequestRedraw(s)
	if s.value > max {
		s.SetValue(max)
	}
}

// SetOnChange sets the callback function that is called when the value changes.
func (s *Slider) SetOnChange(callback func(float64)) {
	s.onChange = callback
}

// Value returns the current value of the slider.
func (s *Slider) Value() float64 {
	return s.value
}

// SetSize sets the size of the slider.
func (s *Slider) SetSize(width, height int) {
	s.width = width
	s.height = height
	guigui.RequestRedraw(s)
}

// Size returns the size of the slider.
func (s *Slider) Size(context *guigui.Context) (int, int) {
	return s.width, s.height
}

// Draw draws the slider.
func (s *Slider) Draw(context *guigui.Context, dst *ebiten.Image) {
	bounds := context.Bounds(s)

	// Draw background
	bgColor := color.RGBA{200, 200, 200, 255}
	vector.DrawFilledRect(dst, float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy()), bgColor, false)

	// --- 元のハンドル描画 ---
	valueRange := s.maximum - s.minimum
	if valueRange == 0 {
		valueRange = 1 // Or handle as a special case
	}
	valueRatio := (s.value - s.minimum) / valueRange
	// Clamp valueRatio to avoid drawing outside bounds (just in case)
	if valueRatio < 0 {
		valueRatio = 0
	} else if valueRatio > 1 {
		valueRatio = 1
	}
	handleX := float32(bounds.Min.X) + float32(bounds.Dx())*float32(valueRatio)
	handleY := float32(bounds.Min.Y)
	handleWidth := float32(10)
	handleHeight := float32(bounds.Dy())

	// Draw handle
	handleColor := color.RGBA{100, 100, 100, 255}
	vector.DrawFilledRect(dst, handleX-handleWidth/2, handleY, handleWidth, handleHeight, handleColor, false)
	// ---
}

// Layout lays out the slider.
func (s *Slider) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	// Slider has no children
}

// Update updates the slider.
func (s *Slider) Update(context *guigui.Context) error {
	bounds := context.Bounds(s)
	x, y := ebiten.CursorPosition()

	// Check if mouse is over slider
	if x >= bounds.Min.X && x < bounds.Max.X &&
		y >= bounds.Min.Y && y < bounds.Max.Y {

		// Start dragging on mouse press
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			s.isDragging = true
		}
	}

	// Update value while dragging
	if s.isDragging {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			// Calculate new value based on mouse position
			valueRange := s.maximum - s.minimum
			valueRatio := float64(x-bounds.Min.X) / float64(bounds.Dx())
			if valueRatio < 0 {
				valueRatio = 0
			}
			if valueRatio > 1 {
				valueRatio = 1
			}
			newValue := s.minimum + valueRange*valueRatio
			s.SetValue(newValue)
		} else {
			s.isDragging = false
		}
	}

	return nil
}

// CursorShape returns the cursor shape for the slider.
func (s *Slider) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	bounds := context.Bounds(s)
	x, y := ebiten.CursorPosition()

	// Change cursor to pointer when over slider
	if x >= bounds.Min.X && x < bounds.Max.X &&
		y >= bounds.Min.Y && y < bounds.Max.Y {
		return ebiten.CursorShapePointer, true
	}

	return ebiten.CursorShapeDefault, true
}
