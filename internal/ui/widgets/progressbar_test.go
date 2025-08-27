package widgets_test

import (
	"testing"

	"github.com/hajimehoshi/guigui"
	"github.com/stretchr/testify/assert"

	"musicplayer/internal/ui/widgets"
)

// MockContext is a test context
type MockContext struct {
	guigui.DefaultWidget
}

func TestNewProgressBar(t *testing.T) {
	t.Parallel()

	pb := widgets.NewProgressBar()
	assert.NotNil(t, pb)
	assert.Equal(t, 0.0, pb.Value())

	w, h := pb.Size(nil)
	assert.Equal(t, 100, w)
	assert.Equal(t, 20, h)
}

func TestProgressBar_SetValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"normal value", 0.5, 0.5},
		{"minimum bound", -0.1, 0.0},
		{"maximum bound", 1.1, 1.0},
		{"zero", 0.0, 0.0},
		{"one", 1.0, 1.0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := widgets.NewProgressBar()
			pb.SetValue(tt.input)
			assert.Equal(t, tt.expected, pb.Value())
		})
	}
}

func TestProgressBar_SetSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		width, height int
	}{
		{"default size", 100, 20},
		{"custom size", 200, 30},
		{"zero size", 0, 0},
		{"negative size", -10, -10}, // Negative values are allowed (implementation dependent)
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pb := widgets.NewProgressBar()
			pb.SetSize(tt.width, tt.height)
			w, h := pb.Size(nil)
			assert.Equal(t, tt.width, w)
			assert.Equal(t, tt.height, h)
		})
	}
}
