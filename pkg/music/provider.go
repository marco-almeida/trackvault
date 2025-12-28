package music

import "context"

type Provider interface {
	Login(context.Context, LoginArgs) error
}

type LoginArgs struct{}
