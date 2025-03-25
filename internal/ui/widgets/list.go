package widgets

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/guigui"
)

// ListItem represents an item in a list
type ListItem struct {
	Content    *Text
	Selectable bool
	Tag        interface{}
}

// List is a widget that displays a list of items
type List struct {
	guigui.DefaultWidget

	items         []ListItem
	selectedIndex int
	size         image.Point
	onItemSelected func(index int)
}

// NewList creates a new list widget
func NewList() *List {
	return &List{
		selectedIndex: -1,
		size:         image.Point{X: 200, Y: 100},
	}
}

// Items returns the list items
func (l *List) Items() []ListItem {
	return l.items
}

// SetItems sets the list items
func (l *List) SetItems(items []ListItem) {
	l.items = items
}

// SelectedIndex returns the selected item index
func (l *List) SelectedIndex() int {
	return l.selectedIndex
}

// SetSelectedIndex sets the selected item index
func (l *List) SetSelectedIndex(index int) {
	if index >= -1 && index < len(l.items) {
		l.selectedIndex = index
		if l.onItemSelected != nil {
			l.onItemSelected(index)
		}
	}
}

// SetOnItemSelected sets the callback for when an item is selected
func (l *List) SetOnItemSelected(callback func(index int)) {
	l.onItemSelected = callback
}

// SetSize sets the size of the list widget
func (l *List) SetSize(size image.Point) {
	l.size = size
}

// Draw draws the list widget
func (l *List) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Get the widget position
	pos := guigui.Position(l)

	// Create options for drawing
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(pos.X), float64(pos.Y))

	// Draw background
	rect := ebiten.NewImage(l.size.X, l.size.Y)
	_, bgColor, _ := Colors()
	rect.Fill(bgColor)
	dst.DrawImage(rect, opts)

	// Draw items
	itemHeight := 20
	for i, item := range l.items {
		// Draw item background if selected
		if i == l.selectedIndex {
			itemRect := ebiten.NewImage(l.size.X, itemHeight)
			_, _, hlColor := Colors()
			itemRect.Fill(hlColor)
			itemOpts := &ebiten.DrawImageOptions{}
			itemOpts.GeoM.Translate(float64(pos.X), float64(pos.Y+i*itemHeight))
			dst.DrawImage(itemRect, itemOpts)
		}

		// Draw item content
		if item.Content != nil {
			item.Content.SetSize(image.Point{X: l.size.X, Y: itemHeight})
			contentOpts := &ebiten.DrawImageOptions{}
			contentOpts.GeoM.Translate(float64(pos.X), float64(pos.Y+i*itemHeight))
			item.Content.Draw(context, dst)
		}
	}
}

// Size returns the size of the list widget
func (l *List) Size(context *guigui.Context) (int, int) {
	return l.size.X, l.size.Y
}

// Layout lays out the list widget
func (l *List) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	// List widget has no child widgets
}

// Update updates the list widget
func (l *List) Update(context *guigui.Context) error {
	return nil
}

// CursorShape returns the cursor shape for this widget
func (l *List) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	return ebiten.CursorShapeDefault, true
} 