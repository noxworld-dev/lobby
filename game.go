package lobby

import "context"

const (
	// DefaultGamePort is a default UDP port for Nox games.
	DefaultGamePort = 18590
)

// GameMode is a Nox game mode.
type GameMode string

const (
	ModeKOTR        = GameMode("kotr")
	ModeCTF         = GameMode("ctf")
	ModeFlagBall    = GameMode("flagball")
	ModeChat        = GameMode("chat")
	ModeArena       = GameMode("arena")
	ModeElimination = GameMode("elimination")
	ModeQuest       = GameMode("quest")
	ModeCoop        = GameMode("coop")
	ModeCustom      = GameMode("custom")
)

// GameAccess specifies access for the game (open, password-protected, etc).
type GameAccess string

const (
	AccessOpen     = GameAccess("open")
	AccessPassword = GameAccess("pass")
	AccessClosed   = GameAccess("closed")
)

// Resolution is a max resolution used for the game.
// Historically Nox used a limited resolution. For HD-aware servers, HighRes should be set.
type Resolution struct {
	HighRes bool `json:"high_res,omitempty"`
	Width   int  `json:"width,omitempty"`
	Height  int  `json:"height,omitempty"`
}

func (v *Resolution) Clone() *Resolution {
	if v == nil {
		return nil
	}
	v2 := *v
	return &v2
}

// GameHost is an interface for the game server.
type GameHost interface {
	// GameInfo returns current information about the active game.
	GameInfo(ctx context.Context) (*Game, error)
}

// Game is an information about the Nox game, as provided by the server hosting it.
// See GameInfo for an information returned by the lobby server.
type Game struct {
	Name    string      `json:"name"`
	Address string      `json:"addr,omitempty"`
	Port    int         `json:"port,omitempty"`
	Map     string      `json:"map"`
	Mode    GameMode    `json:"mode"`
	Access  GameAccess  `json:"access,omitempty"`
	Vers    string      `json:"vers,omitempty"`
	Res     Resolution  `json:"res,omitempty"`
	Players PlayersInfo `json:"players"`
	Quest   *QuestInfo  `json:"quest,omitempty"`
}

func (g *Game) Clone() *Game {
	if g == nil {
		return nil
	}
	g2 := *g
	g2.Players = *g.Players.Clone()
	g2.Res = *g.Res.Clone()
	g2.Quest = g.Quest.Clone()
	return &g2
}

// PlayersInfo is an information about players in a specific game.
type PlayersInfo struct {
	Cur int `json:"cur"`
	Max int `json:"max"`
}

func (v *PlayersInfo) Clone() *PlayersInfo {
	if v == nil {
		return nil
	}
	v2 := *v
	return &v2
}

// QuestInfo is additional information for Nox Quest game mode.
type QuestInfo struct {
	Stage int `json:"stage"`
}

func (v *QuestInfo) Clone() *QuestInfo {
	if v == nil {
		return nil
	}
	v2 := *v
	return &v2
}
