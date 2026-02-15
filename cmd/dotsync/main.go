package main

import (
	"fmt"
	"os"

	"github.com/wtfzambo/dotsync/cmd/dotsync/cmd"
)

// These variables are set at build time by goreleaser
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	builtBy = "unknown"
)

func main() {
	cmd.SetVersion(version, commit, date, builtBy)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
