package music

import (
	"context"
)

type Provider interface {
	User(context.Context) (*User, error)
	ListUserPlaylists(context.Context, ListUserPlaylistsArgs) ([]Playlist, error)
	ListSavedTracks(context.Context, ListSavedTracksArgs) ([]Track, error)
	ListTracksInPlaylist(context.Context, ListTracksInPlaylistArgs) ([]Track, error)
	CreatePlaylist(context.Context, CreatePlaylistArgs) (Playlist, error)
	AddTracksToPlaylist(context.Context, AddTracksToPlaylistArgs) (Playlist, error)
}

type Playlist struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsVirtual   bool   `json:"is_virtual"`
	IsPublic    bool   `json:"public"`
	Provider    string `json:"provider"`
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

type ListUserPlaylistsArgs struct{}

type ListSavedTracksArgs struct{}

type ListTracksInPlaylistArgs struct {
	Playlist Playlist
}

type CreatePlaylistArgs struct {
	PlaylistDetails Playlist
}

type AddTracksToPlaylistArgs struct {
	Playlist Playlist
	Tracks   []Track
}
