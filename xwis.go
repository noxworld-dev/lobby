package lobby

import (
	"context"
	"log"
	"time"

	"github.com/noxworld-dev/xwis"
)

// NewXWISWithClient creates a Lister for a Nox XWIS lobby using an existing xwis.Client.
func NewXWISWithClient(c *xwis.Client) Lister {
	return &xwisLister{c: c}
}

func xwisGameMode(v xwis.MapType) GameMode {
	switch v {
	case xwis.MapTypeKOTR:
		return ModeKOTR
	case xwis.MapTypeCTF:
		return ModeCTF
	case xwis.MapTypeFlagBall:
		return ModeFlagBall
	case xwis.MapTypeChat:
		return ModeChat
	case xwis.MapTypeArena:
		return ModeArena
	case xwis.MapTypeElimination:
		return ModeElimination
	case xwis.MapTypeCoop:
		return ModeCoop
	case xwis.MapTypeQuest:
		return ModeQuest
	}
	return ModeCustom
}

func gameFromXWIS(g *xwis.GameInfo) Game {
	var q *QuestInfo
	if g.MapType == xwis.MapTypeQuest {
		q = &QuestInfo{Stage: g.FragLimit}
	}
	return Game{
		Name:    g.Name,
		Address: g.Addr,
		Port:    DefaultGamePort, // TODO
		Map:     g.Map,
		Mode:    xwisGameMode(g.MapType),
		Players: PlayersInfo{
			Cur: g.Players,
			Max: g.MaxPlayers,
		},
		Quest: q,
	}
}

type xwisLister struct {
	c *xwis.Client
}

func (l *xwisLister) ListGames(ctx context.Context) ([]GameInfo, error) {
	list, err := l.c.ListRooms(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	var out []GameInfo
	for _, r := range list {
		g := r.Game
		if g == nil {
			continue
		}
		out = append(out, GameInfo{
			Game: gameFromXWIS(g), SeenAt: now,
		})
	}
	log.Printf("xwis: %d rooms, %d games", len(list), len(out))
	return out, nil
}
