package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
)

// Spotify API endpoint URLs
const (
	SPOTIFY_API_URL       = "https://api.spotify.com/v1"
	SPOTIFY_SEARCH_URL    = SPOTIFY_API_URL + "/search"
	SPOTIFY_PLAYLISTS_URL = SPOTIFY_API_URL + "/users/%s/playlists"
)

var State = "playlister-" + time.Now().String()

func createAuthClient() (*spotify.Client, error) {
	auth := spotifyauth.New(
		spotifyauth.WithClientID(os.Getenv("SPOTIFY_CLIENT_ID")),
		spotifyauth.WithClientSecret(os.Getenv("SPOTIFY_CLIENT_SECRET")),
		spotifyauth.WithRedirectURL("http://localhost:8008/callback"),
		spotifyauth.WithScopes(spotifyauth.ScopePlaylistModifyPrivate),
	)
	authURL := auth.AuthURL(State)

	fmt.Printf("Please log in to Spotify by visiting the following page in your web browser:\n\n%s\n\n", authURL)
	err := exec.Command("open", "-a", "Google Chrome", authURL).Run()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	// start a web server to listen for the callback request
	authChan := make(chan *spotify.Client)
	errChan := make(chan error)
	go func() {
		http.HandleFunc("/callback", func(_ http.ResponseWriter, r *http.Request) {
			// get the authorization code from the query parameters
			code := r.URL.Query().Get("code")

			// exchange the authorization code for an access token
			token, err := auth.Exchange(r.Context(), code)
			if err != nil {
				errChan <- fmt.Errorf("failed to exchange authorization code: %v", err)
				return
			}

			// create a Spotify client using the access token
			client := spotify.New(auth.Client(r.Context(), token))
			authChan <- client
		})
		if err := http.ListenAndServe(":8008", nil); err != nil {
			errChan <- fmt.Errorf("failed to start HTTP server: %v", err)
		}
	}()

	// wait for the authentication process to complete
	select {
	case client := <-authChan:
		return client, nil
	case err := <-errChan:
		return nil, err
	}
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	csvFilePath := flag.String("csv", "", "path to CSV file")
	flag.Parse()

	if *csvFilePath == "" {
		fmt.Println("Please provide a path to a CSV file using the -csv flag")
		return
	}

	file, err := os.Open(*csvFilePath)
	if err != nil {
		fmt.Printf("Failed to open CSV file: %v\n", err)
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Printf("Failed to read CSV file: %v\n", err)
		return
	}

	// create a Spotify client using client credentials flow
	client, err := createAuthClient()
	if err != nil {
		fmt.Printf("Failed to create Spotify client: %v\n", err)
		return
	}

	ctx := context.Background()
	user, err := client.CurrentUser(ctx)
	if err != nil {
		fmt.Printf("failed to get current user: %v", err)
	}

	// create a new playlist with the name of the CSV file
	playlistName := strings.TrimSuffix(filepath.Base(*csvFilePath), filepath.Ext(*csvFilePath))
	playlist, err := client.CreatePlaylistForUser(ctx, user.ID, playlistName, "", false, false)
	if err != nil {
		fmt.Printf("Failed to create playlist '%s': %v\n", playlistName, err)
		return
	}

	// add each track from the CSV file to the playlist
	for _, record := range records[1:] {
		// search for the track using the artist and track name fields
		query := fmt.Sprintf("%s %s", record[1], record[0])

		searchResp, err := client.Search(ctx, query, spotify.SearchTypeTrack)
		if err != nil {
			fmt.Printf("Failed to search for track '%s': %v\n", record[1], err)
			continue
		}

		// throw in slight delay to avoid rate limiting
		time.Sleep(100 * time.Millisecond)

		// add the first search result to the playlist
		if len(searchResp.Tracks.Tracks) > 0 {
			trackID := searchResp.Tracks.Tracks[0].ID

			_, err := client.AddTracksToPlaylist(ctx, playlist.ID, trackID)
			if err != nil {
				fmt.Printf("Failed to add track '%s - %s' to playlist: %v\n", record[1], record[0], err)
				continue
			}
			fmt.Printf("Added track '%s - %s' to playlist\n", record[1], record[0])
		} else {
			fmt.Printf("Could not find track '%s- %s'\n", record[1], record[0])
		}
	}
}
