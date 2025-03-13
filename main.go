package main

import (
	"bytes"
	"fmt"
	"image/color"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	screenWidth  = 640
	screenHeight = 400
	sampleRate   = 44100 // Changed to 44100 based on specifications

	// Timer settings
	playDuration     = 60 * 60 * 5 // 5 minutes (60fps * 60sec * 5min)
	fadeOutDuration  = 60 * 5      // 5 seconds (60fps * 5sec)
	intervalDuration = 60 * 10     // 10 seconds (60fps * 10sec)
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
}

// Game represents the overall state of the game
type Game struct {
	player         *MusicPlayer
	warningMessage string // Added field to store warning messages
}

// Searches for .wav files in the specified directory
func findWavFiles(dir string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".wav") {
			files = append(files, path)
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

	// Load the WAV file
	fileData, err := os.ReadFile(p.currentPath)
	if err != nil {
		return fmt.Errorf("failed to load file: %v", err)
	}

	// Create WAV decoder (read from byte array)
	d, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("failed to decode WAV: %v", err)
	}

	// Create an infinite loop stream (without intro)
	loopStream := audio.NewInfiniteLoop(d, d.Length())

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

	// Search for .wav files in the musics directory
	musicFiles, err := findWavFiles("musics")
	if err != nil {
		return nil, fmt.Errorf("failed to search for music files: %v", err)
	}

	if len(musicFiles) == 0 {
		return nil, fmt.Errorf("no WAV files found in the musics directory")
	}

	player := &MusicPlayer{
		audioContext: audioContext,
		musicFiles:   musicFiles,
		currentIndex: 0,             // Start from the first track
		state:        StateInterval, // Start in interval state
		counter:      0,
		volume:       1.0, // Start with maximum volume
	}

	return player, nil
}

// Performs update processing
func (p *MusicPlayer) update() error {
	// Skip counter increment and state changes if paused
	if p.paused {
		return nil
	}

	switch p.state {
	case StatePlaying:
		// Start fading out after 10 minutes of play time
		if p.counter >= playDuration {
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
		// 10 second interval
		if p.counter >= intervalDuration {
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

// Performs drawing processing
func (p *MusicPlayer) draw(screen *ebiten.Image) {
	// Fill the screen with black
	screen.Fill(color.RGBA{40, 40, 40, 255})

	currentStatus := ""

	// Show pause status if paused
	if p.paused {
		currentStatus = "PAUSED - Click to resume\n\n"
	}

	// Change display based on state
	switch p.state {
	case StatePlaying, StateFadingOut:
		if p.currentPath != "" {
			var remainingSecs int
			if p.state == StatePlaying {
				remainingSecs = (playDuration - p.counter) / 60
			} else {
				remainingSecs = 0
			}

			currentStatus += fmt.Sprintf("Now Playing: %s\nRemaining: %d seconds", p.currentPath, remainingSecs)
		}
	case StateInterval:
		// During interval
		remainingSecs := (intervalDuration - p.counter) / 60
		currentStatus += fmt.Sprintf("Interval...\nNext Track in: %d seconds", remainingSecs)
	}

	if currentStatus != "" {
		ebitenutil.DebugPrintAt(screen, currentStatus, 20, 20)
	}
}

// NewGame creates a new game
func NewGame() (*Game, error) {
	player, err := NewMusicPlayer()
	if err != nil {
		return nil, err
	}

	return &Game{
		player: player,
	}, nil
}

// Returns instruction message about required files
func GetHowToUseMessage() string {
	message := "Warning: WAV files are needed. Please place WAV files in the musics directory and run again.\n\n"
	message += "Example:\n"
	message += "musics/\n"
	message += "+-- song1.wav\n"
	message += "+-- song2.wav\n"
	message += "+-- album/\n"
	message += "    +-- song3.wav\n"
	message += "    +-- song4.wav\n"
	return message
}

func (g *Game) Update() error {
	// Check for mouse click to toggle pause
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Only toggle pause if we have a player (no warning message)
		if g.warningMessage == "" {
			g.player.togglePause()
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyN) {
		// Only proceed if we have a player (no warning message)
		if g.warningMessage == "" {
			g.player.volume = 1.0
			g.player.state = StatePlaying
			g.player.counter = 0
			if err := g.player.loadNextMusic(); err != nil {
				return err
			}
		}
	}

	// Only update player if it exists
	if g.player != nil {
		return g.player.update()
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Fill the screen with dark gray
	screen.Fill(color.RGBA{40, 40, 40, 255})

	// If there's a warning message, display it
	if g.warningMessage != "" {
		ebitenutil.DebugPrintAt(screen, g.warningMessage, 20, 20)
		return
	}

	// Otherwise draw the player UI
	if g.player != nil {
		g.player.draw(screen)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("Music Asset Tester")

	// Create a game struct
	game := &Game{}

	// Check if the musics directory exists, and create it if not
	if _, err := os.Stat("musics"); os.IsNotExist(err) {
		if err := os.Mkdir("musics", 0755); err != nil {
			log.Fatal("Failed to create musics directory:", err)
		}
		// Set warning message instead of printing
		game.warningMessage = GetHowToUseMessage()
	} else {
		// Check if there are any WAV files
		musicFiles, err := findWavFiles("musics")
		if err != nil {
			log.Fatal("Failed to search for music files:", err)
		}

		if len(musicFiles) == 0 {
			// Set warning message for no WAV files
			game.warningMessage = GetHowToUseMessage()
		} else {
			// Initialize music player
			player, err := NewMusicPlayer()
			if err != nil {
				log.Fatal(err)
			}
			game.player = player
		}
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
