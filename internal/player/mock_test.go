package player_test

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"musicplayer/internal/files"
	"musicplayer/internal/player"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// MockAudioPlayer implements the player.Player interface for testing
type MockAudioPlayer struct {
	volumeValue float64
	isPlaying   bool
	mu          sync.Mutex
}

func NewMockAudioPlayer() *MockAudioPlayer {
	return &MockAudioPlayer{
		volumeValue: 1.0,
		isPlaying:   false,
	}
}

func (m *MockAudioPlayer) Play() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isPlaying = true
}

func (m *MockAudioPlayer) Pause() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.isPlaying = false
}

func (m *MockAudioPlayer) IsPlaying() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isPlaying
}

func (m *MockAudioPlayer) SetVolume(volume float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.volumeValue = volume
}

func (m *MockAudioPlayer) Volume() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.volumeValue
}

func (m *MockAudioPlayer) Current() int64 {
	return 0
}

func (m *MockAudioPlayer) Rewind() error {
	return nil
}

func (m *MockAudioPlayer) Close() error {
	return nil
}

// MockAudioContext implements the audio.Context interface for testing
type MockAudioContext struct {
	sampleRate int
	nextPlayer *MockAudioPlayer
}

func NewMockAudioContext(sampleRate int) *MockAudioContext {
	return &MockAudioContext{
		sampleRate: sampleRate,
		nextPlayer: NewMockAudioPlayer(),
	}
}

func (m *MockAudioContext) NewPlayer(stream io.Reader) (*audio.Player, error) {
	// このモックでは実際のaudio.Playerではなく、常にnilを返します
	// これはテスト環境では実際のオーディオデバイスにアクセスできないためです
	return nil, nil
}

// MockPlayerFactory implements the player.PlayerFactory interface for testing
type MockPlayerFactory struct {
	audioPlayers []*MockAudioPlayer
}

func NewMockPlayerFactory() *MockPlayerFactory {
	return &MockPlayerFactory{
		audioPlayers: make([]*MockAudioPlayer, 0),
	}
}

func (f *MockPlayerFactory) NewPlayer(stream io.Reader) (player.Player, error) {
	// テスト用のモックプレイヤーを作成
	mockPlayer := NewMockAudioPlayer()
	f.audioPlayers = append(f.audioPlayers, mockPlayer)

	// player.Playerインターフェースとして返す
	return mockPlayer, nil
}

// GetLastPlayer returns the last created mock player
func (f *MockPlayerFactory) GetLastPlayer() *MockAudioPlayer {
	if len(f.audioPlayers) == 0 {
		// テストのために常にモックプレイヤーを返す
		return NewMockAudioPlayer()
	}
	return f.audioPlayers[len(f.audioPlayers)-1]
}

// MockReadSeeker implements io.ReadSeeker for testing
type MockReadSeeker struct {
	data        []byte
	currentPos  int64
	lengthValue int64
}

func NewMockReadSeeker(data []byte) *MockReadSeeker {
	return &MockReadSeeker{
		data:        data,
		currentPos:  0,
		lengthValue: int64(len(data)),
	}
}

func (m *MockReadSeeker) Read(p []byte) (n int, err error) {
	if m.currentPos >= int64(len(m.data)) {
		return 0, io.EOF
	}
	n = copy(p, m.data[m.currentPos:])
	m.currentPos += int64(n)
	return n, nil
}

func (m *MockReadSeeker) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		m.currentPos = offset
	case io.SeekCurrent:
		m.currentPos += offset
	case io.SeekEnd:
		m.currentPos = int64(len(m.data)) + offset
	}
	return m.currentPos, nil
}

func (m *MockReadSeeker) Length() int64 {
	return m.lengthValue
}

// 文字列配列を格納するテスト用のMusicDirectoryの実装
type TestMusicDirectory string

// NewMockMusicDirectory creates a mock music directory for testing
func NewMockMusicDirectory(path string, fileList []string, exists bool) files.MusicDirectory {
	return files.MusicDirectory(path)
}

// Path returns the directory path as a string
func (md TestMusicDirectory) Path() string {
	return string(md)
}

// Abs returns the absolute path of the music directory
func (md TestMusicDirectory) Abs() (string, error) {
	return filepath.Abs(md.Path())
}

// FindMusicFiles searches for music files in the music directory
func (md TestMusicDirectory) FindMusicFiles() ([]string, error) {
	// ディレクトリが存在するかチェック
	if _, err := os.Stat(md.Path()); os.IsNotExist(err) {
		if err := os.MkdirAll(md.Path(), 0755); err != nil {
			return nil, err
		}
		return []string{}, nil
	}

	// TestHelper内のSetupTestFilesを使用して実際のファイルを作成
	h := TestHelper{}
	files, cleanup, err := h.SetupTestFiles(md.Path())
	if err != nil {
		return nil, err
	}
	defer cleanup()

	return files, nil
}

// EnsureMusicDirectory ensures that the music directory exists
func (md TestMusicDirectory) EnsureMusicDirectory() (string, error) {
	if _, err := os.Stat(md.Path()); os.IsNotExist(err) {
		if err := os.MkdirAll(md.Path(), 0755); err != nil {
			return "", err
		}
	}
	return md.Path(), nil
}

// GetUsageInstructions returns instructions for using the application
func (md TestMusicDirectory) GetUsageInstructions() string {
	return fmt.Sprintf(`No music files found in the '%s' directory.

Instructions:
1. Place .wav, .ogg, or .mp3 files in the '%s' directory
2. Restart the application
3. Use the list to select and play music
4. Space: Toggle pause
5. N: Skip to next track
6. Use sliders to adjust loop and interval durations
`, md.Path(), md.Path())
}

// TestHelper contains functions that help with testing
type TestHelper struct{}

// SetupTestFiles creates test audio files for testing
func (h *TestHelper) SetupTestFiles(dir string) ([]string, func(), error) {
	// Create test directory if it doesn't exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, nil, err
		}
	}

	// Create test files
	testFiles := []struct {
		name    string
		content []byte
	}{
		{"test1.wav", []byte("test wav data")},
		{"test2.mp3", []byte("test mp3 data")},
		{"test3.ogg", []byte("test ogg data")},
		{"subdir/test4.wav", []byte("test subdir wav data")},
	}

	// Create files
	var createdFiles []string
	for _, tf := range testFiles {
		// Create directories if needed
		filePath := filepath.Join(dir, tf.name)
		fileDir := filepath.Dir(filePath)
		if err := os.MkdirAll(fileDir, 0755); err != nil {
			return nil, nil, err
		}

		// Create file
		err := os.WriteFile(filePath, tf.content, 0644)
		if err != nil {
			return nil, nil, err
		}
		createdFiles = append(createdFiles, filePath)
	}

	// Return cleanup function
	cleanup := func() {
		os.RemoveAll(dir)
	}

	return createdFiles, cleanup, nil
}
