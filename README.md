# Playlister

Playlister is a command-line tool for creating Spotify playlists from a CSV file.
I made this for personal use to curate music, but feel free to use for whatever.

## Requirements

- Go 1.16 or later
- A Spotify account
- A [Spotify Developer Dashboard](https://developer.spotify.com/dashboard) account and an app with a client ID and secret

## Usage

```bash
go run main.go -csv /path/to/playlist.csv
```

Or build and run the binary, use the following commands:

```bash
go build
./playlister -csv /path/to/playlist.csv
```

The CSV file should have the following format:

```csv
artist,track

Led Zeppelin,Stairway to Heaven
Pink Floyd,Wish You Were Here
The Beatles,Hey Jude
```

## Authentication

Before running Playlister, you need to set up a Spotify app and set the `SPOTIFY_CLIENT_ID` and `SPOTIFY_CLIENT_SECRET` environment variables with the app credentials.

To authenticate with Spotify, Playlister will start a web server at http://localhost:8008/callback to receive the access token.

## Limitations

- Amongst other things, Playlister currently does not support adding tracks to an existing playlist. If the playlist already exists, Playlister will create a new one with the same name.

## Feature requests

Kindly create an [issue](https://github.com/aesrael/playlister/issues), and I may be able to attend to a few.

## License

Playlister is licensed under the [MIT License](LICENSE).
