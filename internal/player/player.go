package player

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"

	"musicplayer/internal/files"
)

// --- MusicSelector ---

// MusicSelector manages the list of music files and the current selection.
type MusicSelector struct {
	musicFiles   []string
	currentIndex int
	mu           sync.RWMutex
}

// NewMusicSelector creates a new MusicSelector.
func NewMusicSelector() *MusicSelector {
	return &MusicSelector{
		musicFiles:   make([]string, 0),
		currentIndex: -1, // No initial selection
	}
}

// Update updates the list of music files, trying to preserve the current selection.
func (s *MusicSelector) Update(newFiles []string) (indexChanged bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	currentPath := ""
	if s.currentIndex >= 0 && s.currentIndex < len(s.musicFiles) {
		currentPath = s.musicFiles[s.currentIndex]
	}

	oldIndex := s.currentIndex
	s.musicFiles = newFiles
	newIndex := -1

	// Find the index of the preserved track in the new list
	if currentPath != "" {
		for i, file := range s.musicFiles {
			if file == currentPath {
				newIndex = i
				break
			}
		}
	}

	// If the current track wasn't found or the list is empty
	if newIndex == -1 {
		if len(s.musicFiles) > 0 {
			newIndex = 0 // Default to the first track
		} else {
			newIndex = -1 // No tracks available
		}
	}

	s.currentIndex = newIndex
	return oldIndex != s.currentIndex
}

// CurrentFile returns the path of the currently selected file and true if a valid selection exists.
func (s *MusicSelector) CurrentFile() (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.currentIndex >= 0 && s.currentIndex < len(s.musicFiles) {
		return s.musicFiles[s.currentIndex], true
	}
	return "", false
}

// Files returns a copy of the current music file list.
func (s *MusicSelector) Files() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	// Return a copy to prevent external modification
	filesCopy := make([]string, len(s.musicFiles))
	copy(filesCopy, s.musicFiles)
	return filesCopy
}

// SelectNext selects the next file in the list, looping back to the start if necessary.
// Returns true if the index changed.
func (s *MusicSelector) SelectNext() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.musicFiles) == 0 {
		s.currentIndex = -1
		return false // No change if list is empty
	}

	oldIndex := s.currentIndex
	s.currentIndex++
	if s.currentIndex >= len(s.musicFiles) {
		s.currentIndex = 0
	}
	return oldIndex != s.currentIndex
}

// SelectIndex attempts to select the file at the given index.
// Returns an error if the index is out of bounds.
func (s *MusicSelector) SelectIndex(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.musicFiles) {
		return fmt.Errorf("selector index out of range: %d (count: %d)", index, len(s.musicFiles))
	}
	s.currentIndex = index
	return nil
}

// CurrentIndex returns the current selection index.
func (s *MusicSelector) CurrentIndex() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentIndex
}

// --- MusicLoader ---

// MusicLoader handles loading audio streams from file paths.
type MusicLoader struct {
	// No fields needed for now, could add configuration later (e.g., sample rate)
}

// NewMusicLoader creates a new MusicLoader.
func NewMusicLoader() *MusicLoader {
	return &MusicLoader{}
}

// LoadStream opens and decodes an audio file from the given path.
// It returns a readable and seekable stream, or an error.
func (l *MusicLoader) LoadStream(filePath string) (io.ReadSeeker, error) {
	// Open the file
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("loader: failed to open audio file %s: %v", filePath, err)
	}

	// Decode based on file extension
	var audioStream io.ReadSeeker
	var decodeErr error

	if files.IsWavFile(filePath) {
		audioStream, decodeErr = wav.DecodeWithSampleRate(sampleRate, f)
	} else if files.IsOggFile(filePath) {
		audioStream, decodeErr = vorbis.DecodeWithSampleRate(sampleRate, f)
	} else if files.IsMp3File(filePath) {
		audioStream, decodeErr = mp3.DecodeWithSampleRate(sampleRate, f)
	} else {
		f.Close() // Close the file if format is unsupported
		return nil, fmt.Errorf("loader: unsupported audio format: %s", filePath)
	}

	if decodeErr != nil {
		f.Close() // Close the file if decoding fails
		return nil, fmt.Errorf("loader: failed to decode audio %s: %v", filePath, decodeErr)
	}

	// Note: The file 'f' is kept open by the stream decoder (wav, vorbis, mp3).
	// The stream (and thus the file) should be closed by the consumer (e.g., Player.Close).
	return audioStream, nil
}

// --- Constants & PlayerState ---

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

// Player interface abstracts audio player operations
type Player interface {
	Play()
	Pause()
	Close() error
	SetVolume(volume float64)
}

// PlayerFactory interface abstracts audio player creation
type PlayerFactory interface {
	NewPlayer(stream io.Reader) (Player, error)
}

// --- Music ---

// Music wraps a Player instance and holds metadata or state related to a specific track.
type Music struct {
	player Player // The underlying audio player
	// Future fields: isImpressive bool, notes string, etc.
}

// NewMusic creates a new Music instance wrapping a Player.
func NewMusic(player Player) *Music {
	if player == nil {
		return nil // Avoid creating Music with a nil player
	}
	return &Music{player: player}
}

// Close closes the underlying player.
func (m *Music) Close() error {
	if m.player == nil {
		return nil
	}
	return m.player.Close()
}

// Delegate methods to the underlying player

func (m *Music) Play() {
	if m.player != nil {
		m.player.Play()
	}
}

func (m *Music) Pause() {
	if m.player != nil {
		m.player.Pause()
	}
}

func (m *Music) SetVolume(volume float64) {
	if m.player != nil {
		m.player.SetVolume(volume)
	}
}

// --- MusicPlayer ---

// MusicPlayer handles music playback orchestration
type MusicPlayer struct {
	playerFactory PlayerFactory
	loader        *MusicLoader
	currentMusic  *Music        // Changed from player Player to currentMusic *Music
	audioStream   io.ReadSeeker // Keep track for potential explicit close if needed
	selector      *MusicSelector

	// Control variables
	state            PlayerState
	counter          int
	isPaused         bool
	loopDuration     float64 // in minutes
	intervalDuration float64 // in seconds
	volume           float64 // Current volume (0.0-1.0)
}

// NewMusicPlayer creates a new music player
func NewMusicPlayer(initialMusicFiles []string, playerFactory PlayerFactory) (*MusicPlayer, error) {
	// Create player components
	selector := NewMusicSelector()
	loader := NewMusicLoader() // Create loader

	player := &MusicPlayer{
		playerFactory: playerFactory,
		loader:        loader, // Assign loader
		selector:      selector,
		// currentMusic is initially nil
		state:            StateStopped,
		loopDuration:     5.0,
		intervalDuration: 10.0,
		volume:           1.0,
	}

	// Update selector with the initial list and potentially load the first track
	if selector.Update(initialMusicFiles) {
		if _, ok := selector.CurrentFile(); ok {
			if err := player.loadCurrentMusic(); err != nil {
				log.Printf("Warning: Failed to load initial track: %v", err)
				// Do not return error, player might recover if files are added/changed later
			}
		}
	}

	return player, nil // Return player even if initial load failed
}

// UpdateMusicFiles updates the music list and loads if necessary.
func (p *MusicPlayer) UpdateMusicFiles(newFiles []string) {
	indexChanged := p.selector.Update(newFiles)

	if indexChanged {
		if _, ok := p.selector.CurrentFile(); ok {
			if err := p.loadCurrentMusic(); err != nil {
				log.Printf("Failed to load music after file changes: %v", err)
			}
		} else {
			if p.currentMusic != nil {
				p.currentMusic.Close() // Close the wrapped player
				p.currentMusic = nil
			}
			p.state = StateStopped
			p.isPaused = false
		}
	}
}

// Close cleans up resources
func (p *MusicPlayer) Close() error {
	if p.currentMusic != nil {
		if err := p.currentMusic.Close(); err != nil { // Close the wrapped player
			return fmt.Errorf("failed to close music: %v", err)
		}
		p.currentMusic = nil
	}
	// audioStream might be managed by the player, but explicit close is safer if needed
	// if closer, ok := p.audioStream.(io.Closer); ok {
	// 	 closer.Close()
	// }
	return nil
}

// GetMusicFiles returns the list of music files from the selector.
func (p *MusicPlayer) GetMusicFiles() []string {
	return p.selector.Files()
}

// GetCurrentPath returns the path of the currently playing music from the selector.
func (p *MusicPlayer) GetCurrentPath() string {
	path, _ := p.selector.CurrentFile()
	return path
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

// GetCurrentIndex returns the current selection index from the selector.
func (p *MusicPlayer) GetCurrentIndex() int {
	return p.selector.CurrentIndex()
}

// SetCurrentIndex selects the music at the given index using the selector.
func (p *MusicPlayer) SetCurrentIndex(index int) error {
	if err := p.selector.SelectIndex(index); err != nil {
		return err
	}
	// If selection is successful, load the music
	return p.loadCurrentMusic()
}

// loadCurrentMusic loads the music indicated by the selector's current index.
func (p *MusicPlayer) loadCurrentMusic() error {
	currentPath, ok := p.selector.CurrentFile()
	if !ok {
		if p.currentMusic != nil {
			if err := p.currentMusic.Close(); err != nil {
				log.Printf("Error closing music while stopping: %v", err)
			}
			p.currentMusic = nil
		}
		p.state = StateStopped
		return fmt.Errorf("no music file selected")
	}

	// Close existing music/player if active
	if p.currentMusic != nil {
		if err := p.currentMusic.Close(); err != nil {
			log.Printf("Warning: failed to close previous music: %v", err)
		}
		p.currentMusic = nil
	}

	// Load the audio stream using the loader
	audioStream, err := p.loader.LoadStream(currentPath)
	if err != nil {
		return fmt.Errorf("failed to load audio stream for %s: %v", currentPath, err)
	}
	p.audioStream = audioStream // Keep track of the raw stream

	// Create infinite loop stream
	streamLength, ok := audioStream.(interface{ Length() int64 })
	if !ok {
		if closer, okCloser := audioStream.(io.Closer); okCloser {
			closer.Close()
		}
		return fmt.Errorf("loaded audio stream for %s does not support Length()", currentPath)
	}
	loopStream := audio.NewInfiniteLoop(audioStream, streamLength.Length())

	// Create the actual player instance
	newPlayer, err := p.playerFactory.NewPlayer(loopStream)
	if err != nil {
		if closer, okCloser := audioStream.(io.Closer); okCloser {
			closer.Close()
		}
		return fmt.Errorf("failed to create audio player for %s: %v", currentPath, err)
	}

	// Wrap the player in a Music struct
	p.currentMusic = NewMusic(newPlayer)
	if p.currentMusic == nil { // Should not happen if NewPlayer succeeded
		return fmt.Errorf("failed to wrap player in Music struct for %s", currentPath)
	}
	p.currentMusic.SetVolume(p.volume)

	// Reset counter and state
	p.counter = 0
	p.state = StatePlaying
	p.isPaused = false

	// Start playing
	p.currentMusic.Play()

	return nil
}

// TogglePause toggles pause state
func (p *MusicPlayer) TogglePause() {
	if p.currentMusic == nil { // Check currentMusic instead of player
		return
	}

	if p.isPaused {
		p.currentMusic.Play() // Delegate to Music
		p.isPaused = false
	} else {
		p.currentMusic.Pause() // Delegate to Music
		p.isPaused = true
	}
}

// Update updates the player state
func (p *MusicPlayer) Update() error {
	if p.currentMusic == nil || p.isPaused { // Check currentMusic
		return nil
	}

	p.counter++

	switch p.state {
	case StatePlaying:
		loopDurationFrames := int(p.loopDuration * 60 * 60)
		if p.counter >= loopDurationFrames {
			p.state = StateFadingOut
			p.counter = 0
		}

	case StateFadingOut:
		fadeOutFrames := int(fadeOutDuration.Seconds() * 60)
		if p.counter >= fadeOutFrames {
			p.state = StateInterval
			p.counter = 0
			if p.currentMusic != nil {
				p.currentMusic.Pause() // Pause the wrapped player
			}
		} else {
			fadeRatio := 1.0 - float64(p.counter)/float64(fadeOutFrames)
			p.volume = fadeRatio
			if p.currentMusic != nil {
				p.currentMusic.SetVolume(fadeRatio) // Set volume on Music
			}
		}

	case StateInterval:
		intervalFrames := int(p.intervalDuration * 60)
		if p.counter >= intervalFrames {
			p.volume = 1.0
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
	nextIndex := p.selector.SelectNext()
	if !nextIndex {
		return nil
	}

	p.volume = 1.0
	return p.loadCurrentMusic()
}

// TestSetPlayer is deprecated, use TestSetCurrentMusic
func (p *MusicPlayer) TestSetPlayer(player Player) {
	p.currentMusic = NewMusic(player)
}

// TestSetCurrentMusic directly sets the Music instance for testing
func (p *MusicPlayer) TestSetCurrentMusic(music *Music) {
	p.currentMusic = music
}
