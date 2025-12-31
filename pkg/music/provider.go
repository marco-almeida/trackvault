package music

import (
	"context"
)

type Provider interface {
	Login(context.Context, LoginArgs) error
	User(context.Context) (*User, error)
	ListUserPlaylists(context.Context, ListUserPlaylistsArgs) ([]Playlist, error)
	ListSavedTracks(context.Context, ListSavedTracksArgs) ([]Track, error)
	ListTracksInPlaylist(context.Context, Playlist) ([]Track, error)
}

type Playlist struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsVirtual bool   `json:"is_virtual"`
}

type Track struct {
	ID      string   `json:"id"`
	Artists []string `json:"artists"`
	Name    string   `json:"name"`
	Album   string   `json:"album"`
}

type User struct {
	ID          string
	DisplayName string
}

type LoginArgs struct{}

type ListUserPlaylistsArgs struct{}

type ListSavedTracksArgs struct{}
