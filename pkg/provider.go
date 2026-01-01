package pkg

import (
	"context"

	"github.com/marco-almeida/trackvault/pkg/models"
)

type Provider interface {
	User(context.Context) (*models.User, error)
	ListUserPlaylists(context.Context, ListUserPlaylistsArgs) ([]models.Playlist, error)
	ListSavedTracks(context.Context, ListSavedTracksArgs) ([]models.Track, error)
	ListTracksInPlaylist(context.Context, ListTracksInPlaylistArgs) ([]models.Track, error)
	CreatePlaylist(context.Context, CreatePlaylistArgs) (models.Playlist, error)
	AddTracksToPlaylist(context.Context, AddTracksToPlaylistArgs) (models.Playlist, error)
}

type ListUserPlaylistsArgs struct{}

type ListSavedTracksArgs struct{}

type ListTracksInPlaylistArgs struct {
	Playlist models.Playlist
}

type CreatePlaylistArgs struct {
	PlaylistDetails models.Playlist
}

type AddTracksToPlaylistArgs struct {
	Playlist models.Playlist
	Tracks   []models.Track
}
