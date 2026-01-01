package spotify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/grokify/go-pkce"
	"github.com/zalando/go-keyring"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"

	"github.com/marco-almeida/trackvault/pkg/music"
	utilsURL "github.com/marco-almeida/trackvault/pkg/utils"
)

const (
	redirectHost = "127.0.0.1"
	redirectPort = "43721"
	clientID     = "a0852cfbdbd24dcba2410c011ab29564"
)

var (
	scopes = []string{
		spotifyauth.ScopeUserReadPrivate,
		spotifyauth.ScopePlaylistReadPrivate,
		spotifyauth.ScopePlaylistReadCollaborative,
		spotifyauth.ScopeUserLibraryRead,
		spotifyauth.ScopePlaylistModifyPublic,
		spotifyauth.ScopePlaylistModifyPrivate,
	}
	redirectURI     = fmt.Sprintf("http://%s:%s/spotify/callback", redirectHost, redirectPort)
	auth            = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(scopes...), spotifyauth.WithClientID(clientID))
	ch              = make(chan *oauth2.Token)
	state           = uuid.New().String()
	codeVerifier, _ = pkce.NewCodeVerifier(-1)
	codeChallenge   = pkce.CodeChallengeS256(codeVerifier)
)

type SpotifyClient struct {
	client *spotify.Client
	auth   *spotifyauth.Authenticator
}

// NewSpotifyClient creates an empty SpotifyClient, it needs to be logged in later
func NewSpotifyClient() music.Provider {
	return &SpotifyClient{}
}

// NewSpotifyClientFromToken creates a SpotifyClient from a refresh token
func NewSpotifyClientFromToken(ctx context.Context, token oauth2.Token) music.Provider {
	newtoken, err := auth.RefreshToken(ctx, &token)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not refresh token: %v\n", err)
	}

	if newtoken != nil && newtoken.ExpiresIn != token.ExpiresIn {
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
	httpClient := auth.Client(ctx, &token)
	spotifyClient := spotify.New(httpClient, spotify.WithRetry(true))
	return &SpotifyClient{client: spotifyClient, auth: auth}
}

func (s *SpotifyClient) Login(ctx context.Context, args music.LoginArgs) error { // TODO: this is weird, should probably make this the newClient
	completeAuth := func(w http.ResponseWriter, r *http.Request) {
		tok, err := auth.Token(r.Context(), state, r,
			oauth2.SetAuthURLParam("code_verifier", codeVerifier))
		if err != nil {
			http.Error(w, "couldn't get token", http.StatusForbidden)
			fmt.Fprintln(os.Stderr, err)
			ch <- nil
			return
		}
		if st := r.FormValue("state"); st != state {
			http.NotFound(w, r)
			fmt.Fprintf(os.Stderr, "State mismatch: %s != %s\n", st, state)
			ch <- nil
			return
		}
		// use the token to get an authenticated client
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprintf(w, `
        <html>
            <body>
                <h2>Login completed!</h2>
                <p>You can now return to the terminal.</p>
                <script>window.close()</script>
            </body>
        </html>
    `)

		ch <- tok
	}

	http.HandleFunc("/spotify/callback", completeAuth)
	go http.ListenAndServe(fmt.Sprintf("%s:%s", redirectHost, redirectPort), nil) //nolint:errcheck

	url := auth.AuthURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", pkce.MethodS256),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("client_id", clientID),
	)
	fmt.Println("Log in to Spotify by visiting the following URL in your browser:", url)
	err := utilsURL.OpenURL(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not open URL in browser:", err)
	}
	// wait for auth to complete
	token := <-ch
	if token == nil {
		return fmt.Errorf("could not get token from callback")
	}
	s.client = spotify.New(auth.Client(ctx, token), spotify.WithRetry(true))
	s.auth = auth

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

	return nil
}

func (s *SpotifyClient) User(ctx context.Context) (*music.User, error) {
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get current user: %w", err)
	}
	return &music.User{
		DisplayName: user.DisplayName,
		ID:          user.ID,
	}, nil
}

func (s *SpotifyClient) ListUserPlaylists(ctx context.Context, args music.ListUserPlaylistsArgs) ([]music.Playlist, error) {
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get current user: %w", err)
	}

	playlistPage, err := s.client.CurrentUsersPlaylists(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get user playlists: %w", err)
	}

	ownedPlaylists := make([]music.Playlist, 0)

	for page := 1; ; page++ {
		for _, playlist := range playlistPage.Playlists {
			if playlist.Owner.ID == user.ID {
				ownedPlaylists = append(ownedPlaylists, music.Playlist{
					Name:        playlist.Name,
					ID:          playlist.ID.String(),
					Description: playlist.Description,
					IsPublic:    playlist.IsPublic,
					Provider:    "spotify",
				})
			}
		}
		err = s.client.NextPage(ctx, playlistPage)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("could not get next page of playlists: %w", err)
		}
	}

	return ownedPlaylists, nil
}

func (s *SpotifyClient) ListTracksInPlaylist(ctx context.Context, args music.ListTracksInPlaylistArgs) ([]music.Track, error) {
	tracksList := make([]music.Track, 0)
	playlistTracks, err := s.client.GetPlaylistItems(ctx, spotify.ID(args.Playlist.ID))
	if err != nil {
		return nil, fmt.Errorf("could not get tracks for playlist %s: %w", args.Playlist.Name, err)
	}

	for page := 1; ; page++ {
		for _, playlistItem := range playlistTracks.Items {
			track := playlistItem.Track.Track
			if track == nil {
				fmt.Fprintf(os.Stderr, "Cant get details for track in playlist %s (%s)\n", args.Playlist.Name, args.Playlist.ID)
				continue
			}
			artists := make([]string, 0)
			for _, artist := range track.Artists {
				artists = append(artists, artist.Name)
			}
			tracksList = append(tracksList, music.Track{
				Name:    track.Name,
				Artists: artists,
				Album:   track.Album.Name,
				ID:      track.ID.String(),
			})
		}
		err = s.client.NextPage(ctx, playlistTracks)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("could not get next page of tracks: %w", err)
		}
	}
	return tracksList, nil
}

func (s *SpotifyClient) ListSavedTracks(ctx context.Context, args music.ListSavedTracksArgs) ([]music.Track, error) {
	savedTracksList := make([]music.Track, 0)
	savedTracks, err := s.client.CurrentUsersTracks(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get user saved tracks: %w", err)
	}

	for page := 1; ; page++ {
		for _, track := range savedTracks.Tracks {
			artists := make([]string, 0)
			for _, artist := range track.Artists {
				artists = append(artists, artist.Name)
			}
			savedTracksList = append(savedTracksList, music.Track{
				Name:    track.Name,
				Artists: artists,
				Album:   track.Album.Name,
				ID:      track.ID.String(),
			})
		}
		err = s.client.NextPage(ctx, savedTracks)
		if errors.Is(err, spotify.ErrNoMorePages) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("could not get next page of saved tracks: %w", err)
		}
	}

	return savedTracksList, nil
}

func (s *SpotifyClient) CreatePlaylist(ctx context.Context, args music.CreatePlaylistArgs) (music.Playlist, error) {
	user, err := s.User(ctx)
	if err != nil {
		return music.Playlist{}, fmt.Errorf("could not get current user: %w", err)
	}

	// description shouldnt have newlines, spotifys api doesnt like it
	playlist, err := s.client.CreatePlaylistForUser(ctx, user.ID, args.PlaylistDetails.Name, args.PlaylistDetails.Name, args.PlaylistDetails.IsPublic, false)
	if err != nil {
		return music.Playlist{}, fmt.Errorf("could not create playlist with name %s: %w", args.PlaylistDetails.Name, err)
	}

	return music.Playlist{
		Name:        playlist.Name,
		ID:          playlist.ID.String(),
		Description: playlist.Description,
		IsPublic:    playlist.IsPublic,
		Provider:    "spotify",
	}, nil
}

func (s *SpotifyClient) AddTracksToPlaylist(ctx context.Context, args music.AddTracksToPlaylistArgs) (music.Playlist, error) {
	maximumTracksPerRequest := 100
	totalTracks := len(args.Tracks)

	for i := 0; i < totalTracks; i += maximumTracksPerRequest {
		trackIDs := make([]spotify.ID, 0)
		for j := i; j < i+maximumTracksPerRequest && j < totalTracks; j++ {
			if args.Tracks[j].ID == "" {
				fmt.Fprintf(os.Stderr, "WARNING: Track %+v has no ID\n", args.Tracks[j])
				continue
			}
			trackIDs = append(trackIDs, spotify.ID(args.Tracks[j].ID))
		}
		_, err := s.client.AddTracksToPlaylist(ctx, spotify.ID(args.Playlist.ID), trackIDs...)
		if err != nil {
			return music.Playlist{}, fmt.Errorf("could not add tracks to playlist %s: %w", args.Playlist.Name, err)
		}
	}

	return args.Playlist, nil
}
