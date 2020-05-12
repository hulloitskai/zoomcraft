package graphql

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"

	"go.stevenxie.me/zoomcraft/backend/minecraft"
)

func (r *queryResolver) Players(ctx context.Context) ([]*minecraft.Player, error) {
	return r.Resolver.Players.List(ctx)
}

func (r *queryResolver) Player(ctx context.Context, username string) (*minecraft.Player, error) {
	p, err := r.Resolver.Players.Get(ctx, username)
	if err != nil {
		if errors.Is(err, minecraft.ErrNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}
