package main

import (
	"context"
	"log"
	"net/http"

	"github.com/noxworld-dev/xwis"
	"github.com/spf13/cobra"

	"github.com/noxworld-dev/lobby"
)

func init() {
	cmd := &cobra.Command{
		Use:   "serve",
		Short: "run the lobby web server",
	}
	fHost := cmd.Flags().String("host", ":8080", "host the server will listen on")
	fXWIS := cmd.Flags().Bool("xwis", true, "list games from XWIS as well")
	fXLogin := cmd.Flags().String("xlogin", "", "XWIS login to use")
	fXPass := cmd.Flags().String("xpass", "", "XWIS password to use")
	fXCache := cmd.Flags().Duration("xcache", lobby.DefaultTimeout/2, "XWIS cache duration")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var lb lobby.Lobby = lobby.NewLobby()
		if *fXWIS {
			log.Println("logging in to XWIS")
			c, err := xwis.NewClient(context.Background(), *fXLogin, *fXPass)
			if err != nil {
				return err
			}
			defer c.Close()
			var lx lobby.Lister = lobby.NewXWISWithClient(c)
			if *fXCache > 0 {
				lx = lobby.Cache(lx, *fXCache)
			}
			lb = lobby.Overlay(lb, lx)
		}
		lsrv := lobby.NewServer(lb)
		// TODO: auto TLS with Let's Encrypt
		srv := &http.Server{
			Addr: *fHost,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				log.Println(r.RemoteAddr, r.Method, r.URL)
				lsrv.ServeHTTP(w, r)
			}),
		}
		log.Println("serving on", srv.Addr)
		// TODO: serve metrics + pprof as well
		return srv.ListenAndServe()
	}
	Root.AddCommand(cmd)
}
