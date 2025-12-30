package music

import "context"

type Provider interface {
	Login(context.Context, LoginArgs) (*User, error)
}

type User struct {
	ID          string
	DisplayName string
}

type LoginArgs struct{}
