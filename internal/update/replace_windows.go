//go:build windows

package update

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

// replaceExecutable swaps the running binary on Windows. A running .exe cannot
// be overwritten in place, so we spawn a detached batch script that waits for
// this process to exit, moves the freshly downloaded binary over the old one,
// and removes itself. The current process then exits.
func replaceExecutable(exePath, newPath string) error {
	dir := filepath.Dir(exePath)
	batPath := filepath.Join(dir, "agent_infini_update.bat")

	// %~dp0 keeps paths stable regardless of the working directory.
	script := fmt.Sprintf(`@echo off
ping 127.0.0.1 -n 2 >nul
move /Y "%s" "%s" >nul
del "%%~f0"
`, newPath, exePath)

	if err := os.WriteFile(batPath, []byte(script), 0o644); err != nil {
		return fmt.Errorf("cannot write update script: %w", err)
	}

	cmd := exec.Command("cmd.exe", "/C", batPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x00000008} // DETACHED_PROCESS
	if err := cmd.Start(); err != nil {
		_ = os.Remove(batPath)
		return fmt.Errorf("cannot launch update script: %w", err)
	}

	fmt.Println("\nThe new binary will be installed once this process exits.")
	fmt.Println("Reopen your terminal and run 'agent_infini version' to verify.")

	// Exit so the running .exe is released and the script can replace it.
	os.Exit(0)
	return nil
}
