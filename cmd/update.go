package cmd

import (
	"fmt"
	"os"

	"github.com/chaozwn/infinisynapse-cli/internal/update"
)

var (
	updateFlag      bool
	updateCheckFlag bool
)

// runUpdate executes the self-update flow. It is invoked early from Execute()
// so it never triggers config initialization.
func runUpdate() error {
	return update.Run(update.Options{
		CurrentVersion: Version,
		CheckOnly:      updateCheckFlag,
	})
}

// handleUpdateFlag intercepts `agent_infini --update [--check]` before cobra
// runs, mirroring the early handling used for --skill. It returns true when the
// update flow took over and the program should exit with the given code.
func handleUpdateFlag(args []string) (handled bool, exitCode int) {
	wantUpdate := false
	for _, a := range args {
		switch a {
		case "--update":
			wantUpdate = true
		case "--check":
			updateCheckFlag = true
		}
	}
	if !wantUpdate {
		return false, 0
	}
	if err := runUpdate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return true, 1
	}
	return true, 0
}
