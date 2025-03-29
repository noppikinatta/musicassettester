# 開発計画

## 依存関係と責務の整理

複雑に絡み合った依存関係と責務を整理します。

### files.DirectoryWatcherのcallbackを複数持つことができるようにする

files.DirectoryWatcherはcallbackを１つだけ持つことができますが、変更します。
まず、FileChangeCallback型をFileChangeHandlerに変更します。
そして、files.DirectoryWatcher.callbackを`handlers []FileChangeHandler`に変更します。
そして、files.DirectoryWatcher.AddHandlerメソッドを追加して、handlersに関数を追加できるようにします。

### player.MusicPlayerは、filesのことを知らない方がいい

player.MusicPlayerはfiles.DirectoryWatcherを直接所有するべきではありません。
files.MusicDirectory.Watchでwatcherの作成とイベントハンドラの登録を同時に行うのではなく、
前述のwatcher.AddHandlerによってイベントハンドラを追加できるようにすべきです。
player.MusicPlayerのhandleFileChangesをパブリックのAPI(UpdateMusicFilesなど)に変更し、main.goの中で関連付けを行います。
player.MusicPlayerはfiles.DirectoryWatcherを所有しないようになり、依存関係が単純化します。

### player.MusicPlayerの責務を分割する

player.MusicPlayerは多くのことを行っているので、型を分割すべきです。

#### player.MusicSelectorの追加

楽曲ファイルを選択するMusicSelector構造体を追加します。

```
type MusicSelector struct {
	musicFiles    []string
	currentIndex  int
}
```

を追加します。この型は、音楽ファイルパスのリストと、現在のインデックスを持ちます。

```
func (s *MusicSelector) Update(musicFiles []string) {

}
```

のAPIを定義し、musicFilesの更新とcurrentIndexの更新を行います。
まず既存のcurrentIndexに該当するファイルパスを保持しておき、新しいmusicFilesにそのファイルが含まれていれば、currentIndexを新しいmusicFilesのそのファイルのインデックスに更新します。

また、CurrentFile() (string, bool) のAPIを持ちます。CurrentFileはcurrentIndexがmusicFilesの添え字の範囲外のときにfalseを返します。これは、フォルダ内に楽曲が１つも存在しない場合に便利です。

#### player.Musicの追加

Playerに機能を追加するMusic構造体を追加します。

```
type Music struct {
    Player
}
```

これは、Playerインタフェースには定義されていない機能を提供するものです。

#### player.MusicPlayerの簡素化

追加した２つの型を使うことで、MusicPlayerの機能は簡素化できます。

### ui.Rootの責務を分割する

### ui.Rootのレイアウトを見やすくする

### 楽曲の真の長さを取得する

time.Second * time.Duration(s.Length()) / bytesPerSample / sampleRate