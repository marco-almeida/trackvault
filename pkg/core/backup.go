package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"

	"github.com/marco-almeida/trackvault/pkg/music"
	"github.com/marco-almeida/trackvault/pkg/music/spotify"
)

type BackupArgs struct {
	Provider        string
	DestinationPath string
}

func ListPlaylistsAndLikes(ctx context.Context, args BackupArgs) ([]music.Playlist, error) {
	var musicProvider music.Provider
	switch strings.ToLower(args.Provider) {
	case ProviderNameSpotify:
		oauthToken, err := getOAuthTokenFromKeyring(args.Provider)
		if err != nil {
			return nil, fmt.Errorf("could not get oauth token from keyring: %w", err)
		}
		fmt.Println("OAuth Token ", oauthToken)
		musicProvider = spotify.NewSpotifyClientFromToken(ctx, oauthToken)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", args.Provider)
	}
	playlists, err := musicProvider.ListPlaylistsAndLikes(ctx, music.ListPlaylistsAndLikesArgs{})

	return playlists, err
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
