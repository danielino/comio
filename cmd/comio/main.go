package main

import (
	"fmt"
	"os"

	"github.com/danielino/comio/internal/cli"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	if err := cli.Execute(Version, BuildTime); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
