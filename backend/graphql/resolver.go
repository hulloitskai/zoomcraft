package graphql

import "go.stevenxie.me/covidcraft/backend/minecraft"

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// A Resolver implements a ResolverRoot.
type Resolver struct {
	Players minecraft.PlayerService
}

var _ ResolverRoot = (*Resolver)(nil)
