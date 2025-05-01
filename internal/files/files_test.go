package files_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"musicplayer/internal/files"
)

// TestIsWavFile tests the IsWavFile function
func TestIsWavFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Standard WAV file", "test.wav", true},
		{"Uppercase extension", "test.WAV", true},
		{"Mixed case extension", "test.WaV", true},
		{"No extension", "testwav", false},
		{"Different extension", "test.mp3", false},
		{"Path with dots", "/path/to/test.wav", true},
		{"Windows path", "C:\\path\\to\\test.wav", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := files.IsWavFile(tt.path)
			if result != tt.expected {
				t.Errorf("IsWavFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestIsOggFile tests the IsOggFile function
func TestIsOggFile(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Standard OGG file", "test.ogg", true},
		{"Uppercase extension", "test.OGG", true},
		{"Mixed case extension", "test.OgG", true},
		{"No extension", "testogg", false},
		{"Different extension", "test.wav", false},
		{"Path with dots", "/path/to/test.ogg", true},
		{"Windows path", "C:\\path\\to\\test.ogg", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := files.IsOggFile(tt.path)
			if result != tt.expected {
				t.Errorf("IsOggFile(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestIsMp3File tests the IsMp3File function
func TestIsMp3File(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"Standard MP3 file", "test.mp3", true},
		{"Uppercase extension", "test.MP3", true},
		{"Mixed case extension", "test.Mp3", true},
		{"No extension", "testmp3", false},
		{"Different extension", "test.wav", false},
		{"Path with dots", "/path/to/test.mp3", true},
		{"Windows path", "C:\\path\\to\\test.mp3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := files.IsMp3File(tt.path)
			if result != tt.expected {
				t.Errorf("IsMp3File(%q) = %v, want %v", tt.path, result, tt.expected)
			}
		})
	}
}

// TestMusicDirectory_Path tests the Path method
func TestMusicDirectory_Path(t *testing.T) {
	md := files.MusicDirectory("test_dir")
	if md.Path() != "test_dir" {
		t.Errorf("MusicDirectory.Path() = %q, want %q", md.Path(), "test_dir")
	}
}

// TestMusicDirectory_Abs tests the Abs method
func TestMusicDirectory_Abs(t *testing.T) {
	md := files.MusicDirectory("test_dir")
	absPath, err := md.Abs()
	if err != nil {
		t.Fatalf("MusicDirectory.Abs() error = %v", err)
	}

	expected, err := filepath.Abs("test_dir")
	if err != nil {
		t.Fatalf("filepath.Abs() error = %v", err)
	}

	if absPath != expected {
		t.Errorf("MusicDirectory.Abs() = %q, want %q", absPath, expected)
	}
}

// TestMusicDirectory_FindMusicFiles tests the FindMusicFiles method
func TestMusicDirectory_FindMusicFiles(t *testing.T) {
	t.Run("Search for files in existing testdata directory", func(t *testing.T) {
		// Set testdata directory as the test directory
		md := files.MusicDirectory("testdata")

		// Execute test
		foundFiles, err := md.FindMusicFiles()
		if err != nil {
			t.Fatalf("MusicDirectory.FindMusicFiles() error = %v", err)
		}

		// Check results (4 music files: sample.wav, sample.ogg, sample.mp3, subdir/sample.wav)
		expectedCount := 4
		if len(foundFiles) != expectedCount {
			t.Errorf("MusicDirectory.FindMusicFiles() got %d files, want %d", len(foundFiles), expectedCount)
		}

		// Check filenames
		fileMap := make(map[string]bool)
		for _, file := range foundFiles {
			fileMap[filepath.Base(file)] = true
		}

		// Check if files in subdirectories are included
		subdirFound := false
		for _, file := range foundFiles {
			if strings.Contains(file, "subdir") {
				subdirFound = true
				break
			}
		}
		if !subdirFound {
			t.Errorf("MusicDirectory.FindMusicFiles() should find files in subdirectories")
		}

		// Check that non-music files (.txt) are not included
		for _, file := range foundFiles {
			if strings.HasSuffix(file, ".txt") {
				t.Errorf("MusicDirectory.FindMusicFiles() should not include non-music files: %s", file)
			}
		}
	})

	t.Run("For non-existent directory", func(t *testing.T) {
		// Generate a temporary random directory name
		tempDirName := "non_existent_dir_" + filepath.Base(t.TempDir())
		md := files.MusicDirectory(tempDirName)

		// Delete directory when test completes
		defer os.RemoveAll(tempDirName)

		// Execute test
		foundFiles, err := md.FindMusicFiles()
		if err != nil {
			t.Fatalf("MusicDirectory.FindMusicFiles() with non-existent dir error = %v", err)
		}

		// Check if an empty array is returned
		if len(foundFiles) != 0 {
			t.Errorf("MusicDirectory.FindMusicFiles() with non-existent dir got %d files, want 0", len(foundFiles))
		}

		// Check if directory was automatically created
		if _, err := os.Stat(tempDirName); os.IsNotExist(err) {
			t.Errorf("Directory was not created automatically")
		}
	})
}

// TestMusicDirectory_EnsureMusicDirectory tests the EnsureMusicDirectory method
func TestMusicDirectory_EnsureMusicDirectory(t *testing.T) {
	t.Run("Create non-existent directory", func(t *testing.T) {
		// Generate a temporary random directory name
		tempDirName := "temp_music_dir_" + filepath.Base(t.TempDir())
		md := files.MusicDirectory(tempDirName)

		// Delete directory when test completes
		defer os.RemoveAll(tempDirName)

		// Execute test
		musicDir, err := md.EnsureMusicDirectory()
		if err != nil {
			t.Fatalf("MusicDirectory.EnsureMusicDirectory() error = %v", err)
		}

		// Check if the returned absolute path is correct
		expectedAbsPath, _ := filepath.Abs(tempDirName)
		if musicDir != expectedAbsPath {
			t.Errorf("MusicDirectory.EnsureMusicDirectory() got %s, want %s", musicDir, expectedAbsPath)
		}

		// Check if directory was actually created
		if _, err := os.Stat(tempDirName); os.IsNotExist(err) {
			t.Errorf("MusicDirectory.EnsureMusicDirectory() did not create directory")
		}
	})

	t.Run("For existing directory", func(t *testing.T) {
		// Use testdata directory
		md := files.MusicDirectory("testdata")

		// Execute test
		musicDir, err := md.EnsureMusicDirectory()
		if err != nil {
			t.Fatalf("MusicDirectory.EnsureMusicDirectory() with existing dir error = %v", err)
		}

		// Check if the returned absolute path is correct
		expectedAbsPath, _ := filepath.Abs("testdata")
		if musicDir != expectedAbsPath {
			t.Errorf("MusicDirectory.EnsureMusicDirectory() with existing dir got %s, want %s", musicDir, expectedAbsPath)
		}
	})
}

// TestMusicDirectory_GetUsageInstructions tests the GetUsageInstructions method
func TestMusicDirectory_GetUsageInstructions(t *testing.T) {
	// Use custom directory
	md := files.MusicDirectory("custom_dir")

	// Execute test
	instructions := md.GetUsageInstructions()

	// Check if custom directory name is included
	if !strings.Contains(instructions, "custom_dir") {
		t.Errorf("MusicDirectory.GetUsageInstructions() should contain the music directory name")
	}

	// Check if expected phrases are included
	expectedPhrases := []string{
		"No music files found",
		"Place .wav, .ogg, or .mp3 files",
		"Restart the application",
		"Space: Toggle pause",
		"N: Skip to next track",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(instructions, phrase) {
			t.Errorf("MusicDirectory.GetUsageInstructions() expected to contain %q but did not", phrase)
		}
	}
}

// TestGetHowToUseMessage tests the GetHowToUseMessage function
func TestGetHowToUseMessage(t *testing.T) {
	message := files.GetHowToUseMessage()

	// Check if expected phrases are included
	expectedPhrases := []string{
		"Warning: Music files are needed",
		"Please place WAV, OGG, or MP3 files in the musics directory",
		"song1.wav",
		"song2.mp3",
		"song3.ogg",
	}

	for _, phrase := range expectedPhrases {
		if !strings.Contains(message, phrase) {
			t.Errorf("GetHowToUseMessage() expected to contain %q but did not", phrase)
		}
	}
}

// TestDefaultFunctions tests the functions that use DefaultMusicDir
func TestDefaultFunctions(t *testing.T) {
	// Test for FindMusicFiles function
	t.Run("FindMusicFiles", func(t *testing.T) {
		// Check if FindMusicFiles function executes normally
		_, err := files.FindMusicFiles()
		if err != nil {
			t.Errorf("FindMusicFiles() error = %v", err)
		}

		testDataDir := files.MusicDirectory("testdata")
		testDataFiles, err := testDataDir.FindMusicFiles()
		if err != nil {
			t.Errorf("MusicDirectory(testdata).FindMusicFiles() error = %v", err)
		}

		// Since the default 'musics' directory might not have actual files,
		// only verify the number of files in the testdata directory
		if len(testDataFiles) == 0 {
			t.Errorf("Test data directory has no files, test may be inaccurate")
		}
	})

	// Test for EnsureMusicDirectory function
	t.Run("EnsureMusicDirectory", func(t *testing.T) {
		// Check if function executes normally
		musicDir, err := files.EnsureMusicDirectory()
		if err != nil {
			t.Errorf("EnsureMusicDirectory() error = %v", err)
		}

		// Check if returned value is an absolute path
		if !filepath.IsAbs(musicDir) {
			t.Errorf("EnsureMusicDirectory() returned a relative path: %s", musicDir)
		}

		// Check if default directory exists
		defaultPath := string(files.DefaultMusicDir)
		if _, err := os.Stat(defaultPath); os.IsNotExist(err) {
			// When the test runs, the default directory should be created automatically
			t.Errorf("EnsureMusicDirectory() did not create the default directory")
		} else {
			// No need to delete the directory after the test, it's fine to leave it
		}
	})

	// Test for GetUsageInstructions function
	t.Run("GetUsageInstructions", func(t *testing.T) {
		instructions := files.GetUsageInstructions()
		defaultInstructions := files.DefaultMusicDir.GetUsageInstructions()

		// Check if both results are the same
		if instructions != defaultInstructions {
			t.Errorf("GetUsageInstructions() returned a different result from DefaultMusicDir.GetUsageInstructions()")
		}

		// Check if default directory name is included
		if !strings.Contains(instructions, string(files.DefaultMusicDir)) {
			t.Errorf("GetUsageInstructions() should contain the default music directory name")
		}
	})
}
