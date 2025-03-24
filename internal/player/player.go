package player

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

	"musicplayer/internal/files"
)

// Constants for the player
const (
	sampleRate     = 48000
	bytesPerSample = 4

	// Fade-out constants
	fadeOutDuration = 2 * time.Second // 2 second fadeout
)

// Player state enum
type PlayerState int

const (
	StateStopped PlayerState = iota
	StatePlaying
	StateFadingOut
	StateInterval
)

// PlayerFactory インターフェースは音声プレイヤーの生成を抽象化します
type PlayerFactory interface {
	NewPlayer(stream io.Reader) (*audio.Player, error)
}

// MusicPlayer handles music playback
type MusicPlayer struct {
	playerFactory PlayerFactory
	player        *audio.Player
	audioStream   io.ReadSeeker
	musicFiles    []string
	currentIndex  int
	currentPath   string

	// Control variables
	state            PlayerState
	counter          int
	isPaused         bool
	loopDuration     float64 // in minutes
	intervalDuration float64 // in seconds
	volume           float64 // Current volume (0.0-1.0)
}

// NewMusicPlayer creates a new music player
func NewMusicPlayer(musicDir files.MusicDirectory, playerFactory PlayerFactory) (*MusicPlayer, error) {
	// Create player
	player := &MusicPlayer{
		playerFactory:    playerFactory,
		musicFiles:       []string{},
		currentIndex:     -1,
		state:            StateStopped,
		loopDuration:     5.0,  // Default 5 minutes
		intervalDuration: 10.0, // Default 10 seconds
		volume:           1.0,  // Full volume by default
	}

	// Load music files
	musicFiles, err := musicDir.FindMusicFiles()
	if err != nil {
		log.Printf("Warning: failed to find music files: %v", err)
	}

	// Set the music files
	player.musicFiles = musicFiles

	// Start with the first track if available
	if len(musicFiles) > 0 {
		player.currentIndex = 0
		if err := player.loadCurrentMusic(); err != nil {
			return player, fmt.Errorf("failed to load first track: %v", err)
		}
	}

	return player, nil
}

// GetMusicFiles returns the list of music files
func (p *MusicPlayer) GetMusicFiles() []string {
	return p.musicFiles
}

// GetCurrentPath returns the path of the currently playing music
func (p *MusicPlayer) GetCurrentPath() string {
	return p.currentPath
}

// GetState returns the current state of the player
func (p *MusicPlayer) GetState() PlayerState {
	return p.state
}

// IsPaused returns whether the player is paused
func (p *MusicPlayer) IsPaused() bool {
	return p.isPaused
}

// GetCounter returns the current counter value
func (p *MusicPlayer) GetCounter() int {
	return p.counter
}

// GetLoopDurationMinutes returns the loop duration in minutes
func (p *MusicPlayer) GetLoopDurationMinutes() float64 {
	return p.loopDuration
}

// SetLoopDurationMinutes sets the loop duration in minutes
func (p *MusicPlayer) SetLoopDurationMinutes(minutes float64) {
	p.loopDuration = minutes
}

// GetIntervalSeconds returns the interval duration in seconds
func (p *MusicPlayer) GetIntervalSeconds() float64 {
	return p.intervalDuration
}

// SetIntervalSeconds sets the interval duration in seconds
func (p *MusicPlayer) SetIntervalSeconds(seconds float64) {
	p.intervalDuration = seconds
}

// SetCurrentIndex sets the current index of the music
func (p *MusicPlayer) SetCurrentIndex(index int) error {
	if index < 0 || index >= len(p.musicFiles) {
		return fmt.Errorf("index out of range: %d", index)
	}
	p.currentIndex = index
	return nil
}

// loadCurrentMusic loads the current music
func (p *MusicPlayer) loadCurrentMusic() error {
	// Check if there are music files
	if len(p.musicFiles) == 0 {
		return fmt.Errorf("no music files available")
	}

	// Check if the current index is valid
	if p.currentIndex < 0 || p.currentIndex >= len(p.musicFiles) {
		return fmt.Errorf("current index out of range: %d", p.currentIndex)
	}

	// If a player is already active, close it
	if p.player != nil {
		// Close old player
		if err := p.player.Close(); err != nil {
			log.Printf("Warning: failed to close player: %v", err)
		}
		p.player = nil
	}

	// Get the current music file
	currentPath := p.musicFiles[p.currentIndex]
	p.currentPath = currentPath

	// Open the file
	f, err := os.Open(currentPath)
	if err != nil {
		return fmt.Errorf("failed to open audio file %s: %v", currentPath, err)
	}

	// Decode based on file extension
	var audioStream io.ReadSeeker

	if files.IsWavFile(currentPath) {
		audioStream, err = wav.DecodeWithSampleRate(sampleRate, f)
	} else if files.IsOggFile(currentPath) {
		audioStream, err = vorbis.DecodeWithSampleRate(sampleRate, f)
	} else if files.IsMp3File(currentPath) {
		audioStream, err = mp3.DecodeWithSampleRate(sampleRate, f)
	} else {
		f.Close()
		return fmt.Errorf("unsupported audio format: %s", currentPath)
	}

	if err != nil {
		f.Close()
		return fmt.Errorf("failed to decode audio: %v", err)
	}

	// Create an infinite loop stream
	p.audioStream = audioStream
	loopStream := audio.NewInfiniteLoop(audioStream, audioStream.(interface{ Length() int64 }).Length())

	// Create player using the factory
	player, err := p.playerFactory.NewPlayer(loopStream)
	if err != nil {
		return fmt.Errorf("failed to create audio player: %v", err)
	}
	p.player = player
	p.player.SetVolume(p.volume)

	// Reset counter and state
	p.counter = 0
	p.state = StatePlaying
	p.isPaused = false

	// Start playing
	p.player.Play()

	return nil
}

// TogglePause toggles pause state
func (p *MusicPlayer) TogglePause() {
	if p.player == nil {
		return
	}

	if p.isPaused {
		p.player.Play()
		p.isPaused = false
	} else {
		p.player.Pause()
		p.isPaused = true
	}
}

// Update updates the player state
func (p *MusicPlayer) Update() error {
	// Skip if no player or paused
	if p.player == nil || p.isPaused {
		return nil
	}

	// Increment counter (60 times per second in Ebiten)
	p.counter++

	// Handle different states
	switch p.state {
	case StatePlaying:
		// Check if we reached the time limit (convert minutes to frames)
		loopDurationFrames := int(p.loopDuration * 60 * 60) // minutes * 60 seconds * 60 frames
		if p.counter >= loopDurationFrames {
			// Start fade out
			p.state = StateFadingOut
			p.counter = 0
		}

	case StateFadingOut:
		// Calculate fade-out duration in frames
		fadeOutFrames := int(fadeOutDuration.Seconds() * 60) // 2 seconds * 60 frames
		if p.counter >= fadeOutFrames {
			// Time to switch to interval
			p.state = StateInterval
			p.counter = 0

			// Stop the player
			if p.player != nil {
				p.player.Pause()
			}
		} else {
			// Calculate fade-out volume
			fadeRatio := 1.0 - float64(p.counter)/float64(fadeOutFrames)
			p.volume = fadeRatio
			if p.player != nil {
				p.player.SetVolume(fadeRatio)
			}
		}

	case StateInterval:
		// Calculate interval duration in frames
		intervalFrames := int(p.intervalDuration * 60) // seconds * 60 frames
		if p.counter >= intervalFrames {
			// Time for next song
			p.volume = 1.0 // Reset volume to full
			err := p.SkipToNext()
			if err != nil {
				return fmt.Errorf("failed to skip to next track: %v", err)
			}
		}
	}

	return nil
}

// SkipToNext skips to the next track
func (p *MusicPlayer) SkipToNext() error {
	// Determine next index
	nextIndex := p.currentIndex + 1
	if nextIndex >= len(p.musicFiles) {
		// Loop back to the first track
		nextIndex = 0
	}

	// Set current index
	p.currentIndex = nextIndex

	// Reset volume to full
	p.volume = 1.0

	// Load and play the selected music
	return p.loadCurrentMusic()
}
