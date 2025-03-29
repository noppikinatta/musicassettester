package ui

import (
	"fmt"
	"image"
	"log"
	"strings"
	"time"

	"musicplayer/internal/files"
	"musicplayer/internal/player"
	"musicplayer/internal/ui/widgets"

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
	guigui.RootWidget

	player      *player.MusicPlayer
	warningText string

	// UI components
	musicList      *basicwidget.List
	nowPlayingText *basicwidget.Text
	timeText       *basicwidget.Text
	loopSlider     *widgets.Slider
	intervalSlider *widgets.Slider

	// State
	selectedIndex int
	lastUpdate    time.Time

	// GUI components
	progressBar *widgets.ProgressBar

	// Settings components
	settingsText       *basicwidget.Text
	loopDurationSlider *widgets.Slider
}

// NewRoot creates a new root widget
func NewRoot(player *player.MusicPlayer, warningText string) *Root {
	r := &Root{
		player:        player,
		warningText:   warningText,
		selectedIndex: -1,
		lastUpdate:    time.Now(),
	}
	return r
}

// Layout lays out the root widget
func (r *Root) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	// Create music list if it doesn't exist
	if r.musicList == nil {
		r.musicList = &basicwidget.List{}
		r.musicList.SetOnItemSelected(func(index int) {
			if r.player != nil {
				musicFiles := r.player.GetMusicFiles()
				if index >= 0 && index < len(musicFiles) {
					if err := r.player.SetCurrentIndex(index); err != nil {
						log.Printf("Failed to set current index: %v", err)
						return
					}

					// Reset player state and skip to selected track
					if err := r.player.SkipToNext(); err != nil {
						log.Printf("Failed to load music: %v", err)
					}
				}
			}
		})
	}

	// Create now playing text if it doesn't exist
	if r.nowPlayingText == nil {
		r.nowPlayingText = &basicwidget.Text{}
		r.nowPlayingText.SetText("")
		r.nowPlayingText.SetBold(true)
		r.nowPlayingText.SetScale(1.5)
	}

	// Create time text if it doesn't exist
	if r.timeText == nil {
		r.timeText = &basicwidget.Text{}
		r.timeText.SetText("")
	}

	// Create settings text if it doesn't exist
	if r.settingsText == nil {
		r.settingsText = &basicwidget.Text{}
		r.settingsText.SetText("Settings")
		r.settingsText.SetBold(true)
	}

	// Create loop duration slider if it doesn't exist
	if r.loopDurationSlider == nil {
		r.loopDurationSlider = widgets.NewSlider()
		r.loopDurationSlider.SetMinimum(1)
		r.loopDurationSlider.SetMaximum(60)
		if r.player != nil {
			r.loopDurationSlider.SetValue(float64(r.player.GetLoopDurationMinutes()))
		}
	}

	// Create interval slider if it doesn't exist
	if r.intervalSlider == nil {
		r.intervalSlider = widgets.NewSlider()
		r.intervalSlider.SetMinimum(1)
		r.intervalSlider.SetMaximum(60)
		if r.player != nil {
			r.intervalSlider.SetValue(float64(r.player.GetIntervalSeconds()))
		}
	}

	// Layout music list
	r.musicList.SetSize(200, 300)
	pos := guigui.Position(r)
	guigui.SetPosition(r.musicList, image.Point{X: pos.X + 10, Y: pos.Y + 10})
	appender.AppendChildWidget(r.musicList)

	// Layout now playing text
	r.nowPlayingText.SetSize(400, 30)
	pos = guigui.Position(r)
	guigui.SetPosition(r.nowPlayingText, image.Point{X: pos.X + 220, Y: pos.Y + 10})
	appender.AppendChildWidget(r.nowPlayingText)

	// Layout time text
	r.timeText.SetSize(200, 20)
	pos = guigui.Position(r)
	guigui.SetPosition(r.timeText, image.Point{X: pos.X + 220, Y: pos.Y + 50})
	appender.AppendChildWidget(r.timeText)

	// Layout settings text
	r.settingsText.SetSize(200, 30)
	pos = guigui.Position(r)
	guigui.SetPosition(r.settingsText, image.Point{X: pos.X + 220, Y: pos.Y + 100})
	appender.AppendChildWidget(r.settingsText)

	// Layout loop duration slider
	r.loopDurationSlider.SetSize(200, 20)
	pos = guigui.Position(r)
	guigui.SetPosition(r.loopDurationSlider, image.Point{X: pos.X + 220, Y: pos.Y + 140})
	appender.AppendChildWidget(r.loopDurationSlider)

	// Layout interval slider
	r.intervalSlider.SetSize(200, 20)
	pos = guigui.Position(r)
	guigui.SetPosition(r.intervalSlider, image.Point{X: pos.X + 220, Y: pos.Y + 180})
	appender.AppendChildWidget(r.intervalSlider)
}

// Size returns the size of the root widget
func (r *Root) Size(context *guigui.Context) (int, int) {
	return 800, 600
}

// Update updates the root widget
func (r *Root) Update(context *guigui.Context) error {
	// Update player state
	if r.player != nil {
		if err := r.player.Update(); err != nil {
			return err
		}

		// Check if we need to update the warning text
		musicFiles := r.player.GetMusicFiles()
		if len(musicFiles) == 0 && r.warningText == "" {
			r.warningText = files.DefaultMusicDir.GetUsageInstructions()
		} else if len(musicFiles) > 0 && r.warningText != "" {
			r.warningText = ""
		}

		// Update list items if music files have changed
		if r.musicList != nil {
			currentItems := r.musicList.Items()
			if len(currentItems) != len(musicFiles) {
				r.updateMusicList(musicFiles)
			}
		}
	}

	// Space key to toggle pause
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if r.player != nil {
			r.player.TogglePause()
		}
	}

	// N key to skip to next track
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		if r.player != nil {
			if err := r.player.SkipToNext(); err != nil {
				log.Printf("Failed to skip to next track: %v", err)
			}
		}
	}

	return nil
}

// updateMusicList updates the music list widget with new files
func (r *Root) updateMusicList(musicFiles []string) {
	if r.musicList == nil {
		return
	}

	listItems := make([]widgets.ListItem, 0, len(musicFiles))

	// Get relative paths from musics directory
	for _, path := range musicFiles {
		relPath := path
		if strings.HasPrefix(path, "musics/") || strings.HasPrefix(path, "musics\\") {
			relPath = path[len("musics/"):]
		}

		// Create a text widget for each item
		text := widgets.NewText(relPath)

		// Create list item
		item := widgets.ListItem{
			Content:    text,
			Selectable: true,
			Tag:        path, // Store original path as tag
		}

		listItems = append(listItems, item)
	}

	r.musicList.SetItems(listItems)
}

// Draw draws the root widget
func (r *Root) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Get the widget position
	pos := guigui.Position(r)

	// Create options for drawing
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(pos.X), float64(pos.Y))

	// Draw background
	rect := ebiten.NewImage(800, 600)
	_, bgColor, _ := widgets.Colors()
	rect.Fill(bgColor)
	dst.DrawImage(rect, opts)

	// Draw warning text if needed
	if r.warningText != "" {
		warningText := widgets.NewText(r.warningText)
		warningText.SetSize(image.Point{X: 400, Y: 30})
		warningText.SetBold(true)
		warningTextPos := image.Point{X: pos.X + 200, Y: pos.Y + 200}
		guigui.SetPosition(warningText, warningTextPos)
		warningText.Draw(context, dst)
		return
	}

	// Update now playing text
	if r.player != nil {
		currentPath := r.player.GetCurrentPath()
		if currentPath != "" {
			relPath := currentPath
			if strings.HasPrefix(relPath, "musics/") || strings.HasPrefix(relPath, "musics\\") {
				relPath = relPath[len("musics/"):]
			}

			// Show pause status if paused
			statusText := "Now Playing: " + relPath
			if r.player.IsPaused() {
				statusText = "PAUSED: " + relPath
			}
			r.nowPlayingText.SetText(statusText)
		} else {
			r.nowPlayingText.SetText("No track playing")
		}
	} else {
		r.nowPlayingText.SetText("No track playing")
	}

	// Update time text
	if r.player != nil {
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
		}
	} else {
		r.timeText.SetText("")
	}

	// Update loop duration text
	if r.player != nil {
		r.loopDurationSlider.SetValue(float64(r.player.GetLoopDurationMinutes()))
	}

	// Update interval text
	if r.player != nil {
		r.intervalSlider.SetValue(float64(r.player.GetIntervalSeconds()))
	}
}

// CursorShape returns the cursor shape for this widget
func (r *Root) CursorShape(context *guigui.Context) (ebiten.CursorShapeType, bool) {
	return ebiten.CursorShapeDefault, true
}
