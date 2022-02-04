package main

import (
	"context"
	"os"

	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "nox-lobby",
	Short: "Nox game lobby server",
}

func main() {
	if err := Root.Execute(); err != nil && err != context.Canceled {
		os.Exit(1)
	}
}
