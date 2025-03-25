package player_test

import (
	"musicplayer/internal/player"
	"os"
	"path/filepath"
	"testing"
)

// TestMain handles the setup for all tests
func TestMain(m *testing.M) {
	// Perform test setup
	// Add additional setup here if needed

	// Run tests
	code := m.Run()

	// Perform test cleanup
	// Add additional cleanup here if needed

	os.Exit(code)
}

// Helper function to create a test MusicPlayer
func createTestMusicPlayer(t *testing.T) (*player.MusicPlayer, *MockPlayerFactory) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "music-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	// Create mock directory and factory
	mockDir := NewMockMusicDirectory(tempDir, []string{}, true)
	mockFactory := NewMockPlayerFactory()

	// Create player
	p, err := player.NewMusicPlayer(mockDir, mockFactory)
	if err != nil {
		t.Logf("Warning during player creation: %v", err)
	}

	// Add test music files to the player (actual files are not created)
	p.SetTestMusicFiles([]string{
		filepath.Join(tempDir, "test1.mp3"),
		filepath.Join(tempDir, "test2.wav"),
	})

	return p, mockFactory
}

func TestNewMusicPlayer(t *testing.T) {
	p, mockFactory := createTestMusicPlayer(t)

	if p == nil {
		t.Error("Expected player instance, got nil")
	}

	state := p.GetState()
	if state != player.StateStopped { // Initial state should be Stopped
		t.Errorf("Expected state %d, got %d", player.StateStopped, state)
	}

	// Verify that player was created via the factory
	mockPlayer := mockFactory.GetLastPlayer()
	if mockPlayer == nil {
		t.Error("Expected player to be created, but none was found")
	}
}

func TestGetMusicFiles(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	files := p.GetMusicFiles()
	if len(files) == 0 {
		t.Error("Expected music files to be set, got empty list")
	}
}

func TestGetCurrentPath(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	path := p.GetCurrentPath()
	// Initially empty or set during the test
	t.Logf("Current path: %s", path)
}

func TestGetState(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	state := p.GetState()
	if state != player.StateStopped {
		t.Errorf("Expected initial state to be StateStopped (%d), got %d", player.StateStopped, state)
	}
}

func TestIsPaused(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	isPaused := p.IsPaused()
	if isPaused {
		t.Error("Expected player to not be paused initially")
	}
}

func TestGetCounter(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	counter := p.GetCounter()
	// Initial counter should be 0
	if counter != 0 {
		t.Errorf("Expected initial counter to be 0, got %d", counter)
	}
}

func TestLoopDurationMinutes(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	// Verify default value
	duration := p.GetLoopDurationMinutes()
	if duration != 5.0 {
		t.Errorf("Expected default loop duration to be 5.0 minutes, got %f", duration)
	}

	// Change value and verify
	p.SetLoopDurationMinutes(10.0)
	duration = p.GetLoopDurationMinutes()
	if duration != 10.0 {
		t.Errorf("Expected loop duration to be 10.0 minutes after setting, got %f", duration)
	}
}

func TestIntervalSeconds(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	// Verify default value
	interval := p.GetIntervalSeconds()
	if interval != 10.0 {
		t.Errorf("Expected default interval to be 10.0 seconds, got %f", interval)
	}

	// Change value and verify
	p.SetIntervalSeconds(15.0)
	interval = p.GetIntervalSeconds()
	if interval != 15.0 {
		t.Errorf("Expected interval to be 15.0 seconds after setting, got %f", interval)
	}
}

func TestSetCurrentIndex(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	// Set a valid index
	if len(p.GetMusicFiles()) > 0 {
		err := p.SetCurrentIndex(0)
		if err != nil {
			t.Errorf("Expected SetCurrentIndex(0) to succeed, got error: %v", err)
		}

		// Set an out-of-range index
		err = p.SetCurrentIndex(100)
		if err == nil {
			t.Error("Expected SetCurrentIndex(100) to fail, but it succeeded")
		}
	} else {
		t.Skip("No music files available for this test")
	}
}

func TestTogglePause(t *testing.T) {
	p, mockFactory := createTestMusicPlayer(t)

	// Player instance is needed, so set the current Player to nil and create a new mock player
	mockPlayer := mockFactory.GetLastPlayer()

	// Set MusicPlayer's internal state directly for testing
	p.TestSetPlayer(mockPlayer)
	p.TestSetPaused(false)

	// Initially not paused
	if p.IsPaused() {
		t.Error("Test setup failed: player should not be paused initially")
	}

	// Pause
	p.TogglePause()
	if !p.IsPaused() {
		t.Error("Expected player to be paused after TogglePause")
	}

	// Resume
	p.TogglePause()
	if p.IsPaused() {
		t.Error("Expected player to not be paused after second TogglePause")
	}
}

func TestUpdate(t *testing.T) {
	p, mockFactory := createTestMusicPlayer(t)

	// Set MusicPlayer's internal state directly
	mockPlayer := mockFactory.GetLastPlayer()
	p.TestSetPlayer(mockPlayer)
	p.TestSetCounter(0)
	p.TestSetPaused(false)
	p.TestSetState(player.StatePlaying)

	// Verify initial counter
	if p.GetCounter() != 0 {
		t.Errorf("Test setup failed: expected initial counter to be 0, got %d", p.GetCounter())
	}

	// Update - counter increases in Player existence and non-paused state
	err := p.Update()
	if err != nil {
		t.Errorf("Expected Update() to succeed, got error: %v", err)
	}

	if p.GetCounter() != 1 {
		t.Errorf("Expected counter to be 1 after update, got %d", p.GetCounter())
	}

	// Counter does not update during pause
	p.TestSetPaused(true)
	initialCounter := p.GetCounter() // Save current counter value

	err = p.Update()
	if err != nil {
		t.Errorf("Expected Update() during pause to succeed, got error: %v", err)
	}

	if p.GetCounter() != initialCounter {
		t.Errorf("Expected counter to remain %d during pause, got %d",
			initialCounter, p.GetCounter())
	}

	// Counter does not update if Player is nil
	p.TestSetPaused(false) // Unpause
	p.TestSetPlayer(nil)   // Set Player to nil
	initialCounter = p.GetCounter()

	err = p.Update()
	if err != nil {
		t.Errorf("Expected Update() with nil player to succeed, got error: %v", err)
	}

	if p.GetCounter() != initialCounter {
		t.Errorf("Expected counter to remain %d with nil player, got %d",
			initialCounter, p.GetCounter())
	}
}
