package widgets_test

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/stretchr/testify/assert"

	"musicplayer/internal/ui/widgets"
)

func TestNewSlider(t *testing.T) {
	t.Parallel()

	s := widgets.NewSlider()
	assert.NotNil(t, s)
	assert.Equal(t, 0.0, s.Value())

	w, h := s.Size(nil)
	assert.Equal(t, 200, w)
	assert.Equal(t, 20, h)
}

func TestSlider_SetValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		min      float64
		max      float64
		input    float64
		expected float64
	}{
		{"normal value", 0, 1, 0.5, 0.5},
		{"minimum bound", 0, 1, -0.1, 0.0},
		{"maximum bound", 0, 1, 1.1, 1.0},
		{"custom range min", -10, 10, -15, -10},
		{"custom range max", -10, 10, 15, 10},
		{"custom range normal", -10, 10, 0, 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := widgets.NewSlider()
			s.SetMinimum(tt.min)
			s.SetMaximum(tt.max)
			s.SetValue(tt.input)
			assert.Equal(t, tt.expected, s.Value())
		})
	}
}

func TestSlider_SetMinimum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		initialVal  float64
		min         float64
		expectedVal float64
	}{
		{"normal case", 5, 0, 5},
		{"value below new min", 5, 10, 10},
		{"negative min", 0, -10, 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := widgets.NewSlider()
			s.SetMaximum(20) // Set a sufficiently large maximum value
			s.SetValue(tt.initialVal)
			s.SetMinimum(tt.min)
			assert.Equal(t, tt.expectedVal, s.Value())
		})
	}
}

func TestSlider_SetMaximum(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		initialVal  float64
		max         float64
		expectedVal float64
	}{
		{"normal case", 5, 10, 5},
		{"value above new max", 15, 10, 10},
		{"negative max", 0, -10, -10},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := widgets.NewSlider()
			s.SetMinimum(-20) // Set a sufficiently small minimum value
			s.SetMaximum(20)  // Set a sufficiently large maximum value
			s.SetValue(tt.initialVal)
			s.SetMaximum(tt.max)
			assert.Equal(t, tt.expectedVal, s.Value())
		})
	}
}

func TestSlider_SetOnChange(t *testing.T) {
	t.Parallel()

	var called bool
	var lastValue float64

	s := widgets.NewSlider()
	s.SetOnChange(func(v float64) {
		called = true
		lastValue = v
	})

	// Test that callback is called when value changes
	s.SetValue(0.5)
	assert.True(t, called)
	assert.Equal(t, 0.5, lastValue)
}

func TestSlider_SetSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		width, height int
	}{
		{"default size", 100, 20},
		{"custom size", 200, 30},
		{"zero size", 0, 0},
		{"negative size", -10, -10},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := widgets.NewSlider()
			s.SetSize(tt.width, tt.height)
			w, h := s.Size(nil)
			assert.Equal(t, tt.width, w)
			assert.Equal(t, tt.height, h)
		})
	}
}

func TestSlider_HandleInput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped: Input tests are not run in short mode")
	}

	s := widgets.NewSlider()

	// Test input handling
	result := s.HandleInput(nil)
	assert.NotNil(t, result)
}

func TestSlider_Draw(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipped: Drawing tests are not run in short mode")
	}

	s := widgets.NewSlider()
	img := ebiten.NewImage(200, 50)

	// Test slider at 50% position
	s.SetValue(0.5)
	s.SetSize(200, 50)
	s.Draw(nil, img)

	// Verify drawing result
	assert.NotNil(t, img)

	// Test boundary values
	s.SetValue(0.0)
	s.Draw(nil, img)
	s.SetValue(1.0)
	s.Draw(nil, img)
}
