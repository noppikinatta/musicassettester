package main

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/guigui"

	"musicplayer/internal/files"
	"musicplayer/internal/player"
	"musicplayer/internal/ui"
)

// 音声プレイヤーのサンプルレート
const sampleRate = 48000

// Game represents the Ebiten game
type Game struct {
	player      *player.MusicPlayer
	warningText string
}

// NewGame creates a new game
func NewGame() (*Game, error) {
	// 音楽ディレクトリの設定
	musicDir := files.DefaultMusicDir

	// Ensure the music directory exists
	absDir, err := musicDir.EnsureMusicDirectory()
	if err != nil {
		return nil, err
	}

	// Check if we have any music files
	musicFiles, err := musicDir.FindMusicFiles()
	if err != nil {
		return nil, err
	}
	log.Printf("Found %d music files in %s", len(musicFiles), absDir)

	// Initialize audio context as PlayerFactory
	audioContext := audio.NewContext(sampleRate)

	// Initialize the music player
	musicPlayer, err := player.NewMusicPlayer(musicDir, audioContext)
	if err != nil {
		log.Printf("Warning: %v", err)
	}

	// Set warning message if no music files
	warningText := ""
	if len(musicFiles) == 0 {
		warningText = musicDir.GetUsageInstructions()
	}

	// Create and return the game
	g := &Game{
		player:      musicPlayer,
		warningText: warningText,
	}

	return g, nil
}

func main() {
	// Set up the game
	game, err := NewGame()
	if err != nil {
		log.Fatalf("Failed to initialize game: %v", err)
	}

	// Create the root widget
	root := ui.NewRoot(game.player, game.warningText)

	// Run the application with guigui
	op := &guigui.RunOptions{
		Title:           "Music Cassette Tester",
		WindowMinWidth:  ui.ScreenWidth,
		WindowMinHeight: ui.ScreenHeight,
	}

	if err := guigui.Run(root, op); err != nil {
		log.Fatalf("Error running game: %v", err)
	}
}
