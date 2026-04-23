package main

import (
	"os"

	"github.com/grantcarthew/ajira/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(cli.ExitCodeFromError(err))
	}
}
