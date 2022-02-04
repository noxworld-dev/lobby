package lobby

import (
	"context"
)

// Overlay one lobby implementation over a second one.
// Games from the overlay will override games from the base.
// Registration will happen only on the overlay Lobby.
func Overlay(over Lobby, base Lister) Lobby {
	return &overlay{over: over, base: base}
}

type overlay struct {
	over Lobby
	base Lister
}

func (l *overlay) RegisterGame(ctx context.Context, s *Game) error {
	return l.over.RegisterGame(ctx, s)
}

func (l *overlay) ListGames(ctx context.Context) ([]GameInfo, error) {
	list1, err1 := l.base.ListGames(ctx)
	list2, err2 := l.over.ListGames(ctx)
	if len(list1)+len(list2) == 0 {
		if err2 != nil {
			// overlay error takes priority
			return nil, err2
		}
		// return base error
		return nil, err1
	}
	// if there are any results at all - suppress errors
	byAddr := make(map[gameKey]GameInfo, len(list2))
	for _, g := range list1 {
		byAddr[g.gameKey()] = g
	}
	for _, g := range list2 {
		byAddr[g.gameKey()] = g
	}
	list := make([]GameInfo, 0, len(byAddr))
	for _, g := range byAddr {
		list = append(list, g)
	}
	sortGameInfos(list)
	return list, nil
}
