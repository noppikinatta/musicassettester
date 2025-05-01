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
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
	"github.com/hajimehoshi/guigui/basicwidget/cjkfont"
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
	background basicwidget.Background
	musicList          basicwidget.TextList[string]
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
func (r *Root) Build(context *guigui.Context, appender *guigui.ChildWidgetAppender) error{
	faceSources := []*text.GoTextFaceSource{
		basicwidget.DefaultFaceSource(),
	}
	for _, locale := range context.AppendLocales(nil) {
		fs := cjkfont.FaceSourceFromLocale(locale)
		if fs != nil {
			faceSources = append(faceSources, fs)
			break
		}
	}
	if len(faceSources) == 1 {
		// Set a Japanese font as a fallback. You can use any font you like here.
		faceSources = append(faceSources, cjkfont.FaceSourceJP())
	}
	basicwidget.SetFaceSources(faceSources)

	appender.AppendChildWidgetWithBounds(&r.background, context.AppBounds())


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
	bounds := context.Bounds(r)
	appSize:= context.AppSize() // Get root size

	const margin int = 8

	// ウィジェットの配置計算
	// 利用可能な幅はRootの幅からmarginを両側分引いたもの
	availableWidth := appSize.X - margin*2

	// 各ウィジェットの高さを定義
	const (
		nowPlayingTextHeight = 30
		timeTextHeight       = 20
		settingsTextHeight   = 30
		sliderHeight         = 20
	)

	// ウィジェットの縦方向の配置を下から順に計算
	// intervalSlider
	intervalSliderY := appSize.Y - margin - sliderHeight

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
	appender.AppendChildWidgetWithBounds(
		&r.musicList, 
		image.Rect(bounds.Min.X+margin, 
			bounds.Min.Y+musicListY,
			bounds.Min.X+margin+availableWidth,
			bounds.Min.Y+musicListY+musicListHeight,
			),
	)

	// Now Playing Text
	appender.AppendChildWidgetWithBounds(
		&r.nowPlayingText,
		image.Rect(bounds.Min.X+margin,
			bounds.Min.Y+nowPlayingTextY,
			bounds.Min.X+margin+availableWidth,
			bounds.Min.Y+nowPlayingTextY+nowPlayingTextHeight,
		),
	)
	// Time Text
	appender.AppendChildWidgetWithBounds(
		&r.timeText,
		image.Rect(bounds.Min.X+margin,
			bounds.Min.Y+timeTextY,
			bounds.Min.X+margin+availableWidth,
			bounds.Min.Y+timeTextY+timeTextHeight,
		),
	)

	// Settings Text
	appender.AppendChildWidgetWithBounds(
		&r.settingsText,
		image.Rect(bounds.Min.X+margin,
			bounds.Min.Y+settingsTextY,
			bounds.Min.X+margin+availableWidth,
			bounds.Min.Y+settingsTextY+settingsTextHeight,
		),
	)

	// Loop Duration Slider
	appender.AppendChildWidgetWithBounds(
		&r.loopDurationSlider,
		image.Rect(bounds.Min.X+margin,
			bounds.Min.Y+loopDurationSliderY,
			bounds.Min.X+margin+availableWidth,
			bounds.Min.Y+loopDurationSliderY+sliderHeight,
		),
	)

	// Interval Slider
	appender.AppendChildWidgetWithBounds(
		&r.intervalSlider,
		image.Rect(bounds.Min.X+margin,
			bounds.Min.Y+intervalSliderY,
			bounds.Min.X+margin+availableWidth,
			bounds.Min.Y+intervalSliderY+sliderHeight,
		),
	)

	return nil
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

		// 選択状態の更新はここでは行わない (無限ループの原因)
		// currentIndex := r.player.GetCurrentIndex()
		// r.musicList.SetSelectedItemIndex(currentIndex)
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
	listItems := make([]basicwidget.TextListItem[string], 0, len(musicFiles))

	for _, path := range musicFiles {
		relPath := path
		if strings.HasPrefix(path, "musics/") || strings.HasPrefix(path, "musics\\") {
			relPath = path[len("musics/"):]
		}

		item := basicwidget.TextListItem[string]{
			Text: relPath, // ListItem still needs a Widget (pointer)
			Tag:  path,
		}
		listItems = append(listItems, item)
	}

	// Call method on value type
	r.musicList.SetItems(listItems)

	// 現在再生中の曲のインデックスを選択状態にする
	currentIndex := r.player.GetCurrentIndex()
	if currentIndex >= 0 && currentIndex < len(musicFiles) {
		r.musicList.SelectItemByIndex(currentIndex)
	}
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
