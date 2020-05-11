package minecraft

import (
	"context"
	"time"

	"golang.org/x/sync/singleflight"
)

// A PlayerServiceCache is used to cache requests on a PlayerService.
type PlayerServiceCache struct {
	MaxAge time.Duration `json:"maxAge"`
}

// Apply returns a PlayerService that caches requests using PlayerServiceCache.
func (cache *PlayerServiceCache) Apply(svc PlayerService) PlayerService {
	return &playerServiceCache{
		cache:     cache,
		origin:    svc,
		getResult: make(map[string]*Player),
		getCalled: make(map[string]time.Time),
	}
}

type playerServiceCache struct {
	origin PlayerService
	cache  *PlayerServiceCache

	getGroup  singleflight.Group
	getResult map[string]*Player
	getCalled map[string]time.Time

	listGroup  singleflight.Group
	listResult []*Player
	listCalled time.Time
}

func (svc *playerServiceCache) List(ctx context.Context) ([]*Player, error) {
	v, err, _ := svc.listGroup.Do("", func() (interface{}, error) {
		due := svc.listCalled.Add(svc.cache.MaxAge)
		if time.Now().After(due) {
			players, err := svc.origin.List(ctx)
			if err != nil {
				return nil, err
			}
			svc.listResult = players
			svc.listCalled = time.Now()
		}
		return svc.listResult, nil
	})
	if v == nil {
		return nil, err
	}
	return v.([]*Player), err
}

func (svc *playerServiceCache) Get(ctx context.Context, username string) (*Player, error) {
	v, err, _ := svc.getGroup.Do(username, func() (interface{}, error) {
		due := svc.getCalled[username].Add(svc.cache.MaxAge)
		if time.Now().After(due) {
			player, err := svc.origin.Get(ctx, username)
			if err != nil {
				return nil, err
			}
			svc.getResult[username] = player
			svc.getCalled[username] = time.Now()
		}
		return svc.getResult[username], nil
	})
	if v == nil {
		return nil, err
	}
	return v.(*Player), err
}
