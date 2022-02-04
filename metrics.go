package lobby

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	sourceOpenNox = "opennox"
	sourceXWIS    = "xwis"
)

var (
	serverLabelNames = []string{"src", "addr", "port", "name", "vers", "mode", "map"}
	cntGameSeen      = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nox_game_seen",
		Help: "Number of times the game was seen online",
	}, serverLabelNames)
	cntGameExpired = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nox_game_expired",
		Help: "Number of times the game registration expired",
	}, serverLabelNames)
	cntGamePlayers = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "nox_game_players",
		Help: "Number of players in the game",
	}, serverLabelNames)
	cntXWISRooms = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "nox_xwis_rooms",
		Help: "Number of XWIS rooms",
	})
	cntXWISGames = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "nox_xwis_games",
		Help: "Number of XWIS rooms",
	})
	cntRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "nox_http_requests",
		Help: "Number of HTTP requests to the API",
	}, []string{"method", "endpoint", "agent"})
)

func serverLabels(src string, g *Game) []string {
	return []string{
		src,
		g.Address,
		strconv.Itoa(g.Port),
		g.Name,
		g.Vers,
		string(g.Mode),
		g.Map,
	}
}
