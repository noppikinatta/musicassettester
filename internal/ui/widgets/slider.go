package widgets

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/guigui"
)

// Slider is a custom widget for value selection
type Slider struct {
	guigui.DefaultWidget

	value    float64
	min      float64
	max      float64
	width    int
	height   int
	onChange func(float64)

	dragging       bool
	dragStartX     int
	dragStartValue float64
}

// NewSlider creates a new slider
func NewSlider() *Slider {
	return &Slider{
		value:  0,
		min:    0,
		max:    1,
		width:  100,
		height: 20,
	}
}

// SetValue sets the slider value
func (s *Slider) SetValue(value float64) {
	if value < s.min {
		value = s.min
	}
	if value > s.max {
		value = s.max
	}
	s.value = value
}

// Value returns the slider value
func (s *Slider) Value() float64 {
	return s.value
}

// SetMinimum sets the minimum value
func (s *Slider) SetMinimum(min float64) {
	s.min = min
	if s.value < min {
		s.value = min
	}
}

// SetMaximum sets the maximum value
func (s *Slider) SetMaximum(max float64) {
	s.max = max
	if s.value > max {
		s.value = max
	}
}

// SetOnChange sets the change callback
func (s *Slider) SetOnChange(f func(float64)) {
	s.onChange = f
}

// SetSize sets the size of the slider
func (s *Slider) SetSize(width, height int) {
	s.width = width
	s.height = height
}

// Size returns the size of the slider
func (s *Slider) Size(context *guigui.Context) (int, int) {
	return s.width, s.height
}

// HandleInput handles input events
func (s *Slider) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	x, y := guigui.Position(s).X, guigui.Position(s).Y
	w, h := s.Size(context)

	mx, my := ebiten.CursorPosition()

	// Check if mouse is over the slider
	if mx >= x && mx < x+w && my >= y && my < y+h {
		// Start dragging on mouse press
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			s.dragging = true
			s.dragStartX = mx
			s.dragStartValue = s.value
			return guigui.HandleInputByWidget(s)
		}
	}

	// Handle dragging
	if s.dragging {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			relativeX := float64(mx-x) / float64(w)
			newValue := s.min + relativeX*(s.max-s.min)

			// Clamp value
			if newValue < s.min {
				newValue = s.min
			}
			if newValue > s.max {
				newValue = s.max
			}

			// Only call onChange if value actually changed
			if newValue != s.value {
				s.value = newValue
				if s.onChange != nil {
					s.onChange(s.value)
				}
			}

			return guigui.HandleInputByWidget(s)
		} else {
			// Stop dragging when button is released
			s.dragging = false
		}
	}

	return guigui.HandleInputResult{}
}

// Draw draws the slider
func (s *Slider) Draw(context *guigui.Context, dst *ebiten.Image) {
	x, y := guigui.Position(s).X, guigui.Position(s).Y
	w, h := s.Size(context)

	// Draw track
	trackY := y + h/2 - 2
	trackHeight := 4
	vector.DrawFilledRect(dst, float32(x), float32(trackY), float32(w), float32(trackHeight), color.RGBA{100, 100, 100, 255}, false)

	// Calculate thumb position
	valueRatio := (s.value - s.min) / (s.max - s.min)
	thumbX := float32(x + int(float64(w)*valueRatio))
	thumbSize := float32(12)
	thumbY := float32(y + h/2 - int(thumbSize)/2)

	// Draw thumb
	vector.DrawFilledCircle(dst, thumbX, thumbY+thumbSize/2, thumbSize/2, color.RGBA{200, 200, 200, 255}, false)
	vector.StrokeCircle(dst, thumbX, thumbY+thumbSize/2, thumbSize/2, 1, color.RGBA{150, 150, 150, 255}, false)
}
