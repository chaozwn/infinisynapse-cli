package main

import (
	"os"

	"github.com/chaozwn/infinisynapse-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
