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

const ProviderNameSpotify = "spotify"

func Login(ctx context.Context, args LoginArgs) (music.Provider, error) {
	var musicProvider music.Provider
	switch strings.ToLower(args.Provider) {
	case ProviderNameSpotify:
		musicProvider = spotify.NewSpotifyClient()
	default:
		return nil, fmt.Errorf("unsupported provider: %s", args.Provider)
	}

	err := musicProvider.Login(ctx, music.LoginArgs{})
	if err != nil {
		return nil, fmt.Errorf("error logging in to %s: %w", args.Provider, err)
	}
	user, err := musicProvider.User(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user info from %s: %w", args.Provider, err)
	}
	fmt.Printf("Logged in to %s as %s (%s)\n", args.Provider, user.DisplayName, user.ID)
	return musicProvider, nil
}
