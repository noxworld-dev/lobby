package lobby

import (
	"context"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testTimeout = 30 * time.Millisecond
)

var (
	server1     = Game{Name: "test1", Address: "1.1.1.1"}
	initServers = []Game{
		server1,
		{Name: "test2", Address: "2.2.2.2"},
		{Name: "test3", Address: "3.3.3.3"},
		{Name: "test4", Address: "4.4.4.4"},
		{Name: "test5", Address: "5.5.5.5"},
		{Name: "test6", Address: "6.6.6.6"},
	}
)

func testSetDefault(s *Game) {
	s.Map = "testmap"
	s.Mode = ModeArena
	s.Vers = "v0.0.0"
	s.Players.Max = 32
	s.Port = DefaultGamePort
}

func init() {
	testSetDefault(&server1)
	for i := range initServers {
		testSetDefault(&initServers[i])
	}
}

var testList = []struct {
	name string
	test func(t testing.TB, srv Lobby)
}{
	{name: "register", test: testLobbyRegister},
	{name: "keep registered", test: testLobbyKeepRegistered},
	{name: "register concurrent", test: testLobbyRegisterConcurrent},
	{name: "list concurrent", test: testLobbyListConcurrent},
	{name: "mix concurrent", test: testLobbyMixConcurrent},
}

// RunLobbyTests runs all lobby tests using the constructor provided.
func RunLobbyTests(t *testing.T, fnc func(t testing.TB) Lobby) {
	for _, c := range testList {
		t.Run(c.name, func(t *testing.T) {
			newLobby := fnc(t)
			c.test(t, newLobby)
		})
	}
}

// TestLobby tests core implementation of lobby
func TestLobby(t *testing.T) {
	RunLobbyTests(t, func(t testing.TB) Lobby {
		l := NewLobby()
		l.SetTimeout(testTimeout)
		return l
	})
}

// TestLobbyHTTP tests HTTP client-server pair wrapping the lobby
func TestLobbyHTTP(t *testing.T) {
	RunLobbyTests(t, func(t testing.TB) Lobby {
		l := NewLobby()
		l.SetTimeout(testTimeout)
		api := NewServer(l)
		// have to set it to emulate multiple clients
		api.trustAddr = true

		srv := &http.Server{Handler: api}
		t.Cleanup(func() {
			_ = srv.Close()
		})
		lis, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := lis.Addr()
		t.Logf("using address: %q", addr)
		t.Cleanup(func() {
			_ = lis.Close()
		})
		go srv.Serve(lis)
		return NewClient("http://" + addr.String())
	})
}

func testLobbyRegister(t testing.TB, l Lobby) {
	ctx := context.Background()
	full := Game{
		Name:    "test",
		Address: "1.1.1.1",
	}
	testSetDefault(&full)

	s := full
	s.Address = ""
	err := l.RegisterGame(ctx, &s)
	require.Error(t, err, "expected error for empty address")

	s = full
	s.Name = ""
	err = l.RegisterGame(ctx, &s)
	require.Error(t, err, "expected error for empty name")

	s = full
	s.Vers = ""
	err = l.RegisterGame(ctx, &s)
	require.Error(t, err, "expected error for empty version")

	s = full
	s.Map = ""
	err = l.RegisterGame(ctx, &s)
	require.Error(t, err, "expected error for empty map")

	s = full
	s.Mode = ""
	err = l.RegisterGame(ctx, &s)
	require.Error(t, err, "expected error for empty mode")

	s = full
	s.Name = " test "
	err = l.RegisterGame(ctx, &s)
	require.Error(t, err, "expected error for spaces in name")

	s = full
	s.Players.Max = 0
	err = l.RegisterGame(ctx, &s)
	require.Error(t, err, "expected error for empty max players")

	expectServers(t, l, nil)

	s = full
	err = l.RegisterGame(ctx, &s)
	require.NoError(t, err)
	expectServers(t, l, []Game{s})

	// wait half of timeout - should still be there
	time.Sleep(testTimeout / 2)
	expectServers(t, l, []Game{s})

	// wait for the whole timeout - should expire
	time.Sleep(testTimeout)
	expectServers(t, l, nil)

	// register again, try removing the port to get a default
	s.Port = 0
	err = l.RegisterGame(ctx, &s)
	require.NoError(t, err)
	s = full
	expectServers(t, l, []Game{s})

	// check if we can cause duplicates by changing name
	s.Name = "test1"
	err = l.RegisterGame(ctx, &s)
	require.NoError(t, err)
	expectServers(t, l, []Game{s})

	// wait for half timeout, refresh, then wait for 3/4, should still be there
	time.Sleep(testTimeout / 2)
	err = l.RegisterGame(ctx, &s)
	require.NoError(t, err)
	time.Sleep(testTimeout * 3 / 4)
	expectServers(t, l, []Game{s})

	// refresh again, write second server on a different IP and the third on different port, but same address
	err = l.RegisterGame(ctx, &s)
	require.NoError(t, err)
	s2 := full
	s2.Name = "test 2"
	s2.Address = "2.2.2.2"
	err = l.RegisterGame(ctx, &s2)
	require.NoError(t, err)
	s3 := full
	s3.Name = "test 3"
	s3.Port = DefaultGamePort + 10
	err = l.RegisterGame(ctx, &s3)
	require.NoError(t, err)
	expectServers(t, l, []Game{s, s3, s2})

	// make all expire
	time.Sleep(testTimeout)
	expectServers(t, l, nil)
}

type testGameHost struct {
	info *Game
}

func (h testGameHost) GameInfo(ctx context.Context) (*Game, error) {
	info := *h.info
	return &info, nil
}

func testLobbyKeepRegistered(t testing.TB, l Lobby) {
	ctx := context.Background()
	errc := make(chan error, 1)
	ctx, cancel := context.WithTimeout(context.Background(), testTimeout*5)
	defer cancel()
	go func() {
		ticker := time.NewTicker(testTimeout / 3)
		defer ticker.Stop()
		errc <- KeepRegistered(ctx, l, ticker.C, testGameHost{info: &server1})
	}()
	err := <-errc
	require.NoError(t, err)
	expectServers(t, l, []Game{server1})
}

func registerServer(t testing.TB, testLobby Lobby, srv Game) {
	ctx := context.Background()
	err := testLobby.RegisterGame(ctx, &srv)
	require.NoError(t, err)
}

func expectServers(t testing.TB, l Lobby, exp []Game) {
	ctx := context.Background()
	got, err := l.ListGames(ctx)
	require.NoError(t, err)
	var goti []Game
	for _, s := range got {
		goti = append(goti, s.Game)
	}
	require.Equal(t, exp, goti)
}

func testLobbyRegisterConcurrent(t testing.TB, testLobby Lobby) {
	ready := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			registerServer(t, testLobby, server1)
		}()
	}
	close(ready)
	wg.Wait()
}

func testLobbyListConcurrent(t testing.TB, testLobby Lobby) {
	for i, s := range initServers {
		if i == 2 {
			// expire first two records
			time.Sleep(2 * testTimeout)
			expectServers(t, testLobby, nil)
		}
		registerServer(t, testLobby, s)
	}
	expectServers(t, testLobby, initServers[2:])
	ready := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			expectServers(t, testLobby, initServers[2:])
		}()
	}
	close(ready)
	wg.Wait()
}

func testLobbyMixConcurrent(t testing.TB, testLobby Lobby) {
	ctx := context.Background()
	ready := make(chan struct{})
	var wg sync.WaitGroup
	// Run 10 routines calling ListGames, each should call it 3 times in a loop.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			for j := 0; j < 3; j++ {
				_, err := testLobby.ListGames(ctx)
				require.NoError(t, err)
			}
		}()
	}
	// Without waiting, run 10 more routines that call RegisterGame.
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-ready
			registerServer(t, testLobby, server1)
		}()
	}
	close(ready)
	wg.Wait()
}
