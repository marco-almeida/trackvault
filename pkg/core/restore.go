package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/schollz/progressbar/v3"
	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"

	"github.com/marco-almeida/trackvault/pkg/music"
	"github.com/marco-almeida/trackvault/pkg/music/spotify"
)

type RestoreArgs struct {
	Provider   string
	BackupPath string
}

func Restore(ctx context.Context, args RestoreArgs) error {
	var musicProvider music.Provider
	var err error
	switch strings.ToLower(args.Provider) {
	case ProviderNameSpotify:
		oauthToken, err := getOAuthTokenFromKeyring(args.Provider)
		if err != nil {
			return fmt.Errorf("could not get oauth token from keyring: %w", err)
		}
		musicProvider, err = getSpotifyClientFromToken(ctx, oauthToken)
		if err != nil {
			return fmt.Errorf("could not create spotify client: %w", err)
		}
	default:
		return fmt.Errorf("unsupported provider: %s", args.Provider)
	}

	// get backup path or default to data dir,latest directory
	finalDestinationPath := args.BackupPath
	if finalDestinationPath == "" {
		finalDestinationPath = DefaultBackupDir()
		entries, err := os.ReadDir(finalDestinationPath)
		if err != nil {
			return fmt.Errorf("could not read backup directory: %w", err)
		}

		finalDestinationPath = filepath.Join(finalDestinationPath, entries[len(entries)-1].Name())
	}

	// backup folder must exist and must have the backups
	_, err = os.Stat(finalDestinationPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("backup directory %s does not exist", finalDestinationPath)
		}

		return fmt.Errorf("could not stat backup directory: %w", err)
	}

	entries, err := os.ReadDir(finalDestinationPath)
	if err != nil {
		return fmt.Errorf("could not read backup directory: %w", err)
	}

	var backupMetadata BackupMetadata
	playlistsWithTracksBackup := make([]PlaylistWithTracksBackup, 0)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		if e.Name() == "metadata.json" {
			backupMetadataBytes, err := os.ReadFile(filepath.Join(finalDestinationPath, e.Name()))
			if err != nil {
				return fmt.Errorf("could not read metadata file: %w", err)
			}

			err = json.Unmarshal(backupMetadataBytes, &backupMetadata)
			if err != nil {
				return fmt.Errorf("could not unmarshal metadata file: %w", err)
			}
		} else {
			var playlistWithTracksBackup PlaylistWithTracksBackup
			playlistWithTracksBackupBytes, err := os.ReadFile(filepath.Join(finalDestinationPath, e.Name()))
			if err != nil {
				return fmt.Errorf("could not read playlist file: %w", err)
			}
			err = json.Unmarshal(playlistWithTracksBackupBytes, &playlistWithTracksBackup)
			if err != nil {
				return fmt.Errorf("could not unmarshal playlist file: %w", err)
			}
			playlistsWithTracksBackup = append(playlistsWithTracksBackup, playlistWithTracksBackup)
		}
	}

	fmt.Printf("Found %d playlists backed up at %s from provider %s\n", len(playlistsWithTracksBackup), backupMetadata.Timestamp, backupMetadata.Provider)

	var wg sync.WaitGroup

	bar := progressbar.Default(int64(len(playlistsWithTracksBackup)), "Restoring Playlists")

	for _, playlist := range playlistsWithTracksBackup {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer bar.Add(1) //nolint:errcheck
			processErr := processPlaylistRestore(ctx, musicProvider, playlist)
			if processErr != nil {
				fmt.Fprintf(os.Stderr, "could not process playlist %s: %s\n", playlist.Playlist.Name, processErr)
			}
		}()
	}

	wg.Wait()

	return err
}

func processPlaylistRestore(ctx context.Context, musicProvider music.Provider, playlist PlaylistWithTracksBackup) error {
	// create playlist
	createdPlaylist, err := musicProvider.CreatePlaylist(ctx, music.CreatePlaylistArgs{
		PlaylistDetails: playlist.Playlist,
	})
	if err != nil {
		if strings.Contains(err.Error(), "Error while loading resource") {
			time.Sleep(150 * time.Millisecond)
			// retry once
			createdPlaylist, err = musicProvider.CreatePlaylist(ctx, music.CreatePlaylistArgs{
				PlaylistDetails: playlist.Playlist,
			})
		} else {
			return fmt.Errorf("could not create playlist: %w", err)
		}
	}

	if err != nil {
		if strings.Contains(err.Error(), "Error while loading resource") {
			time.Sleep(150 * time.Millisecond)
			// retry once
			createdPlaylist, err = musicProvider.CreatePlaylist(ctx, music.CreatePlaylistArgs{
				PlaylistDetails: playlist.Playlist,
			})
		} else {
			return fmt.Errorf("could not create playlist: %w", err)
		}
	}

	if err != nil {
		return fmt.Errorf("could not create playlist: %w", err)
	}

	time.Sleep(150 * time.Millisecond)

	// add tracks to playlist
	_, err = musicProvider.AddTracksToPlaylist(ctx, music.AddTracksToPlaylistArgs{
		Playlist: createdPlaylist,
		Tracks:   playlist.Tracks,
	})
	if err != nil {
		return fmt.Errorf("could not add track to playlist: %w", err)
	}

	return nil
}

func getSpotifyClientFromToken(ctx context.Context, oauthToken oauth2.Token) (music.Provider, error) {
	musicProvider, newtoken, err := spotify.NewSpotifyClientFromToken(ctx, oauthToken)
	if err != nil {
		return nil, fmt.Errorf("could not create spotify client: %w", err)
	}

	if newtoken != nil && newtoken.ExpiresIn != oauthToken.ExpiresIn {
		// store newtoken in keyring
		tokenJson, err := json.Marshal(newtoken)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not marshal token to json:", err)
		}

		// store oauth newtoken in OS
		err = keyring.Set("trackvault", "spotify", string(tokenJson))
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not store refresh token in keyring:", err)
		}
	}
	return musicProvider, nil
}
