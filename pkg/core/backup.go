package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"

	"github.com/marco-almeida/trackvault/pkg/music"
	"github.com/marco-almeida/trackvault/pkg/music/spotify"
)

type BackupArgs struct {
	Provider        string
	DestinationPath string
}

type BackupMetadata struct {
	Provider  string    `json:"provider"`
	Timestamp time.Time `json:"timestamp"`
	Result    string    `json:"status"`
}

type PlaylistWithTracksBackup struct {
	Playlist      music.Playlist `json:"playlist"`
	BackupSuccess bool           `json:"backup_success"`
	Tracks        []music.Track  `json:"tracks"`
}

func Backup(ctx context.Context, args BackupArgs) error {
	var musicProvider music.Provider
	switch strings.ToLower(args.Provider) {
	case ProviderNameSpotify:
		oauthToken, err := getOAuthTokenFromKeyring(args.Provider)
		if err != nil {
			return fmt.Errorf("could not get oauth token from keyring: %w", err)
		}
		fmt.Println("OAuth Token ", oauthToken)
		musicProvider = spotify.NewSpotifyClientFromToken(ctx, oauthToken)
	default:
		return fmt.Errorf("unsupported provider: %s", args.Provider)
	}
	playlists, err := musicProvider.ListUserPlaylists(ctx, music.ListUserPlaylistsArgs{})
	if err != nil {
		return fmt.Errorf("could not list user playlists: %w", err)
	}

	// add liked songs as a virtual playlist
	playlists = append(playlists, music.Playlist{
		ID:        fmt.Sprintf("%s-collection-liked", args.Provider),
		Name:      "Liked Songs",
		IsVirtual: true,
	})

	fmt.Printf("Found %d playlists (including liked songs)\n", len(playlists))

	finalDestinationPath := args.DestinationPath
	if finalDestinationPath == "" {
		finalDestinationPath = DefaultBackupDir()
	}

	datetimeNow := time.Now().UTC()
	datetimeNowString := datetimeNow.Format("2006-01-02T15-04-05.000Z")
	finalDestinationPath = filepath.Join(finalDestinationPath, datetimeNowString)
	finalDestinationPath = filepath.Clean(finalDestinationPath)

	err = os.MkdirAll(finalDestinationPath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create backup directory: %w", err)
	}

	var wg sync.WaitGroup
	var errors atomic.Int64
	errors.Store(0)

	bar := progressbar.Default(int64(len(playlists)), "Backing up playlists")

	for _, playlist := range playlists {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer bar.Add(1) //nolint:errcheck
			processErr := processPlaylistBackup(ctx, musicProvider, playlist, BackupArgs{
				Provider:        args.Provider,
				DestinationPath: finalDestinationPath,
			})
			if processErr != nil {
				fmt.Fprintf(os.Stderr, "could not process playlist %s: %s\n", playlist.Name, processErr)
				errors.Add(1)
			}
		}()
	}

	wg.Wait()

	resultStatus := "success"
	if errors.Load() > 0 {
		resultStatus = "partial_failure"
	}
	if errors.Load() == int64(len(playlists)) {
		resultStatus = "failure"
	}

	// write metadata to folder
	metadata := BackupMetadata{
		Provider:  args.Provider,
		Timestamp: datetimeNow,
		Result:    resultStatus,
	}
	metadataJSON, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal metadata to json: %w", err)
	}
	metadataPath := filepath.Join(finalDestinationPath, "metadata.json")
	err = os.WriteFile(metadataPath, metadataJSON, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not write metadata to file: %w", err)
	}

	fmt.Println("Backup folder created at:", finalDestinationPath)

	return err
}

func processPlaylistBackup(ctx context.Context, musicProvider music.Provider, playlist music.Playlist, backupArgs BackupArgs) error {
	// handle liked songs virtual playlist
	var tracks []music.Track
	var err error
	if playlist.ID == fmt.Sprintf("%s-collection-liked", backupArgs.Provider) {
		tracks, err = musicProvider.ListSavedTracks(ctx, music.ListSavedTracksArgs{})
	} else {
		tracks, err = musicProvider.ListTracksInPlaylist(ctx, playlist)
	}

	if err != nil {
		return fmt.Errorf("could not list tracks in playlist %s: %w", playlist.Name, err)
	}

	playlistWithTracks := PlaylistWithTracksBackup{
		Playlist:      playlist,
		BackupSuccess: err == nil,
		Tracks:        tracks,
	}

	tracksInLastPlaylistBackupPath := filepath.Join(backupArgs.DestinationPath, fmt.Sprintf("%s_%s.json", playlist.Name, playlist.ID))
	tracksJSON, err := json.MarshalIndent(playlistWithTracks, "", "  ")
	if err != nil {
		return fmt.Errorf("could not marshal tracks to json: %w", err)
	}
	err = os.WriteFile(tracksInLastPlaylistBackupPath, tracksJSON, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not write tracks to file: %w", err)
	}
	return nil
}

func getOAuthTokenFromKeyring(provider string) (oauth2.Token, error) {
	provider = strings.ToLower(provider)
	oauthTokenJSON, err := keyring.Get("trackvault", provider)
	if err != nil {
		return oauth2.Token{}, fmt.Errorf("could not get oauth token from keyring: %w", err)
	}
	fmt.Println(oauthTokenJSON)
	var oauthToken oauth2.Token
	err = json.Unmarshal([]byte(oauthTokenJSON), &oauthToken)
	if err != nil {
		return oauth2.Token{}, fmt.Errorf("could not unmarshal oauth token: %w", err)
	}
	return oauthToken, nil
}
