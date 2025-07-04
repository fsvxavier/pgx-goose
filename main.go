package main

import (
	"os"

	"github.com/fsvxavier/pgx-goose/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
