package main

import (
	"os"

	"github.com/gcarthew/ajira/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(cli.ExitCodeFromError(err))
	}
}
