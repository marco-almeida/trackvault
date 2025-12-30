package spotify

import (
	"context"
	"encoding/json"
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
	redirectURI     = fmt.Sprintf("http://%s:%s/spotify/callback", redirectHost, redirectPort)
	auth            = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate), spotifyauth.WithClientID(clientID))
	ch              = make(chan *oauth2.Token)
	state           = uuid.New().String()
	codeVerifier, _ = pkce.NewCodeVerifier(-1)
	codeChallenge   = pkce.CodeChallengeS256(codeVerifier)
)

type SpotifyClient struct {
	client *spotify.Client
}

// NewSpotifyClient creates an empty SpotifyClient, it needs to be logged in later
func NewSpotifyClient() music.Provider {
	return &SpotifyClient{}
}

// NewSpotifyClientFromToken creates a SpotifyClient from a refresh token
func NewSpotifyClientFromToken(ctx context.Context, token oauth2.Token) music.Provider {
	httpClient := auth.Client(ctx, &token)
	spotifyClient := spotify.New(httpClient, spotify.WithRetry(true))
	return &SpotifyClient{client: spotifyClient}
}

func (s *SpotifyClient) Login(ctx context.Context, args music.LoginArgs) error {
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

func (s *SpotifyClient) ListPlaylistsAndLikes(ctx context.Context, args music.ListPlaylistsAndLikesArgs) ([]music.Playlist, error) {
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get current user: %w", err)
	}
	fmt.Printf("Listing playlists and likes for spotify user %s (%s)\n", user.DisplayName, user.ID)
	return nil, nil
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

func completeAuth(w http.ResponseWriter, r *http.Request) {
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
