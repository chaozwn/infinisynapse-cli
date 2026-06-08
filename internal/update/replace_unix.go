//go:build !windows

package update

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// replaceExecutable swaps the running binary on Unix-like systems. The current
// binary is renamed to a .old backup before the new one takes its place, so a
// failed rename leaves the original intact.
func replaceExecutable(exePath, newPath string) error {
	if err := os.Chmod(newPath, 0o755); err != nil {
		return fmt.Errorf("chmod new binary: %w", err)
	}

	backup := exePath + ".old"
	_ = os.Remove(backup)

	if err := os.Rename(exePath, backup); err != nil {
		return fmt.Errorf("cannot back up current binary (need write permission on %s?): %w", exePath, err)
	}
	if err := os.Rename(newPath, exePath); err != nil {
		// Roll back so the user is not left without a working binary.
		_ = os.Rename(backup, exePath)
		return fmt.Errorf("cannot install new binary: %w", err)
	}

	// Clear the macOS quarantine attribute so Gatekeeper does not block it.
	if runtime.GOOS == "darwin" {
		if xattr, err := exec.LookPath("xattr"); err == nil {
			_ = exec.Command(xattr, "-d", "com.apple.quarantine", exePath).Run()
		}
	}

	_ = os.Remove(backup)
	return nil
}
