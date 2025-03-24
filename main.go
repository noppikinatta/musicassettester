package main

import (
	"io"
	"log"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/guigui"

	"musicplayer/internal/files"
	"musicplayer/internal/player"
	"musicplayer/internal/ui"
)

// 音声プレイヤーのサンプルレート
const sampleRate = 48000

// AudioPlayerWrapper は ebiten の audio.Player を player.Player インターフェースに適合させるラッパーです
type AudioPlayerWrapper struct {
	*audio.Player
}

// SetVolume は音量を設定します（インターフェースの実装）
func (w *AudioPlayerWrapper) SetVolume(volume float64) {
	w.Player.SetVolume(volume)
}

// AudioContextWrapper は audio.Context を player.PlayerFactory インターフェースに適合させるラッパーです
type AudioContextWrapper struct {
	*audio.Context
}

// NewPlayer は audio.Context.NewPlayer をラップして player.Player を返します
func (w *AudioContextWrapper) NewPlayer(stream io.Reader) (player.Player, error) {
	p, err := w.Context.NewPlayer(stream)
	if err != nil {
		return nil, err
	}
	return &AudioPlayerWrapper{Player: p}, nil
}

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

	// ラッパーを作成
	playerFactory := &AudioContextWrapper{Context: audioContext}

	// Initialize the music player
	musicPlayer, err := player.NewMusicPlayer(musicDir, playerFactory)
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
