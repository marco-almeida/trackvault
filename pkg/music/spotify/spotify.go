package spotify

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/google/uuid"
	"github.com/grokify/go-pkce"
	"github.com/marco-almeida/trackvault/pkg/music"
	utilsURL "github.com/marco-almeida/trackvault/pkg/utils"
	"github.com/zmb3/spotify/v2"
	spotifyauth "github.com/zmb3/spotify/v2/auth"
	"golang.org/x/oauth2"
)

const redirectHost = "127.0.0.1"
const redirectPort = "43721"
const clientID = "a0852cfbdbd24dcba2410c011ab29564"

var (
	redirectURI     = fmt.Sprintf("http://%s:%s/spotify/callback", redirectHost, redirectPort)
	auth            = spotifyauth.New(spotifyauth.WithRedirectURL(redirectURI), spotifyauth.WithScopes(spotifyauth.ScopeUserReadPrivate), spotifyauth.WithClientID(clientID))
	ch              = make(chan *spotify.Client)
	state           = uuid.New().String()
	codeVerifier, _ = pkce.NewCodeVerifier(-1)
	codeChallenge   = pkce.CodeChallengeS256(codeVerifier)
)

type SpotifyClient struct {
	Client *spotify.Client
}

// NewSpotifyClient creates an empty SpotifyClient, it needs to be logged in later
func NewSpotifyClient() music.Provider {
	return &SpotifyClient{}
}

func (s *SpotifyClient) Login(ctx context.Context, args music.LoginArgs) error {
	http.HandleFunc("/spotify/callback", completeAuth)
	go http.ListenAndServe(fmt.Sprintf("%s:%s", redirectHost, redirectPort), nil)

	url := auth.AuthURL(state,
		oauth2.SetAuthURLParam("code_challenge_method", pkce.MethodS256),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("client_id", clientID),
	)
	fmt.Fprintln(os.Stdout, "Please log in to Spotify by visiting the following page in your browser:", url)
	err := utilsURL.OpenURL(url)
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not open URL in browser:", err)
	}
	// wait for auth to complete
	client := <-ch

	s.Client = client
	return nil
}

func completeAuth(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.Token(r.Context(), state, r,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier))
	if err != nil {
		http.Error(w, "Couldn't get token", http.StatusForbidden)
		fmt.Fprintln(os.Stderr, err)
	}
	if st := r.FormValue("state"); st != state {
		http.NotFound(w, r)
		fmt.Fprintf(os.Stderr, "State mismatch: %s != %s\n", st, state)
	}
	// use the token to get an authenticated client
	client := spotify.New(auth.Client(r.Context(), tok))
	fmt.Fprintf(w, "Login Completed! You can now close this window.")
	ch <- client
}
