package lobby

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

var _ Lobby = &Client{}

// Client is an HTTP Nox lobby client.
type Client struct {
	client    *http.Client
	serverURL string
	agent     string
}

// NewClient will create new client for our server
func NewClient(serverURL string) *Client {
	return NewClientWith(serverURL, http.DefaultClient)
}

// NewClientWith accepts URL and custom HTTP client to use.
func NewClientWith(serverURL string, client *http.Client) *Client {
	return &Client{
		client:    client,
		serverURL: serverURL,
	}
}

// SetUserAgent sets the User-Agent header for requests. Format should be "AppName/1.2.3".
// It is advised to set this to something unique for each app using this library.
func (c *Client) SetUserAgent(agent string) {
	c.agent = agent
}

// ListGames implements Lobby.
func (c *Client) ListGames(ctx context.Context) ([]GameInfo, error) {
	var out ServerListResp
	err := c.sendRequest(ctx, http.MethodGet, "/api/v0/games/list", nil, &out)
	return out, err
}

// RegisterGame implements Lobby.
func (c *Client) RegisterGame(ctx context.Context, s *Game) error {
	if err := c.sendRequest(ctx, http.MethodPost, "/api/v0/games/register", s, nil); err != nil {
		return err
	}
	return nil
}

func (c *Client) sendRequest(ctx context.Context, meth string, path string, body interface{}, dst interface{}) error {
	var rbody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rbody = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, meth, c.serverURL+path, rbody)
	if err != nil {
		return err
	}
	if c.agent != "" {
		req.Header.Set("User-Agent", c.agent)
	}
	if rbody != nil {
		req.Header.Add("Content-Type", "application/json")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var out Response
	out.Result = dst
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return err
	}
	if out.Err != "" {
		return errors.New(out.Err)
	} else if resp.StatusCode/100 != 2 {
		return errors.New("status: " + resp.Status)
	}
	return nil
}
