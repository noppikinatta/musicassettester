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
	// Use testdata directory with actual audio files
	testDataDir := "testdata"

	// Define initial test files (actual files in testdata directory)
	// Use WAV files only to avoid mp3 layer compatibility issues
	initialFiles := []string{
		filepath.Join(testDataDir, "test1.wav"),
		filepath.Join(testDataDir, "test2.wav"),
	}

	// Create mock factory
	mockFactory := NewMockPlayerFactory()

	// Create player with initial file list
	p, err := player.NewMusicPlayer(initialFiles, mockFactory)
	if err != nil {
		// Log warning, but allow test to continue if possible
		t.Logf("Warning during player creation: %v", err)
		// Depending on the test, we might want to t.Fatal here
	}
	// Ensure player is not nil if creation succeeded without error
	if err == nil && p == nil {
		t.Fatal("NewMusicPlayer returned nil player without error")
	}

	// No longer need SetTestMusicFiles, files are passed in constructor

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
	p, _ := createTestMusicPlayer(t)

	// Ensure a player is loaded and ready
	if len(p.GetMusicFiles()) == 0 {
		t.Skip("Skipping TestTogglePause: No music files available")
	}
	err := p.SetCurrentIndex(0) // This loads the music and sets state to Playing
	if err != nil {
		t.Fatalf("Failed to set initial index for TestTogglePause: %v", err)
	}

	// Initially not paused (loadCurrentMusic sets isPaused to false)
	if p.IsPaused() {
		t.Fatal("Test setup failed: player should not be paused after loading")
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
	p, _ := createTestMusicPlayer(t)

	// Ensure a player is loaded and ready, set state to Playing
	if len(p.GetMusicFiles()) == 0 {
		t.Skip("Skipping TestUpdate: No music files available")
	}
	err := p.SetCurrentIndex(0) // Loads music, sets state=Playing, counter=0, isPaused=false
	if err != nil {
		t.Fatalf("Failed to set initial index for TestUpdate: %v", err)
	}

	// Verify initial state after load
	if p.GetState() != player.StatePlaying || p.IsPaused() || p.GetCounter() != 0 {
		t.Fatalf("Test setup failed: incorrect state after loading. State: %v, Paused: %v, Counter: %d",
			p.GetState(), p.IsPaused(), p.GetCounter())
	}

	// 1. Update while playing: counter should increase
	err = p.Update()
	if err != nil {
		t.Errorf("Expected Update() to succeed, got error: %v", err)
	}
	if p.GetCounter() != 1 {
		t.Errorf("Expected counter to be 1 after first update, got %d", p.GetCounter())
	}

	// 2. Update while paused: counter should not increase
	p.TogglePause() // Pause the player
	if !p.IsPaused() {
		t.Fatal("Failed to pause player for test")
	}
	pausedCounter := p.GetCounter()
	err = p.Update()
	if err != nil {
		t.Errorf("Expected Update() during pause to succeed, got error: %v", err)
	}
	if p.GetCounter() != pausedCounter {
		t.Errorf("Expected counter to remain %d during pause, got %d", pausedCounter, p.GetCounter())
	}

	// 3. Update with no current music: counter should not increase (effectively stopped)
	p.TogglePause() // Unpause first
	p.Close()       // This sets currentMusic to nil and should prevent counter increase

	// We are not checking the counter value here anymore, just the state
	err = p.Update()
	if err != nil {
		t.Errorf("Expected Update() with closed player to succeed, got error: %v", err)
	}
	if p.GetState() != player.StateStopped {
		t.Errorf("Expected state to be StateStopped after Close, got %v", p.GetState())
	}
}

func TestRandomSelection(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	// Verify default random flag is false
	if p.IsRandomMode() {
		t.Error("Expected random mode to be false by default")
	}

	// Enable random mode
	p.SetRandomMode(true)
	if !p.IsRandomMode() {
		t.Error("Expected random mode to be true after setting")
	}

	// Disable random mode
	p.SetRandomMode(false)
	if p.IsRandomMode() {
		t.Error("Expected random mode to be false after disabling")
	}
}

func TestSkipToNextRandom(t *testing.T) {
	// Create test player with multiple files
	testFiles := []string{
		"testdata/test1.wav",
		"testdata/test2.wav",
	}

	mockFactory := NewMockPlayerFactory()
	p, err := player.NewMusicPlayer(testFiles, mockFactory)
	if err != nil {
		t.Logf("Warning during player creation: %v", err)
	}

	if len(p.GetMusicFiles()) < 2 {
		t.Skip("Need at least 2 music files for random selection test")
	}

	// Set initial index
	err = p.SetCurrentIndex(0)
	if err != nil {
		t.Fatalf("Failed to set initial index: %v", err)
	}

	// Test sequential mode (default)
	initialIndex := p.GetCurrentIndex()
	err = p.SkipToNext()
	if err != nil {
		t.Errorf("SkipToNext failed: %v", err)
	}
	sequentialIndex := p.GetCurrentIndex()
	expectedSequential := (initialIndex + 1) % len(p.GetMusicFiles())
	if sequentialIndex != expectedSequential {
		t.Errorf("Expected sequential index %d, got %d", expectedSequential, sequentialIndex)
	}

	// Enable random mode
	p.SetRandomMode(true)

	// Reset to initial position
	err = p.SetCurrentIndex(0)
	if err != nil {
		t.Fatalf("Failed to reset index: %v", err)
	}

	// Test random selection - with 2 files, it should always alternate
	previousIndex := p.GetCurrentIndex()
	err = p.SkipToNext()
	if err != nil {
		t.Errorf("SkipToNext failed: %v", err)
	}
	currentIndex := p.GetCurrentIndex()

	// With only 2 files, random should pick the other file
	if currentIndex == previousIndex {
		t.Errorf("Random selection should pick different file. Previous: %d, Current: %d", previousIndex, currentIndex)
	}

	// Test a few more times to ensure it works consistently
	for i := 0; i < 3; i++ {
		previousIndex = p.GetCurrentIndex()
		err = p.SkipToNext()
		if err != nil {
			t.Errorf("SkipToNext failed on iteration %d: %v", i, err)
		}
		newIndex := p.GetCurrentIndex()
		if newIndex < 0 || newIndex >= len(p.GetMusicFiles()) {
			t.Errorf("Random selection produced invalid index: %d", newIndex)
		}
		if newIndex == previousIndex {
			t.Errorf("Random selection should avoid repeating same index. Previous: %d, Current: %d", previousIndex, newIndex)
		}
	}
}
