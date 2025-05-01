package main

import (
	"image"
	"io"
	"log"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/guigui"

	"musicplayer/internal/files"
	"musicplayer/internal/player"
	"musicplayer/internal/ui"
)

// Sample rate for audio player
const sampleRate = 48000

// AudioContextWrapper wraps audio.Context to implement the player.PlayerFactory interface
type AudioContextWrapper struct {
	*audio.Context
}

// NewPlayer wraps audio.Context.NewPlayer to return a player.Player
func (w *AudioContextWrapper) NewPlayer(stream io.Reader) (player.Player, error) {
	p, err := w.Context.NewPlayer(stream)
	if err != nil {
		return nil, err
	}
	return p, nil
}

// Game represents the Ebiten game
type Game struct {
	player      *player.MusicPlayer
	warningText string
	watcher     *files.DirectoryWatcher
}

// NewGame creates a new game
func NewGame() (*Game, error) {
	// Set up music directory
	musicDir := files.DefaultMusicDir

	// Ensure the music directory exists
	absDir, err := musicDir.EnsureMusicDirectory()
	if err != nil {
		return nil, err
	}

	// Check if we have any music files (logging purposes)
	musicFiles, err := musicDir.FindMusicFiles()
	if err != nil {
		// Log warning but continue
		log.Printf("Warning: Failed to initially find music files: %v", err)
	}
	log.Printf("Found %d music files in %s", len(musicFiles), absDir)

	// Initialize audio context as PlayerFactory
	audioContext := audio.NewContext(sampleRate)

	// Create wrapper
	playerFactory := &AudioContextWrapper{Context: audioContext}

	// Initialize the music player with the initial list of files
	musicPlayer, err := player.NewMusicPlayer(musicFiles, playerFactory)
	if err != nil {
		// Log warning but continue as player might recover if files are added
		log.Printf("Warning: Failed to initialize music player: %v", err)
		// Ensure musicPlayer is nil if initialization truly failed, though NewMusicPlayer currently doesn't return errors
		// musicPlayer = nil
	}

	// Create and start the directory watcher
	watcher, err := musicDir.Watch()
	if err != nil {
		// Log warning but continue, file watching won't work
		log.Printf("Warning: Failed to start directory watcher: %v", err)
		watcher = nil // Ensure watcher is nil if creation failed
	}

	// Create and return the game
	g := &Game{
		player:  musicPlayer,
		watcher: watcher,
	}

	return g, nil
}

func main() {
	// Set up the game
	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to initialize game: %v", err)
	}

	// Ensure cleanup on exit
	defer func() {
		if game.player != nil {
			if err := game.player.Close(); err != nil {
				log.Printf("Error closing player: %v", err)
			}
		}
		// Close the watcher as well
		if game.watcher != nil {
			if err := game.watcher.Close(); err != nil {
				log.Printf("Error closing watcher: %v", err)
			}
		}
	}()

	// Create the root widget
	root := ui.NewRoot(game.player)

	// ---- Connect Watcher to Root's Handler ----
	if game.watcher != nil {
		// Add Root's HandleFileChanges as a handler
		game.watcher.AddHandler(root.HandleFileChanges)

		// Optionally trigger initial notification if needed,
		// although NewRoot already handles initial state.
		// game.watcher.NotifyChange() // Depends on DirectoryWatcher implementation
	}
	// ---- End Connection ----

	// Run the application with guigui
	op := &guigui.RunOptions{
		Title:      "Music asset tester",
		WindowSize: image.Point{X: ui.ScreenWidth, Y: ui.ScreenHeight},
	}

	if err := guigui.Run(root, op); err != nil {
		log.Fatalf("Error running game: %v", err)
	}
}
