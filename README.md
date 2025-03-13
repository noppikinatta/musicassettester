# Music Asset Tester

A game music asset tester application using Ebitengine.

## Features

- Automatically searches and plays WAV files from the musics directory in sequence
- Automatic playback system:
  - Loops a single track for 5 minutes
  - Fades out volume over 5 seconds
  - After a 10-second interval, selects the next track in sequence
- Simple UI: Displays currently playing file path and status

## Specifications

1. At startup, searches through the musics directory including subfolders and lists all *.wav file paths
2. Plays music tracks sequentially from the found paths
3. Fades out after 5 minutes over a 5-second period
4. After a 10-second interval, selects the next track
5. Fixed sample rate of 44100

## Limitations

- Currently only supports WAV file format. MP3 and OGG formats are not supported.
- Fixed sample rate of 44100Hz. Other sample rates are not supported.
- If you need support for additional formats or sample rates, please fork this repository and make the necessary modifications.

## How to Run

```
go run main.go
```

## Dependencies

- [Ebitengine](https://ebiten.org/) - v2.6.6
- Go 1.23.4 or higher

## Directory Structure

- `main.go` - Main application code
- `musics/` - Directory to store WAV files

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

noppikinatta
