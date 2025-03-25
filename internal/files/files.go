package files

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// MusicDirectory represents a directory where music files are stored
type MusicDirectory string

// DefaultMusicDir is the default music directory path
const DefaultMusicDir MusicDirectory = "musics"

// FileChangeCallback is a function type for file change notifications
type FileChangeCallback func([]string)

// DirectoryWatcher watches for changes in the music directory
type DirectoryWatcher struct {
	watcher     *fsnotify.Watcher
	callback    FileChangeCallback
	debounceMap map[string]time.Time
	mu          sync.Mutex
	done        chan struct{}
}

// NewDirectoryWatcher creates a new directory watcher
func NewDirectoryWatcher(callback FileChangeCallback) (*DirectoryWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %v", err)
	}

	dw := &DirectoryWatcher{
		watcher:     watcher,
		callback:    callback,
		debounceMap: make(map[string]time.Time),
		done:        make(chan struct{}),
	}

	go dw.watchLoop()
	return dw, nil
}

// watchLoop handles file system events
func (dw *DirectoryWatcher) watchLoop() {
	const debounceInterval = 500 * time.Millisecond

	for {
		select {
		case event, ok := <-dw.watcher.Events:
			if !ok {
				return
			}

			// Skip temporary files and directories
			if strings.HasPrefix(filepath.Base(event.Name), ".") {
				continue
			}

			// Handle the event
			if event.Op&(fsnotify.Create|fsnotify.Remove) != 0 {
				dw.mu.Lock()
				lastEvent, exists := dw.debounceMap[event.Name]
				now := time.Now()
				dw.debounceMap[event.Name] = now
				dw.mu.Unlock()

				// Debounce events
				if exists && now.Sub(lastEvent) < debounceInterval {
					continue
				}

				// If a directory is created, watch it
				if event.Op&fsnotify.Create != 0 {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						dw.watchDirectory(event.Name)
					}
				}

				// Notify about the change
				go dw.notifyChange()
			}

		case err, ok := <-dw.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Error watching directory: %v\n", err)

		case <-dw.done:
			return
		}
	}
}

// watchDirectory adds a directory and its subdirectories to the watch list
func (dw *DirectoryWatcher) watchDirectory(dir string) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return dw.watcher.Add(path)
		}
		return nil
	})
}

// notifyChange notifies the callback with updated file list
func (dw *DirectoryWatcher) notifyChange() {
	// Get the updated file list
	files, err := DefaultMusicDir.FindMusicFiles()
	if err != nil {
		fmt.Printf("Error finding music files: %v\n", err)
		return
	}

	// Notify the callback
	if dw.callback != nil {
		dw.callback(files)
	}
}

// Close stops watching and cleans up resources
func (dw *DirectoryWatcher) Close() error {
	close(dw.done)
	return dw.watcher.Close()
}

// Watch starts watching the music directory for changes
func (md MusicDirectory) Watch(callback FileChangeCallback) (*DirectoryWatcher, error) {
	// Create watcher
	dw, err := NewDirectoryWatcher(callback)
	if err != nil {
		return nil, err
	}

	// Ensure directory exists
	dir, err := md.EnsureMusicDirectory()
	if err != nil {
		dw.Close()
		return nil, err
	}

	// Start watching the directory
	if err := dw.watchDirectory(dir); err != nil {
		dw.Close()
		return nil, fmt.Errorf("failed to watch directory: %v", err)
	}

	return dw, nil
}

// IsWavFile checks if the file is a WAV file
func IsWavFile(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".wav"
}

// IsOggFile checks if the file is an OGG file
func IsOggFile(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".ogg"
}

// IsMp3File checks if the file is an MP3 file
func IsMp3File(path string) bool {
	return strings.ToLower(filepath.Ext(path)) == ".mp3"
}

// Path returns the directory path as a string
func (md MusicDirectory) Path() string {
	return string(md)
}

// Abs returns the absolute path of the music directory
func (md MusicDirectory) Abs() (string, error) {
	return filepath.Abs(md.Path())
}

// FindMusicFiles searches for music files in the music directory
func (md MusicDirectory) FindMusicFiles() ([]string, error) {
	musicFiles := []string{}

	// Check if the directory exists
	if _, err := os.Stat(md.Path()); os.IsNotExist(err) {
		if err := os.MkdirAll(md.Path(), 0755); err != nil {
			return nil, fmt.Errorf("failed to create music directory: %v", err)
		}
		return musicFiles, nil
	}

	// Walk through the music directory
	err := filepath.Walk(md.Path(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Check if the file is a supported audio file
		if IsWavFile(path) || IsOggFile(path) || IsMp3File(path) {
			// Add the file to the list
			musicFiles = append(musicFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk music directory: %v", err)
	}

	return musicFiles, nil
}

// EnsureMusicDirectory ensures that the music directory exists
func (md MusicDirectory) EnsureMusicDirectory() (string, error) {
	// Create the music directory if it doesn't exist
	musicDir, err := md.Abs()
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %v", err)
	}

	if _, err := os.Stat(musicDir); os.IsNotExist(err) {
		if err := os.MkdirAll(musicDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create music directory: %v", err)
		}
	}

	return musicDir, nil
}

// GetUsageInstructions returns instructions for using the application
func (md MusicDirectory) GetUsageInstructions() string {
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

// GetHowToUseMessage returns instruction message about required files
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

// Keep the original functions for compatibility
// FindMusicFiles searches for music files in the default music directory
func FindMusicFiles() ([]string, error) {
	return DefaultMusicDir.FindMusicFiles()
}

// EnsureMusicDirectory ensures that the default music directory exists
func EnsureMusicDirectory() (string, error) {
	return DefaultMusicDir.EnsureMusicDirectory()
}

// GetUsageInstructions returns instructions for using the application
func GetUsageInstructions() string {
	return DefaultMusicDir.GetUsageInstructions()
}
