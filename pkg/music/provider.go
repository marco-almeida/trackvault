package music

import (
	"context"
)

type Provider interface {
	Login(context.Context, LoginArgs) error
	ListPlaylistsAndLikes(context.Context, ListPlaylistsAndLikesArgs) ([]Playlist, error)
	User(context.Context) (*User, error)
}

type Playlist struct {
	ID string
}

type User struct {
	ID          string
	DisplayName string
}

type LoginArgs struct{}

type ListPlaylistsAndLikesArgs struct{}
