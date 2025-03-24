package ui

import (
	"fmt"
	"log"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"

	"musicplayer/internal/player"
	"musicplayer/internal/ui/widgets"
)

const (
	ScreenWidth  = 800
	ScreenHeight = 400
)

// Root is our guigui RootWidget
type Root struct {
	guigui.RootWidget

	player         *player.MusicPlayer
	warningMessage string

	// GUI components
	musicList     *basicwidget.List
	selectedIndex int

	// Now playing info components
	nowPlayingText *basicwidget.Text
	timeText       *basicwidget.Text
	progressBar    *widgets.ProgressBar // Custom progress bar

	// Settings components
	loopDurationSlider *widgets.Slider // Custom slider
	intervalSlider     *widgets.Slider // Custom slider
	settingsText       *basicwidget.Text
}

// NewRoot creates a new root widget
func NewRoot(musicPlayer *player.MusicPlayer, warningMessage string) *Root {
	return &Root{
		player:         musicPlayer,
		warningMessage: warningMessage,
	}
}

// Layout implements the guigui.Widget interface
func (r *Root) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	// Calculate main layout areas
	listWidth := 250
	listHeight := ScreenHeight - 40 // Leave some space at bottom

	infoAreaX := listWidth + 40 // After list + margin
	infoAreaY := 20
	infoAreaWidth := ScreenWidth - infoAreaX - 20

	// Create music list if it doesn't exist
	if r.musicList == nil {
		r.musicList = &basicwidget.List{}

		// Set music files to the list box if available
		if r.player != nil {
			musicFiles := r.player.GetMusicFiles()
			if len(musicFiles) > 0 {
				listItems := make([]basicwidget.ListItem, 0, len(musicFiles))

				// Get relative paths from musics directory
				for _, path := range musicFiles {
					relPath := path
					if strings.HasPrefix(path, "musics/") || strings.HasPrefix(path, "musics\\") {
						relPath = path[len("musics/"):]
					}

					// Create a text widget for each item
					text := &basicwidget.Text{}
					text.SetText(relPath)

					// Create list item
					item := basicwidget.ListItem{
						Content:    text,
						Selectable: true,
						Tag:        path, // Store original path as tag
					}

					listItems = append(listItems, item)
				}

				r.musicList.SetItems(listItems)
			}
		}

		// Set selection callback
		r.musicList.SetOnItemSelected(func(index int) {
			r.selectedIndex = index
			if r.player != nil {
				musicFiles := r.player.GetMusicFiles()
				if index >= 0 && index < len(musicFiles) {
					// Set the new index
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

	// Set the size of the list box
	r.musicList.SetSize(listWidth, listHeight)

	// Set the position of the list box (top left corner with margin)
	listPos := guigui.Position(r)
	listPos.X += 20
	listPos.Y += 20
	guigui.SetPosition(r.musicList, listPos)

	// Append the list box to the UI
	appender.AppendChildWidget(r.musicList)

	// Create now playing text if it doesn't exist
	if r.nowPlayingText == nil {
		r.nowPlayingText = &basicwidget.Text{}
		r.nowPlayingText.SetBold(true)
		r.nowPlayingText.SetScale(1.5)
	}

	// Set current playing track text
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

	// Set position and size of the now playing text
	r.nowPlayingText.SetSize(infoAreaWidth, 30)
	nowPlayingPos := guigui.Position(r)
	nowPlayingPos.X += infoAreaX
	nowPlayingPos.Y += infoAreaY
	guigui.SetPosition(r.nowPlayingText, nowPlayingPos)
	appender.AppendChildWidget(r.nowPlayingText)

	// Create time text if it doesn't exist
	if r.timeText == nil {
		r.timeText = &basicwidget.Text{}
	}

	// Set time text content - updated for variable duration
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

	// Set position and size of the time text
	r.timeText.SetSize(infoAreaWidth, 30)
	timeTextPos := guigui.Position(r)
	timeTextPos.X += infoAreaX
	timeTextPos.Y += infoAreaY + 40
	guigui.SetPosition(r.timeText, timeTextPos)
	appender.AppendChildWidget(r.timeText)

	// Create progress bar if it doesn't exist
	if r.progressBar == nil {
		r.progressBar = widgets.NewProgressBar()
	}

	// Set progress bar value - updated for variable duration
	if r.player != nil {
		switch r.player.GetState() {
		case player.StatePlaying:
			progress := float64(r.player.GetCounter()) / float64(int(r.player.GetLoopDurationMinutes()*60)*60)
			r.progressBar.SetValue(progress)
		case player.StateFadingOut:
			r.progressBar.SetValue(1.0)
		default:
			r.progressBar.SetValue(0)
		}
	} else {
		r.progressBar.SetValue(0)
	}

	// Set position and size of the progress bar
	r.progressBar.SetSize(infoAreaWidth, 20)
	progressPos := guigui.Position(r)
	progressPos.X += infoAreaX
	progressPos.Y += infoAreaY + 80
	guigui.SetPosition(r.progressBar, progressPos)
	appender.AppendChildWidget(r.progressBar)

	// Settings section
	// Create settings text if it doesn't exist
	if r.settingsText == nil {
		r.settingsText = &basicwidget.Text{}
		r.settingsText.SetBold(true)
	}
	r.settingsText.SetText("Settings")

	// Set position and size
	r.settingsText.SetSize(infoAreaWidth, 30)
	settingsTextPos := guigui.Position(r)
	settingsTextPos.X += infoAreaX
	settingsTextPos.Y += infoAreaY + 120
	guigui.SetPosition(r.settingsText, settingsTextPos)
	appender.AppendChildWidget(r.settingsText)

	// Loop duration slider
	if r.loopDurationSlider == nil {
		r.loopDurationSlider = widgets.NewSlider()
		r.loopDurationSlider.SetMinimum(1)
		r.loopDurationSlider.SetMaximum(10)

		if r.player != nil {
			r.loopDurationSlider.SetValue(r.player.GetLoopDurationMinutes())
		} else {
			r.loopDurationSlider.SetValue(5)
		}

		r.loopDurationSlider.SetOnChange(func(v float64) {
			if r.player != nil {
				r.player.SetLoopDurationMinutes(v)
			}
		})
	}

	// Create label for the loop duration slider
	loopDurationText := &basicwidget.Text{}
	if r.player != nil {
		loopDurationText.SetText(fmt.Sprintf("Loop Duration: %.1f minutes", r.player.GetLoopDurationMinutes()))
	} else {
		loopDurationText.SetText("Loop Duration: 5 minutes")
	}

	// Position and size the loop duration label
	loopDurationText.SetSize(infoAreaWidth/2, 20)
	loopDurationTextPos := guigui.Position(r)
	loopDurationTextPos.X += infoAreaX
	loopDurationTextPos.Y += infoAreaY + 150
	guigui.SetPosition(loopDurationText, loopDurationTextPos)
	appender.AppendChildWidget(loopDurationText)

	// Position and size the loop duration slider
	r.loopDurationSlider.SetSize(infoAreaWidth, 20)
	loopDurationSliderPos := guigui.Position(r)
	loopDurationSliderPos.X += infoAreaX
	loopDurationSliderPos.Y += infoAreaY + 170
	guigui.SetPosition(r.loopDurationSlider, loopDurationSliderPos)
	appender.AppendChildWidget(r.loopDurationSlider)

	// Interval slider
	if r.intervalSlider == nil {
		r.intervalSlider = widgets.NewSlider()
		r.intervalSlider.SetMinimum(1)
		r.intervalSlider.SetMaximum(60)

		if r.player != nil {
			r.intervalSlider.SetValue(r.player.GetIntervalSeconds())
		} else {
			r.intervalSlider.SetValue(10)
		}

		r.intervalSlider.SetOnChange(func(v float64) {
			if r.player != nil {
				r.player.SetIntervalSeconds(v)
			}
		})
	}

	// Create label for the interval slider
	intervalText := &basicwidget.Text{}
	if r.player != nil {
		intervalText.SetText(fmt.Sprintf("Interval Between Tracks: %.1f seconds", r.player.GetIntervalSeconds()))
	} else {
		intervalText.SetText("Interval Between Tracks: 10 seconds")
	}

	// Position and size the interval label
	intervalText.SetSize(infoAreaWidth/2, 20)
	intervalTextPos := guigui.Position(r)
	intervalTextPos.X += infoAreaX
	intervalTextPos.Y += infoAreaY + 210
	guigui.SetPosition(intervalText, intervalTextPos)
	appender.AppendChildWidget(intervalText)

	// Position and size the interval slider
	r.intervalSlider.SetSize(infoAreaWidth, 20)
	intervalSliderPos := guigui.Position(r)
	intervalSliderPos.X += infoAreaX
	intervalSliderPos.Y += infoAreaY + 230
	guigui.SetPosition(r.intervalSlider, intervalSliderPos)
	appender.AppendChildWidget(r.intervalSlider)
}

// Update implements the guigui.Widget interface
func (r *Root) Update(context *guigui.Context) error {
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
				return err
			}
		}
	}

	// Update player if it exists
	if r.player != nil {
		return r.player.Update()
	}

	return nil
}

// Draw implements the guigui.Widget interface
func (r *Root) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Fill the background
	basicwidget.FillBackground(dst, context)

	// If there's a warning message, display it
	if r.warningMessage != "" {
		ebitenutil.DebugPrintAt(dst, r.warningMessage, 20, 20)
		return
	}

	// Display player information
	if r.player != nil {
		currentStatus := ""

		// Show pause status if paused
		if r.player.IsPaused() {
			currentStatus = "PAUSED - Press space to resume\n\n"
		}

		// Change display based on state
		switch r.player.GetState() {
		case player.StatePlaying, player.StateFadingOut:
			currentPath := r.player.GetCurrentPath()
			if currentPath != "" {
				var remainingSecs int
				if r.player.GetState() == player.StatePlaying {
					playDurationSec := int(r.player.GetLoopDurationMinutes() * 60)
					remainingSecs = playDurationSec - (r.player.GetCounter() / 60)
				} else {
					remainingSecs = 0
				}

				currentStatus += fmt.Sprintf("Now Playing: %s\nRemaining: %d seconds", currentPath, remainingSecs)
			}
		case player.StateInterval:
			// During interval
			intervalSec := int(r.player.GetIntervalSeconds()) - (r.player.GetCounter() / 60)
			currentStatus += fmt.Sprintf("Interval...\nNext Track in: %d seconds", intervalSec)
		}

		if currentStatus != "" {
			ebitenutil.DebugPrintAt(dst, currentStatus, 20, 20)
		}
	}
}
