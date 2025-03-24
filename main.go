package main

import (
	"bytes"
	"fmt"
	"image/color"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/hajimehoshi/guigui"
	"github.com/hajimehoshi/guigui/basicwidget"
)

const (
	screenWidth  = 800
	screenHeight = 400
	sampleRate   = 44100

	// Timer settings
	fadeOutDuration = 60 * 5 // 5 seconds (60fps * 5sec)
)

type PlayerState int

const (
	StatePlaying PlayerState = iota
	StateFadingOut
	StateInterval
)

// MusicPlayer represents the state of the music player
type MusicPlayer struct {
	audioContext *audio.Context
	audioPlayer  *audio.Player
	musicFiles   []string
	currentPath  string
	currentIndex int // Index of the current track
	state        PlayerState
	counter      int // Counter for the current state
	volume       float64
	paused       bool // Whether playback is paused

	// Settings
	loopDurationMinutes float64 // Loop duration in minutes (1-10)
	intervalSeconds     float64 // Interval duration in seconds (1-60)
}

// Root is our guigui RootWidget
type Root struct {
	guigui.RootWidget

	player         *MusicPlayer
	warningMessage string

	// GUI components
	musicList     *basicwidget.List
	selectedIndex int

	// Now playing info components
	nowPlayingText *basicwidget.Text
	timeText       *basicwidget.Text
	progressBar    *ProgressBar // Custom progress bar

	// Settings components
	loopDurationSlider *Slider // Custom slider
	intervalSlider     *Slider // Custom slider
	settingsText       *basicwidget.Text
}

// Custom progress bar implementation
type ProgressBar struct {
	guigui.DefaultWidget

	value  float64
	width  int
	height int
}

func (p *ProgressBar) SetValue(value float64) {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	p.value = value
}

func (p *ProgressBar) Value() float64 {
	return p.value
}

func (p *ProgressBar) SetSize(width, height int) {
	p.width = width
	p.height = height
}

func (p *ProgressBar) Size(context *guigui.Context) (int, int) {
	return p.width, p.height
}

func (p *ProgressBar) Draw(context *guigui.Context, dst *ebiten.Image) {
	// Draw background
	x, y := guigui.Position(p).X, guigui.Position(p).Y
	w, h := p.Size(context)

	// Background (gray)
	vector.DrawFilledRect(dst, float32(x), float32(y), float32(w), float32(h), color.RGBA{100, 100, 100, 255}, false)

	// Progress (green)
	progressWidth := float32(float64(w) * p.value)
	if progressWidth > 0 {
		vector.DrawFilledRect(dst, float32(x), float32(y), progressWidth, float32(h), color.RGBA{0, 200, 100, 255}, false)
	}

	// Border
	vector.StrokeRect(dst, float32(x), float32(y), float32(w), float32(h), 1, color.RGBA{150, 150, 150, 255}, false)
}

// Custom slider implementation
type Slider struct {
	guigui.DefaultWidget

	value    float64
	min      float64
	max      float64
	width    int
	height   int
	onChange func(float64)

	dragging       bool
	dragStartX     int
	dragStartValue float64
}

func (s *Slider) SetValue(value float64) {
	if value < s.min {
		value = s.min
	}
	if value > s.max {
		value = s.max
	}
	s.value = value
}

func (s *Slider) Value() float64 {
	return s.value
}

func (s *Slider) SetMinimum(min float64) {
	s.min = min
	if s.value < min {
		s.value = min
	}
}

func (s *Slider) SetMaximum(max float64) {
	s.max = max
	if s.value > max {
		s.value = max
	}
}

func (s *Slider) SetOnChange(f func(float64)) {
	s.onChange = f
}

func (s *Slider) SetSize(width, height int) {
	s.width = width
	s.height = height
}

func (s *Slider) Size(context *guigui.Context) (int, int) {
	return s.width, s.height
}

func (s *Slider) HandleInput(context *guigui.Context) guigui.HandleInputResult {
	x, y := guigui.Position(s).X, guigui.Position(s).Y
	w, h := s.Size(context)

	mx, my := ebiten.CursorPosition()

	// Check if mouse is over the slider
	if mx >= x && mx < x+w && my >= y && my < y+h {
		// Start dragging on mouse press
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			s.dragging = true
			s.dragStartX = mx
			s.dragStartValue = s.value
			return guigui.HandleInputByWidget(s)
		}
	}

	// Handle dragging
	if s.dragging {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			relativeX := float64(mx-x) / float64(w)
			newValue := s.min + relativeX*(s.max-s.min)

			// Clamp value
			if newValue < s.min {
				newValue = s.min
			}
			if newValue > s.max {
				newValue = s.max
			}

			// Only call onChange if value actually changed
			if newValue != s.value {
				s.value = newValue
				if s.onChange != nil {
					s.onChange(s.value)
				}
			}

			return guigui.HandleInputByWidget(s)
		} else {
			// Stop dragging when button is released
			s.dragging = false
		}
	}

	return guigui.HandleInputResult{}
}

func (s *Slider) Draw(context *guigui.Context, dst *ebiten.Image) {
	x, y := guigui.Position(s).X, guigui.Position(s).Y
	w, h := s.Size(context)

	// Draw track
	trackY := y + h/2 - 2
	trackHeight := 4
	vector.DrawFilledRect(dst, float32(x), float32(trackY), float32(w), float32(trackHeight), color.RGBA{100, 100, 100, 255}, false)

	// Calculate thumb position
	valueRatio := (s.value - s.min) / (s.max - s.min)
	thumbX := float32(x + int(float64(w)*valueRatio))
	thumbSize := float32(12)
	thumbY := float32(y + h/2 - int(thumbSize)/2)

	// Draw thumb
	vector.DrawFilledCircle(dst, thumbX, thumbY+thumbSize/2, thumbSize/2, color.RGBA{200, 200, 200, 255}, false)
	vector.StrokeCircle(dst, thumbX, thumbY+thumbSize/2, thumbSize/2, 1, color.RGBA{150, 150, 150, 255}, false)
}

// Searches for music files (.wav, .ogg, .mp3) in the specified directory
func findMusicFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext == ".wav" || ext == ".ogg" || ext == ".mp3" {
				files = append(files, path)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}

// Loads the next music track in sequence
func (p *MusicPlayer) loadNextMusic() error {
	if p.audioPlayer != nil {
		if err := p.audioPlayer.Close(); err != nil {
			return fmt.Errorf("failed to close player: %v", err)
		}
		p.audioPlayer = nil
	}

	if len(p.musicFiles) == 0 {
		return fmt.Errorf("no music files available")
	}

	// Select tracks in sequence (loop back to the beginning after the last track)
	p.currentPath = p.musicFiles[p.currentIndex]

	// Prepare the index for the next track (return to 0 if it's the last track)
	p.currentIndex = (p.currentIndex + 1) % len(p.musicFiles)

	// Load the music file
	fileData, err := os.ReadFile(p.currentPath)
	if err != nil {
		return fmt.Errorf("failed to load file: %v", err)
	}

	// Create decoder based on file extension
	ext := strings.ToLower(filepath.Ext(p.currentPath))

	var stream io.ReadSeeker

	switch ext {
	case ".wav":
		stream, err = wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(fileData))
	case ".ogg":
		stream, err = vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(fileData))
	case ".mp3":
		stream, err = mp3.DecodeWithSampleRate(sampleRate, bytes.NewReader(fileData))
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}

	if err != nil {
		return fmt.Errorf("failed to decode %s: %v", ext, err)
	}

	// Create an infinite loop stream (without intro)
	var length int64
	switch ext {
	case ".wav":
		length = stream.(*wav.Stream).Length()
	case ".ogg":
		length = stream.(*vorbis.Stream).Length()
	case ".mp3":
		length = stream.(*mp3.Stream).Length()
	}
	loopStream := audio.NewInfiniteLoop(stream, length)

	// Create player
	p.audioPlayer, err = p.audioContext.NewPlayer(loopStream)
	if err != nil {
		return fmt.Errorf("failed to create player: %v", err)
	}

	// Set volume and start playing
	p.audioPlayer.SetVolume(p.volume)
	p.audioPlayer.Play()

	return nil
}

// NewMusicPlayer creates a new music player
func NewMusicPlayer() (*MusicPlayer, error) {
	audioContext := audio.NewContext(sampleRate)

	// Search for music files in the musics directory
	musicFiles, err := findMusicFiles("musics")
	if err != nil {
		return nil, fmt.Errorf("failed to search for music files: %v", err)
	}

	if len(musicFiles) == 0 {
		return nil, fmt.Errorf("no music files found in the musics directory")
	}

	player := &MusicPlayer{
		audioContext:        audioContext,
		musicFiles:          musicFiles,
		currentIndex:        0,             // Start from the first track
		state:               StateInterval, // Start in interval state
		counter:             0,
		volume:              1.0,  // Start with maximum volume
		loopDurationMinutes: 5.0,  // Default: 5 minutes
		intervalSeconds:     10.0, // Default: 10 seconds
	}

	return player, nil
}

// Performs update processing
func (p *MusicPlayer) update() error {
	// Skip counter increment and state changes if paused
	if p.paused {
		return nil
	}

	// Calculate durations based on settings
	playDurationFrames := int(p.loopDurationMinutes * 60 * 60) // minutes * 60 seconds * 60 frames
	intervalDurationFrames := int(p.intervalSeconds * 60)      // seconds * 60 frames

	switch p.state {
	case StatePlaying:
		// Start fading out after configured play time
		if p.counter >= playDurationFrames {
			p.state = StateFadingOut
			p.counter = 0
		}

	case StateFadingOut:
		// Fade out the volume over 5 seconds
		if p.counter < fadeOutDuration {
			newVolume := 1.0 - float64(p.counter)/float64(fadeOutDuration)
			p.volume = newVolume
			if p.audioPlayer != nil {
				p.audioPlayer.SetVolume(p.volume)
			}
		} else {
			// Fade out complete
			if p.audioPlayer != nil {
				p.audioPlayer.Close()
				p.audioPlayer = nil
			}
			p.state = StateInterval
			p.counter = 0
		}

	case StateInterval:
		// Interval based on settings
		if p.counter >= intervalDurationFrames {
			// Next track
			p.volume = 1.0
			p.state = StatePlaying
			p.counter = 0
			if err := p.loadNextMusic(); err != nil {
				return err
			}
		} else if p.counter == 0 {
			// Called only once at the start of the interval
			// When first launched, playback starts from here
			if p.audioPlayer == nil {
				if err := p.loadNextMusic(); err != nil {
					return err
				}
				p.state = StatePlaying
			}
		}
	}

	p.counter++

	return nil
}

// TogglePause toggles the pause state
func (p *MusicPlayer) togglePause() {
	p.paused = !p.paused

	if p.audioPlayer != nil {
		if p.paused {
			p.audioPlayer.Pause()
		} else {
			p.audioPlayer.Play()
		}
	}
}

// Layout implements the guigui.Widget interface
func (r *Root) Layout(context *guigui.Context, appender *guigui.ChildWidgetAppender) {
	// Calculate main layout areas
	listWidth := 250
	listHeight := screenHeight - 40 // Leave some space at bottom

	infoAreaX := listWidth + 40 // After list + margin
	infoAreaY := 20
	infoAreaWidth := screenWidth - infoAreaX - 20

	// Create music list if it doesn't exist
	if r.musicList == nil {
		r.musicList = &basicwidget.List{}

		// Set music files to the list box if available
		if r.player != nil && len(r.player.musicFiles) > 0 {
			listItems := make([]basicwidget.ListItem, 0, len(r.player.musicFiles))

			// Get relative paths from musics directory
			for _, path := range r.player.musicFiles {
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

		// Set selection callback
		r.musicList.SetOnItemSelected(func(index int) {
			r.selectedIndex = index
			if r.player != nil && index >= 0 && index < len(r.player.musicFiles) {
				// Set the new index
				r.player.currentIndex = index

				// Reset player state
				r.player.volume = 1.0
				r.player.state = StatePlaying
				r.player.counter = 0

				// Load and play the selected music
				err := r.player.loadNextMusic()
				if err != nil {
					log.Printf("Failed to load music: %v", err)
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
	if r.player != nil && r.player.currentPath != "" {
		relPath := r.player.currentPath
		if strings.HasPrefix(relPath, "musics/") || strings.HasPrefix(relPath, "musics\\") {
			relPath = relPath[len("musics/"):]
		}

		// Show pause status if paused
		statusText := "Now Playing: " + relPath
		if r.player.paused {
			statusText = "PAUSED: " + relPath
		}
		r.nowPlayingText.SetText(statusText)
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
	if r.player != nil && r.player.state == StatePlaying {
		currentTimeSec := r.player.counter / 60
		totalTimeSec := int(r.player.loopDurationMinutes * 60)
		r.timeText.SetText(fmt.Sprintf("%d:%02d / %d:%02d",
			currentTimeSec/60, currentTimeSec%60,
			totalTimeSec/60, totalTimeSec%60))
	} else if r.player != nil && r.player.state == StateFadingOut {
		r.timeText.SetText("Fading out...")
	} else if r.player != nil && r.player.state == StateInterval {
		intervalSec := (int(r.player.intervalSeconds)*60 - r.player.counter) / 60
		r.timeText.SetText(fmt.Sprintf("Next track in: %d seconds", intervalSec))
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
		r.progressBar = &ProgressBar{}
	}

	// Set progress bar value - updated for variable duration
	if r.player != nil && r.player.state == StatePlaying {
		progress := float64(r.player.counter) / float64(int(r.player.loopDurationMinutes*60)*60)
		r.progressBar.SetValue(progress)
	} else if r.player != nil && r.player.state == StateFadingOut {
		r.progressBar.SetValue(1.0)
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
		r.loopDurationSlider = &Slider{}
		r.loopDurationSlider.SetMinimum(1)
		r.loopDurationSlider.SetMaximum(10)

		if r.player != nil {
			r.loopDurationSlider.SetValue(r.player.loopDurationMinutes)
		} else {
			r.loopDurationSlider.SetValue(5)
		}

		r.loopDurationSlider.SetOnChange(func(v float64) {
			if r.player != nil {
				r.player.loopDurationMinutes = v
			}
		})
	}

	// Create label for the loop duration slider
	loopDurationText := &basicwidget.Text{}
	if r.player != nil {
		loopDurationText.SetText(fmt.Sprintf("Loop Duration: %.1f minutes", r.player.loopDurationMinutes))
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
		r.intervalSlider = &Slider{}
		r.intervalSlider.SetMinimum(1)
		r.intervalSlider.SetMaximum(60)

		if r.player != nil {
			r.intervalSlider.SetValue(r.player.intervalSeconds)
		} else {
			r.intervalSlider.SetValue(10)
		}

		r.intervalSlider.SetOnChange(func(v float64) {
			if r.player != nil {
				r.player.intervalSeconds = v
			}
		})
	}

	// Create label for the interval slider
	intervalText := &basicwidget.Text{}
	if r.player != nil {
		intervalText.SetText(fmt.Sprintf("Interval Between Tracks: %.1f seconds", r.player.intervalSeconds))
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
			r.player.togglePause()
		}
	}

	// N key to skip to next track
	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		if r.player != nil {
			r.player.volume = 1.0
			r.player.state = StatePlaying
			r.player.counter = 0
			if err := r.player.loadNextMusic(); err != nil {
				return err
			}
		}
	}

	// Update player if it exists
	if r.player != nil {
		return r.player.update()
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
		if r.player.paused {
			currentStatus = "PAUSED - Press space to resume\n\n"
		}

		// Change display based on state
		switch r.player.state {
		case StatePlaying, StateFadingOut:
			if r.player.currentPath != "" {
				var remainingSecs int
				if r.player.state == StatePlaying {
					playDurationSec := int(r.player.loopDurationMinutes * 60)
					remainingSecs = playDurationSec - (r.player.counter / 60)
				} else {
					remainingSecs = 0
				}

				currentStatus += fmt.Sprintf("Now Playing: %s\nRemaining: %d seconds", r.player.currentPath, remainingSecs)
			}
		case StateInterval:
			// During interval
			intervalSec := int(r.player.intervalSeconds) - (r.player.counter / 60)
			currentStatus += fmt.Sprintf("Interval...\nNext Track in: %d seconds", intervalSec)
		}

		if currentStatus != "" {
			ebitenutil.DebugPrintAt(dst, currentStatus, 20, 20)
		}
	}
}

// Returns instruction message about required files
func GetHowToUseMessage() string {
	message := "Warning: Music files are needed. Please place WAV, OGG, or MP3 files in the musics directory and run again.\n\n"
	message += "Example:\n"
	message += "musics/\n"
	message += "├── song1.wav\n"
	message += "├── song2.mp3\n"
	message += "└── album/\n"
	message += "    ├── song3.ogg\n"
	message += "    └── song4.wav\n"
	return message
}

func main() {
	// Create a root widget
	root := &Root{}

	// Check if the musics directory exists, and create it if not
	if _, err := os.Stat("musics"); os.IsNotExist(err) {
		if err := os.Mkdir("musics", 0755); err != nil {
			log.Fatal("Failed to create musics directory:", err)
		}
		// Set warning message
		root.warningMessage = GetHowToUseMessage()
	} else {
		// Check if there are any music files
		musicFiles, err := findMusicFiles("musics")
		if err != nil {
			log.Fatal("Failed to search for music files:", err)
		}

		if len(musicFiles) == 0 {
			// Set warning message for no music files
			root.warningMessage = GetHowToUseMessage()
		} else {
			// Initialize music player
			player, err := NewMusicPlayer()
			if err != nil {
				log.Fatal(err)
			}
			root.player = player
		}
	}

	// Run the application with guigui
	op := &guigui.RunOptions{
		Title:           "Music Asset Tester",
		WindowMinWidth:  screenWidth,
		WindowMinHeight: screenHeight,
	}

	if err := guigui.Run(root, op); err != nil {
		log.Fatal(err)
	}
}
