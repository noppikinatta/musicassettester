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
	rootWidth, rootHeight := r.Size(context) // Get root size

	const margin int = 8

	// ウィジェットの配置計算
	// 利用可能な幅はRootの幅からmarginを両側分引いたもの
	availableWidth := rootWidth - margin*2

	// 各ウィジェットの高さを定義
	const (
		nowPlayingTextHeight = 30
		timeTextHeight       = 20
		settingsTextHeight   = 30
		sliderHeight         = 20
	)

	// ウィジェットの縦方向の配置を下から順に計算
	// intervalSlider
	intervalSliderY := rootHeight - margin - sliderHeight

	// loopDurationSlider
	loopDurationSliderY := intervalSliderY - margin - sliderHeight

	// settingsText
	settingsTextY := loopDurationSliderY - margin - settingsTextHeight

	// timeText
	timeTextY := settingsTextY - margin - timeTextHeight

	// nowPlayingText
	nowPlayingTextY := timeTextY - margin - nowPlayingTextHeight

	// musicList （残りの高さを全て使用）
	musicListHeight := nowPlayingTextY - margin*2
	musicListY := margin

	// ウィジェットの配置と追加
	// Music List
	r.musicList.SetSize(availableWidth, musicListHeight)
	guigui.SetPosition(&r.musicList, image.Point{X: pos.X + margin, Y: pos.Y + musicListY})
	appender.AppendChildWidget(&r.musicList)

	// Now Playing Text
	r.nowPlayingText.SetSize(availableWidth, nowPlayingTextHeight)
	guigui.SetPosition(&r.nowPlayingText, image.Point{X: pos.X + margin, Y: pos.Y + nowPlayingTextY})
	appender.AppendChildWidget(&r.nowPlayingText)

	// Time Text
	r.timeText.SetSize(availableWidth, timeTextHeight)
	guigui.SetPosition(&r.timeText, image.Point{X: pos.X + margin, Y: pos.Y + timeTextY})
	appender.AppendChildWidget(&r.timeText)

	// Settings Text
	r.settingsText.SetSize(availableWidth, settingsTextHeight)
	guigui.SetPosition(&r.settingsText, image.Point{X: pos.X + margin, Y: pos.Y + settingsTextY})
	appender.AppendChildWidget(&r.settingsText)

	// Loop Duration Slider
	r.loopDurationSlider.SetSize(availableWidth, sliderHeight)
	guigui.SetPosition(&r.loopDurationSlider, image.Point{X: pos.X + margin, Y: pos.Y + loopDurationSliderY})
	appender.AppendChildWidget(&r.loopDurationSlider)

	// Interval Slider
	r.intervalSlider.SetSize(availableWidth, sliderHeight)
	guigui.SetPosition(&r.intervalSlider, image.Point{X: pos.X + margin, Y: pos.Y + intervalSliderY})
	appender.AppendChildWidget(&r.intervalSlider)
}

// Size returns the size of the root widget
func (r *Root) Size(context *guigui.Context) (int, int) {
	return context.AppSize()
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

	r.updateCurrentMusicState()

	r.loopDurationSlider.SetValue(float64(r.player.GetLoopDurationMinutes()))
	r.intervalSlider.SetValue(float64(r.player.GetIntervalSeconds()))

	return nil
}

// updateCurrentMusicState updates the UI elements related to the current music state.
func (r *Root) updateCurrentMusicState() {
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
		r.nowPlayingText.SetText("No track playing. Locate music files in musics/ directory.")
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

	// Set initial slider values and configure callbacks
	r.loopDurationSlider.SetValue(float64(r.player.GetLoopDurationMinutes()))
	r.loopDurationSlider.SetOnChange(func(value float64) {
		r.player.SetLoopDurationMinutes(value)
	})

	r.intervalSlider.SetValue(float64(r.player.GetIntervalSeconds()))
	r.intervalSlider.SetOnChange(func(value float64) {
		r.player.SetIntervalSeconds(value)
	})

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
