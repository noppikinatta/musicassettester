package ui

import (
	"fmt"
	"image"
	"log"
	"strings"

	// Keep time for potential future use in Update
	// Keep time for potential future use in Update
	// Needed for HandleFileChanges
	"musicplayer/internal/player"
	"musicplayer/internal/ui/widgets" // Keep widgets for Slider

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 400
)

// Root is the root widget of the application
type Root struct {
	guigui.DefaultWidget

	player *player.MusicPlayer

	// UI components (Value types for basicwidget again)
	musicList          basicwidget.List
	nowPlayingText     basicwidget.Text
	timeText           basicwidget.Text
	settingsText       basicwidget.Text
	loopDurationSlider widgets.Slider
	intervalSlider     widgets.Slider
	initialized        bool // 初期化フラグ
}

// NewRoot creates a new root widget
func NewRoot(player *player.MusicPlayer) *Root {
	// Initialize struct with zero values for value types and initial state
	r := &Root{
		player: player,
		// initialized is false by default
	}

	// DO NOT configure widgets here, as it might trigger RequestRedraw before Run

	return r
}

// Layout lays out the root widget
func (r *Root) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	// Configure Text widgets (Safe to call Setters here)
	r.nowPlayingText.SetBold(true)
	r.nowPlayingText.SetScale(1.5)
	r.settingsText.SetText("Settings")
	r.settingsText.SetBold(true)

	// Configure Sliders Min/Max (Safe to call Setters here)
	r.loopDurationSlider.SetMinimum(1)
	r.loopDurationSlider.SetMaximum(60)
	r.intervalSlider.SetMinimum(1)
	r.intervalSlider.SetMaximum(60)

	// --- Position and Append Widgets ---
	pos := guigui.Position(r)
	w, h := r.Size(context) // Get root size

	// Main Layout (List, Now Playing, Time, Settings, Sliders)
	listWidth := 200
	contentX := pos.X + listWidth + 20
	contentWidth := w - listWidth - 30

	// Music List
	r.musicList.SetSize(listWidth, h-20)
	// Pass ADDRESS of value types
	guigui.SetPosition(&r.musicList, image.Point{X: pos.X + 10, Y: pos.Y + 10})
	appender.AppendChildWidget(&r.musicList)

	// Now Playing Text
	r.nowPlayingText.SetSize(contentWidth, 30)
	// Pass ADDRESS of value types
	guigui.SetPosition(&r.nowPlayingText, image.Point{X: contentX, Y: pos.Y + 10})
	appender.AppendChildWidget(&r.nowPlayingText)

	// Time Text
	r.timeText.SetSize(contentWidth, 20)
	// Pass ADDRESS of value types
	guigui.SetPosition(&r.timeText, image.Point{X: contentX, Y: pos.Y + 50})
	appender.AppendChildWidget(&r.timeText)

	// Settings Text
	r.settingsText.SetSize(contentWidth, 30)
	// Pass ADDRESS of value types
	guigui.SetPosition(&r.settingsText, image.Point{X: contentX, Y: pos.Y + 100})
	appender.AppendChildWidget(&r.settingsText)

	// Loop Duration Slider (Pass pointer directly)
	r.loopDurationSlider.SetSize(contentWidth, 20)
	guigui.SetPosition(&r.loopDurationSlider, image.Point{X: contentX, Y: pos.Y + 140})
	appender.AppendChildWidget(&r.loopDurationSlider)

	// Interval Slider (Pass pointer directly)
	r.intervalSlider.SetSize(contentWidth, 20)
	guigui.SetPosition(&r.intervalSlider, image.Point{X: contentX, Y: pos.Y + 180})
	appender.AppendChildWidget(&r.intervalSlider)
}

// Size returns the size of the root widget
func (r *Root) Size(context *guigui.Context) (int, int) {
	return 800, 600
}

// Update updates the root widget
func (r *Root) Update(context *guigui.Context) error {
	// --- One-time Initialization ---
	if !r.initialized {
		r.initialize()
		r.initialized = true
	}

	// --- Regular Update Logic ---
	// Access value types directly for reads/method calls
	if err := r.player.Update(); err != nil {
		return err
	}

	currentPath := r.player.GetCurrentPath()
	if currentPath != "" {
		relPath := currentPath
		if strings.HasPrefix(relPath, "musics/") || strings.HasPrefix(relPath, "musics\\") {
			relPath = relPath[len("musics/"):]
		}
		statusText := "Now Playing: " + relPath
		if r.player.IsPaused() {
			statusText = "PAUSED: " + relPath
		}
		r.nowPlayingText.SetText(statusText) // Call method on value
	} else {
		r.nowPlayingText.SetText("No track playing")
	}

	switch r.player.GetState() {
	case player.StatePlaying:
		currentTimeSec := r.player.GetCounter() / 60
		totalTimeSec := int(r.player.GetLoopDurationMinutes() * 60)
		r.timeText.SetText(fmt.Sprintf("%d:%02d / %d:%02d",
			currentTimeSec/60, currentTimeSec%60,
			totalTimeSec/60, totalTimeSec%60))
	case player.StateFadingOut:
		r.timeText.SetText("Fading out...")
	case player.StateInterval:
		intervalSec := (int(r.player.GetIntervalSeconds())*60 - r.player.GetCounter()) / 60
		r.timeText.SetText(fmt.Sprintf("Next track in: %d seconds", intervalSec))
	default:
		r.timeText.SetText("")
	}

	r.loopDurationSlider.SetValue(float64(r.player.GetLoopDurationMinutes()))
	r.intervalSlider.SetValue(float64(r.player.GetIntervalSeconds()))

	return nil
}

// initialize performs the one-time setup for the root widget.
// This should be called only once from Update.
func (r *Root) initialize() {
	// Configure List OnItemSelected callback
	r.musicList.SetOnItemSelected(func(index int) {
		musicFiles := r.player.GetMusicFiles()
		if index >= 0 && index < len(musicFiles) {
			if err := r.player.SetCurrentIndex(index); err != nil {
				log.Printf("Failed to set current index: %v", err)
			}
		}
	})

	// Set initial slider values
	r.loopDurationSlider.SetValue(float64(r.player.GetLoopDurationMinutes()))
	r.intervalSlider.SetValue(float64(r.player.GetIntervalSeconds()))
	// Initial population of the list
	r.updateMusicList(r.player.GetMusicFiles())
}

// updateMusicList updates the music list widget
// Called by HandleFileChanges and initialize
func (r *Root) updateMusicList(musicFiles []string) {
	// Access value type directly
	listItems := make([]basicwidget.ListItem, 0, len(musicFiles))

	for _, path := range musicFiles {
		relPath := path
		if strings.HasPrefix(path, "musics/") || strings.HasPrefix(path, "musics\\") {
			relPath = path[len("musics/"):]
		}

		textWidget := &basicwidget.Text{} // Create pointer for ListItem content
		textWidget.SetText(relPath)

		item := basicwidget.ListItem{
			Content: textWidget, // ListItem still needs a Widget (pointer)
		}
		listItems = append(listItems, item)
	}

	// Call method on value type
	r.musicList.SetItems(listItems)
}

// Draw draws the root widget
func (r *Root) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Draw background using basicwidget helper
	basicwidget.FillBackground(dst, context)

	// Child widget drawing is handled by guigui automatically after Layout appends them
}

// CursorShape returns the cursor shape for this widget
func (r *Root) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	return ebiten.CursorShapeDefault, true
}

// HandleInput handles global key presses
func (r *Root) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	// Space key to toggle pause
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		r.player.TogglePause()
		return guigui.HandleInputByWidget(r) // Input handled by this widget
	}

	// N key to skip to next track
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		if err := r.player.SkipToNext(); err != nil {
			log.Printf("Failed to skip to next track: %v", err)
		}
		return guigui.HandleInputByWidget(r) // Input handled by this widget
	}

	// If not handled, return zero value to let guigui propagate to children
	return guigui.HandleInputResult{}
}

// HandleFileChanges is the event handler for directory changes.
func (r *Root) HandleFileChanges(musicFiles []string) {
	// Update the music list UI
	r.updateMusicList(musicFiles)

	// Request redraw or relayout if needed (might be handled by guigui automatically)
	// guigui.RequestLayout(r)
}
