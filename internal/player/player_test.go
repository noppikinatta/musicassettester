package player_test

import (
	"musicplayer/internal/player"
	"os"
	"path/filepath"
	"testing"
)

// PlayerStateのエイリアスを定義して、テストで定数を使いやすくする
const (
	StateStopped      = player.PlayerState(0)
	StatePlaying      = player.PlayerState(1)
	StateFadingOut    = player.PlayerState(2)
	StateInterval     = player.PlayerState(3)
	StateInitializing = player.PlayerState(0) // StateStoppedと同じ値
)

// TestMainはテスト全体のセットアップを行います
func TestMain(m *testing.M) {
	// テストのセットアップを行う
	// 追加のセットアップが必要な場合はここに記述

	// テストを実行
	code := m.Run()

	// テストの後処理を行う
	// 追加の後処理が必要な場合はここに記述

	// プロセスを終了
	os.Exit(code)
}

// テスト用のMusicPlayerを作成するヘルパー関数
func createTestMusicPlayer(t *testing.T) (*player.MusicPlayer, *MockPlayerFactory) {
	// テスト用の一時ディレクトリを作成
	tempDir, err := os.MkdirTemp("", "music-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(tempDir) })

	// モックディレクトリとファクトリを作成
	mockDir := NewMockMusicDirectory(tempDir, []string{}, true)
	mockFactory := NewMockPlayerFactory()

	// プレイヤーを作成
	p, err := player.NewMusicPlayer(mockDir, mockFactory)
	if err != nil {
		t.Logf("Warning during player creation: %v", err)
	}

	// テスト用に音楽ファイルを追加 (実際にはファイルは作成しない)
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
	if state != StateStopped { // 初期状態はStoppedになる
		t.Errorf("Expected state %d, got %d", StateStopped, state)
	}

	// ファクトリを通じてPlayerが作成されたことを確認
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
	// 最初は空か、テスト中に設定された値
	t.Logf("Current path: %s", path)
}

func TestGetState(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	state := p.GetState()
	if state != StateStopped {
		t.Errorf("Expected initial state to be StateStopped (%d), got %d", StateStopped, state)
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
	// 初期カウンターは0を期待
	if counter != 0 {
		t.Errorf("Expected initial counter to be 0, got %d", counter)
	}
}

func TestLoopDurationMinutes(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	// デフォルト値の確認
	duration := p.GetLoopDurationMinutes()
	if duration != 5.0 {
		t.Errorf("Expected default loop duration to be 5.0 minutes, got %f", duration)
	}

	// 値を変更して確認
	p.SetLoopDurationMinutes(10.0)
	duration = p.GetLoopDurationMinutes()
	if duration != 10.0 {
		t.Errorf("Expected loop duration to be 10.0 minutes after setting, got %f", duration)
	}
}

func TestIntervalSeconds(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	// デフォルト値の確認
	interval := p.GetIntervalSeconds()
	if interval != 10.0 {
		t.Errorf("Expected default interval to be 10.0 seconds, got %f", interval)
	}

	// 値を変更して確認
	p.SetIntervalSeconds(15.0)
	interval = p.GetIntervalSeconds()
	if interval != 15.0 {
		t.Errorf("Expected interval to be 15.0 seconds after setting, got %f", interval)
	}
}

func TestSetCurrentIndex(t *testing.T) {
	p, _ := createTestMusicPlayer(t)

	// 有効なインデックスを設定
	if len(p.GetMusicFiles()) > 0 {
		err := p.SetCurrentIndex(0)
		if err != nil {
			t.Errorf("Expected SetCurrentIndex(0) to succeed, got error: %v", err)
		}

		// 範囲外のインデックスを設定
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

	// Playerインスタンスが必要なので、現在のPlayerをnilに設定し、新しいモックプレイヤーを作成します
	mockPlayer := mockFactory.GetLastPlayer()

	// MusicPlayerの内部状態を直接設定（テスト用）
	p.TestSetPlayer(mockPlayer)
	p.TestSetPaused(false)

	// 最初は一時停止していない状態
	if p.IsPaused() {
		t.Error("Test setup failed: player should not be paused initially")
	}

	// 一時停止
	p.TogglePause()
	if !p.IsPaused() {
		t.Error("Expected player to be paused after TogglePause")
	}

	// 再開
	p.TogglePause()
	if p.IsPaused() {
		t.Error("Expected player to not be paused after second TogglePause")
	}
}

func TestUpdate(t *testing.T) {
	p, mockFactory := createTestMusicPlayer(t)

	// MusicPlayerの内部状態を直接設定
	mockPlayer := mockFactory.GetLastPlayer()
	p.TestSetPlayer(mockPlayer)
	p.TestSetCounter(0)
	p.TestSetPaused(false)
	p.TestSetState(StatePlaying)

	// 初期カウンターを確認
	if p.GetCounter() != 0 {
		t.Errorf("Test setup failed: expected initial counter to be 0, got %d", p.GetCounter())
	}

	// Player存在・非一時停止状態でのUpdate - カウンターが増加する
	err := p.Update()
	if err != nil {
		t.Errorf("Expected Update() to succeed, got error: %v", err)
	}

	if p.GetCounter() != 1 {
		t.Errorf("Expected counter to be 1 after update, got %d", p.GetCounter())
	}

	// 一時停止中はカウンターが更新されない
	p.TestSetPaused(true)
	initialCounter := p.GetCounter() // 現在のカウンター値を保存

	err = p.Update()
	if err != nil {
		t.Errorf("Expected Update() during pause to succeed, got error: %v", err)
	}

	if p.GetCounter() != initialCounter {
		t.Errorf("Expected counter to remain %d during pause, got %d",
			initialCounter, p.GetCounter())
	}

	// Playerがnilの場合もカウンターは更新されない
	p.TestSetPaused(false) // 一時停止解除
	p.TestSetPlayer(nil)   // Playerをnilに設定
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
