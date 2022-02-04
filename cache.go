package lobby

import (
	"context"
	"sync"
	"time"
)

// Cache creates a cache over a game Lister.
// Expiration time controls how often the cache is invalidated.
func Cache(l Lister, expire time.Duration) Lister {
	if expire == 0 {
		expire = DefaultTimeout / 2
	}
	return &listCache{l: l, exp: expire}
}

type listCache struct {
	l   Lister
	exp time.Duration

	mu   sync.RWMutex
	last time.Time
	list []GameInfo
}

func (l *listCache) listGames(ctx context.Context) ([]GameInfo, error) {
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.last.Add(l.exp).After(now) {
		return l.list, nil
	}
	list, err := l.l.ListGames(ctx)
	if err != nil {
		return nil, err
	}
	l.list, l.last = list, now
	return l.list, nil
}

func (l *listCache) ListGames(ctx context.Context) ([]GameInfo, error) {
	now := time.Now()
	l.mu.RLock()
	ok := l.last.Add(l.exp).After(now)
	list := l.list
	l.mu.RUnlock()
	if !ok {
		var err error
		list, err = l.listGames(ctx)
		if err != nil {
			return nil, err
		}
	}
	out := make([]GameInfo, 0, len(list))
	for _, g := range list {
		out = append(out, *g.Clone())
	}
	return out, nil
}
