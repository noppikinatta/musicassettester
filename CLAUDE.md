# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Music Asset Tester is a GUI application for testing game loop audio files. It automatically discovers, plays, and cycles through audio files with configurable fade-out and interval timing. Built with Go using Ebitengine and GuiGui for the UI.

## Build and Run Commands

```bash
# Run the application
go run main.go

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run a single test
go test -run TestName ./internal/player/

# Format code
gofmt -w .
```

## Architecture

### Package Structure

- `main.go` - Application bootstrap, audio context initialization, game loop setup
- `internal/files/` - File discovery and directory watching (fsnotify-based hot-reload)
- `internal/player/` - Audio playback state machine, music selection, format decoding
- `internal/ui/` - GuiGui-based UI with custom widgets (sliders, progress bars)

### Key Components

**MusicPlayer** (`internal/player/player.go`): Core orchestrator with state machine:
- `StatePlaying` → plays track, increments counter
- `StateFadingOut` → decreases volume over 2 seconds
- `StateInterval` → silence between tracks
- `StateStopped` → no playback

**MusicSelector** (`internal/player/player.go`): Manages file list and track selection (sequential or random mode).

**DirectoryWatcher** (`internal/files/files.go`): Monitors music directory for changes with 500ms debounce.

**Root** (`internal/ui/root.go`): Main UI widget containing music list, playback controls, and settings sliders.

### Audio Pipeline

```
File → MusicLoader.LoadStream() → Decoder (WAV/OGG/MP3) → audio.NewInfiniteLoop() → Player
```

All audio is resampled to 48000 Hz.

### Keyboard Shortcuts

- Space: Toggle pause
- N: Skip to next track
- R: Toggle random mode

## Testing Patterns

- Use table-driven tests
- Create unique mocks per test (don't share mock implementations)
- Test files use `_test` package suffix (e.g., `player_test`)
- Test data goes in `testdata/` directory
- Use testify for assertions

## Development Rules (from Cursor rules)

- Follow TDD (red-green-refactor cycle)
- Check for duplicate implementations before coding
- Comments and documentation in English
- Use `gofmt` for formatting
- Do not auto-commit; propose commit messages instead

### Commit Message Format

Use emoji prefix based on change type:
- ✨ New features
- 🔧 Modified existing features
- ♻️ Refactoring
- 🩹 Bug fixes
- 🧪 Test code changes
- 📝 Documentation changes

## Supported Audio Formats

- WAV, OGG, MP3
- Fixed 48000 Hz sample rate

## Dependencies

- Ebitengine v2.9.0-alpha - Graphics, input, audio
- GuiGui - Widget library
- fsnotify - File system watching
- testify - Test assertions
