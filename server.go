package lobby

import (
	"encoding/json"
	"errors"
	"net"
	"net/http"
)

var _ http.Handler = (*Server)(nil)

// Server is an HTTP Nox lobby server.
type Server struct {
	l         Lobby
	mux       *http.ServeMux
	trustAddr bool // trust IP sent by a remote
}

// IPResp represents a response to the address request.
type IPResp struct {
	IP string `json:"ip"`
}

// ServerListResp represents a response to the server list request.
type ServerListResp []GameInfo

// Response wraps all other HTTP responses to separate errors from the rest of the response.
type Response struct {
	Result interface{} `json:"data,omitempty"`
	Err    string      `json:"error,omitempty"`
}

// NewServer creates a new http.Handler from a Lobby implementation.
func NewServer(l Lobby) *Server {
	api := &Server{l: l, mux: http.NewServeMux()}
	api.mux.HandleFunc("/api/v0/address", api.Address)
	api.mux.HandleFunc("/api/v0/games/list", api.ServersList)
	api.mux.HandleFunc("/api/v0/games/register", api.RegisterServer)
	return api
}

func (api *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	api.mux.ServeHTTP(w, r)
}

// jsonResponse writes response, wrapping it into JSON format.
func (api *Server) jsonResponse(w http.ResponseWriter, code int, data interface{}) {
	if code == 0 {
		code = http.StatusOK
	}
	if data == nil {
		// result field in JSON is omitempty, but we want to keep it for this case
		data = (*struct{})(nil)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(&Response{
		Result: data,
	})
}

// jsonError writes an error, wrapping it into JSON format.
func (api *Server) jsonError(w http.ResponseWriter, code int, err error) {
	if code == 0 {
		code = http.StatusInternalServerError
	}
	if err == nil {
		err = errors.New(http.StatusText(code))
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(&Response{
		Err: err.Error(),
	})
}

func (api *Server) getAddress(r *http.Request) (string, error) {
	if r.RemoteAddr == "" {
		return "", errors.New("cannot detect IP address")
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	return ip, err
}

func (api *Server) Address(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		ip, err := api.getAddress(r)
		if err != nil {
			api.jsonError(w, http.StatusBadRequest, err)
			return
		}
		api.jsonResponse(w, 0, IPResp{IP: ip})
	default:
		api.jsonError(w, http.StatusMethodNotAllowed, nil)
	}
}

func (api *Server) RegisterServer(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		// set limit to avoid giant requests - 1MB
		body := http.MaxBytesReader(w, r.Body, 1024*1024)
		var req Game
		if err := json.NewDecoder(body).Decode(&req); err != nil {
			api.jsonError(w, http.StatusBadRequest, err)
			return
		}
		if !api.trustAddr {
			addr, err := api.getAddress(r)
			if err != nil {
				api.jsonError(w, http.StatusBadRequest, err)
				return
			}
			req.Address = addr
		}
		err := api.l.RegisterGame(r.Context(), &req)
		if err != nil {
			api.jsonError(w, http.StatusBadRequest, err)
			return
		}
		api.jsonResponse(w, 0, nil)
	default:
		api.jsonError(w, http.StatusMethodNotAllowed, nil)
	}
}

func (api *Server) ServersList(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		list, err := api.l.ListGames(r.Context())
		if err != nil {
			api.jsonError(w, http.StatusInternalServerError, err)
			return
		}
		api.jsonResponse(w, 0, ServerListResp(list))
	default:
		api.jsonError(w, http.StatusMethodNotAllowed, nil)
	}
}
