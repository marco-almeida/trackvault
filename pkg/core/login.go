package core

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/zalando/go-keyring"
	"golang.org/x/oauth2"

	"github.com/marco-almeida/trackvault/pkg/music"
	"github.com/marco-almeida/trackvault/pkg/music/spotify"
)

type LoginArgs struct {
	Provider string
}

const ProviderNameSpotify = "spotify"

func Login(ctx context.Context, args LoginArgs) (music.Provider, error) {
	var musicProvider music.Provider
	var err error
	var token oauth2.Token
	switch strings.ToLower(args.Provider) {
	case ProviderNameSpotify:
		musicProvider, token, err = spotify.NewSpotifyClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("error creating spotify client: %w", err)
		}
		// store token in keyring
		tokenJson, err := json.Marshal(token)
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not marshal token to json:", err)
		}

		// store oauth token in OS
		err = keyring.Set("trackvault", "spotify", string(tokenJson))
		if err != nil {
			fmt.Fprintln(os.Stderr, "could not store refresh token in keyring:", err)
		}
	default:
		return nil, fmt.Errorf("unsupported provider: %s", args.Provider)
	}

	user, err := musicProvider.User(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting user info from %s: %w", args.Provider, err)
	}
	fmt.Printf("Logged in to %s as %s (%s)\n", args.Provider, user.DisplayName, user.ID)
	return musicProvider, nil
}
