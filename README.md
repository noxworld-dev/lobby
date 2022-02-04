# Nox lobby server

[![Go Reference](https://pkg.go.dev/badge/github.com/noxworld-dev/lobby.svg)](https://pkg.go.dev/github.com/noxworld-dev/lobby)

This project provides a Nox game lobby which exposes a simple HTTP API for both listing and registering Nox game servers.

XWIS games will also appear in the list returned by the API, so the lobby is backward-compatible.
Due to technical limitations, however, games registered via HTTP won't be registered on XWIS.

The main use case for the lobby is to support OpenNox, but the API can also be used for bots
that want to notify players about currently active Nox games.

The simplest way to run the lobby server is via Docker:

```bash
docker run -d --restart always --name nox-lobby -p 8080:80 ghcr.io/noxworld-dev/lobby
```

To get a list of games via local lobby:

```bash
curl 'http://127.0.0.1:8080/api/v0/games/list'
```

A Go client library for HTTP API is also available (see [docs](https://pkg.go.dev/github.com/noxworld-dev/lobby)).