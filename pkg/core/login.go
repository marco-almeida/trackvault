package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/marco-almeida/trackvault/pkg/music"
	"github.com/marco-almeida/trackvault/pkg/music/spotify"
)

type LoginArgs struct {
	Provider string
}

const providerNameSpotify = "spotify"

func Login(ctx context.Context, args LoginArgs) (music.Provider, error) {
	switch strings.ToLower(args.Provider) {
	case providerNameSpotify:
		// TODO: check if already logged in.
		spotifyClient := spotify.NewSpotifyClient()
		err := spotifyClient.Login(ctx, music.LoginArgs{})
		if err != nil {
			return nil, fmt.Errorf("error logging in to Spotify: %v", err)
		}
		return spotifyClient, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", args.Provider)
	}
}
