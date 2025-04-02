package ui

import (
	"fmt"
	"image"
	"log"
	"strings"

	// Keep time for potential future use in Update
	// Keep time for potential future use in Update
	"musicplayer/internal/files" // Needed for HandleFileChanges
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
	// warningText string // Removed, use warningLabel instead

	// UI components (Value types for basicwidget again)
	musicList          basicwidget.List
	nowPlayingText     basicwidget.Text
	timeText           basicwidget.Text
	settingsText       basicwidget.Text
	loopDurationSlider widgets.Slider
	intervalSlider     widgets.Slider
	warningLabel       basicwidget.Text
	warningText        string // 警告テキストの保持用
}

// NewRoot creates a new root widget
func NewRoot(player *player.MusicPlayer, initialWarningText string) *Root {
	// Initialize struct with zero values for value types
	r := &Root{
		player:      player,
		warningText: initialWarningText,
	}

	// Configure List (Access value type directly)
	r.musicList.SetOnItemSelected(func(index int) {
		if r.player != nil {
			musicFiles := r.player.GetMusicFiles()
			if index >= 0 && index < len(musicFiles) {
				if err := r.player.SetCurrentIndex(index); err != nil {
					log.Printf("Failed to set current index: %v", err)
				}
			}
		}
	})

	// Configure Sliders (Pointers are accessed directly)
	r.loopDurationSlider.SetMinimum(1)
	r.loopDurationSlider.SetMaximum(60)
	r.intervalSlider.SetMinimum(1)
	r.intervalSlider.SetMaximum(60)
	if r.player != nil {
		r.loopDurationSlider.SetValue(float64(r.player.GetLoopDurationMinutes()))
		r.intervalSlider.SetValue(float64(r.player.GetIntervalSeconds()))
	}

	// Initial population of the list
	if r.player != nil {
		r.updateMusicList(r.player.GetMusicFiles())
	}

	return r
}

// Layout lays out the root widget
func (r *Root) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	// Configure Text widgets
	r.nowPlayingText.SetBold(true)
	r.nowPlayingText.SetScale(1.5)
	r.settingsText.SetText("Settings")
	r.settingsText.SetBold(true)
	r.warningLabel.SetText(r.warningText) // 保持している値を設定
	r.warningLabel.SetScale(1.2)

	// --- Position and Append Widgets ---
	pos := guigui.Position(r)
	w, h := r.Size(context) // Get root size

	// Conditionally add EITHER warning label OR the main layout
	if r.warningLabel.Text() != "" {
		// Warning Label takes up main space
		r.warningLabel.SetSize(w-40, h-40)
		// Pass ADDRESS of value types
		guigui.SetPosition(&r.warningLabel, image.Point{X: pos.X + 20, Y: pos.Y + 20})
		appender.AppendChildWidget(&r.warningLabel)
	} else {
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
}

// Size returns the size of the root widget
func (r *Root) Size(context *guigui.Context) (int, int) {
	return 800, 600
}

// Update updates the root widget
func (r *Root) Update(context *guigui.Context) error {
	// Access value types directly for reads/method calls
	if r.player != nil {
		if err := r.player.Update(); err != nil {
			return err
		}

		if r.warningLabel.Text() == "" {
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
		}

	} else {
		if r.warningLabel.Text() == "" {
			r.warningLabel.SetText("Player Initialization Error")
		}
	}

	return nil
}

// updateMusicList updates the music list widget
// Called by HandleFileChanges
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
		if r.player != nil {
			r.player.TogglePause()
			return guigui.HandleInputByWidget(r) // Input handled by this widget
		}
	}

	// N key to skip to next track
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		if r.player != nil {
			if err := r.player.SkipToNext(); err != nil {
				log.Printf("Failed to skip to next track: %v", err)
			}
			return guigui.HandleInputByWidget(r) // Input handled by this widget
		}
	}

	// If not handled, return zero value to let guigui propagate to children
	return guigui.HandleInputResult{}
}

// HandleFileChanges is the event handler for directory changes.
func (r *Root) HandleFileChanges(musicFiles []string) {
	// Update the music list UI
	r.updateMusicList(musicFiles)

	// Update the warning label based on whether files exist (Access value type directly)
	if len(musicFiles) == 0 {
		r.warningLabel.SetText(files.DefaultMusicDir.GetUsageInstructions())
	} else {
		r.warningLabel.SetText("") // Clear warning if files exist
	}

	// Request redraw or relayout if needed (might be handled by guigui automatically)
	// guigui.RequestLayout(r)
}
