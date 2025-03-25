package widgets

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// CursorShape returns the cursor shape and whether it should be shown
type CursorShape func() (ebiten.CursorShapeType, bool)

// DefaultCursorShape returns the default cursor shape
func DefaultCursorShape() (ebiten.CursorShapeType, bool) {
	return ebiten.CursorShapeDefault, true
}

// PointerCursorShape returns the pointer cursor shape
func PointerCursorShape() (ebiten.CursorShapeType, bool) {
	return ebiten.CursorShapePointer, true
}

// TextCursorShape returns the text cursor shape
func TextCursorShape() (ebiten.CursorShapeType, bool) {
	return ebiten.CursorShapeText, true
}

// MoveCursorShape returns the move cursor shape
func MoveCursorShape() (ebiten.CursorShapeType, bool) {
	return ebiten.CursorShapeMove, true
}
