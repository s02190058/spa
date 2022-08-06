package http

import (
	"context"
	"errors"

	"github.com/s02190058/spa/internal/entity"
)

var (
	ErrBadContext = errors.New("bad context")
)

type ctxUserKey int

var userKey ctxUserKey

func contextWithUser(ctx context.Context, user *entity.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

func userFromContext(ctx context.Context) (*entity.User, error) {
	user, ok := ctx.Value(userKey).(*entity.User)
	if !ok {
		return nil, ErrBadContext
	}

	return user, nil
}
