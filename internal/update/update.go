// Package update implements self-update for the agent_infini CLI.
//
// It reads a manifest published alongside the install scripts at
// plugins/infini_cli/manifest.json, compares the latest version against the
// running binary, downloads the matching platform binary, verifies its SHA256
// checksum, and atomically replaces the current executable in place.
package update

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	// DefaultManifestURL points at the version manifest used for self-update.
	DefaultManifestURL = "https://infinisynapse.oss-cn-shanghai.aliyuncs.com/plugins/infini_cli/manifest.json"
	// DefaultBaseURL is the prefix for versioned platform binaries.
	DefaultBaseURL = "https://infinisynapse.oss-cn-shanghai.aliyuncs.com/plugins/infini_cli"

	httpTimeout = 60 * time.Second
)

// Options controls a single update run.
type Options struct {
	// CurrentVersion is the version compiled into the running binary.
	CurrentVersion string
	// CheckOnly only reports whether an update is available without installing.
	CheckOnly bool
	// Force reinstalls even when the latest version is not newer.
	Force bool
	// TargetVersion pins a specific version instead of the manifest's latest.
	TargetVersion string
	// ManifestURL and BaseURL allow overriding the default endpoints (testing).
	ManifestURL string
	BaseURL     string
}

// Manifest is the JSON document describing the latest CLI release.
type Manifest struct {
	Version    string              `json:"version"`
	ReleasedAt string              `json:"released_at"`
	Artifacts  map[string]Artifact `json:"artifacts"`
}

// Artifact describes a single platform binary in the manifest.
type Artifact struct {
	File   string `json:"file"`
	SHA256 string `json:"sha256"`
	Size   int64  `json:"size"`
}

// Run performs the update flow described by opts.
func Run(opts Options) error {
	if opts.ManifestURL == "" {
		opts.ManifestURL = DefaultManifestURL
	}
	if opts.BaseURL == "" {
		opts.BaseURL = DefaultBaseURL
	}

	platform, fileName, err := resolvePlatform()
	if err != nil {
		return err
	}

	fmt.Println("Checking for updates...")
	manifest, err := fetchManifest(opts.ManifestURL)
	if err != nil {
		return fmt.Errorf("failed to fetch manifest: %w", err)
	}

	latest := strings.TrimPrefix(manifest.Version, "v")
	current := strings.TrimPrefix(opts.CurrentVersion, "v")
	target := latest
	if opts.TargetVersion != "" {
		target = strings.TrimPrefix(opts.TargetVersion, "v")
	}

	fmt.Printf("  Current : v%s\n", current)
	fmt.Printf("  Latest  : v%s\n", latest)

	if isDevVersion(current) {
		fmt.Println("  Note    : current build looks like a development version.")
	}

	upToDate := compareVersion(current, target) >= 0
	if opts.TargetVersion == "" && upToDate && !opts.Force {
		fmt.Printf("\nagent_infini is up to date (v%s).\n", current)
		return nil
	}

	if opts.CheckOnly {
		fmt.Printf("\nUpdate available: v%s -> v%s\n", current, target)
		fmt.Println("Run 'agent_infini --update' to install it.")
		return nil
	}

	artifact, ok := manifest.Artifacts[platform]
	if !ok {
		return fmt.Errorf("manifest has no artifact for platform %q", platform)
	}
	if artifact.File == "" {
		artifact.File = fileName
	}

	exePath, err := currentExecutable()
	if err != nil {
		return err
	}

	downloadURL := fmt.Sprintf("%s/%s/%s/%s",
		strings.TrimRight(opts.BaseURL, "/"), platform, target, artifact.File)

	fmt.Printf("\nDownloading agent_infini v%s (%s)...\n", target, platform)
	fmt.Printf("  %s\n", downloadURL)

	tmpPath := exePath + ".new"
	if err := downloadFile(downloadURL, tmpPath); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer os.Remove(tmpPath)

	if artifact.SHA256 != "" {
		fmt.Print("Verifying checksum... ")
		sum, err := sha256File(tmpPath)
		if err != nil {
			return err
		}
		if !strings.EqualFold(sum, artifact.SHA256) {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", artifact.SHA256, sum)
		}
		fmt.Println("ok")
	}

	fmt.Printf("Installing to %s...\n", exePath)
	if err := replaceExecutable(exePath, tmpPath); err != nil {
		return err
	}

	fmt.Printf("\nUpdate complete (v%s -> v%s).\n", current, target)
	fmt.Println("Run 'agent_infini version' to verify.")
	return nil
}

// resolvePlatform maps the runtime OS/arch to the manifest platform key and
// binary file name, matching the install scripts' naming convention.
func resolvePlatform() (platform, fileName string, err error) {
	switch runtime.GOOS {
	case "darwin":
		switch runtime.GOARCH {
		case "arm64":
			return "darwin-arm64", "agent_infini", nil
		case "amd64":
			return "darwin-x64", "agent_infini", nil
		}
	case "linux":
		switch runtime.GOARCH {
		case "amd64":
			return "linux-x64", "agent_infini", nil
		case "arm64":
			return "linux-arm64", "agent_infini", nil
		}
	case "windows":
		// Only an x64 Windows binary is shipped; it also runs on ARM64.
		return "win32-x64", "agent_infini.exe", nil
	}
	return "", "", fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
}

func currentExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("cannot locate current executable: %w", err)
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return exe, nil
}

func fetchManifest(url string) (*Manifest, error) {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}
	var m Manifest
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return nil, fmt.Errorf("invalid manifest JSON: %w", err)
	}
	if m.Version == "" {
		return nil, fmt.Errorf("manifest missing version field")
	}
	return &m, nil
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: httpTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d for %s", resp.StatusCode, url)
	}

	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// compareVersion compares two dotted numeric versions (MAJOR.MINOR.PATCH).
// Returns -1 if a < b, 0 if equal, 1 if a > b. Non-numeric or missing
// segments are treated as 0.
func compareVersion(a, b string) int {
	pa := splitVersion(a)
	pb := splitVersion(b)
	for i := 0; i < 3; i++ {
		if pa[i] < pb[i] {
			return -1
		}
		if pa[i] > pb[i] {
			return 1
		}
	}
	return 0
}

func splitVersion(v string) [3]int {
	var out [3]int
	// Drop any pre-release/build suffix after '-' or '+'.
	if idx := strings.IndexAny(v, "-+"); idx >= 0 {
		v = v[:idx]
	}
	parts := strings.Split(v, ".")
	for i := 0; i < 3 && i < len(parts); i++ {
		n, _ := strconv.Atoi(strings.TrimSpace(parts[i]))
		out[i] = n
	}
	return out
}

func isDevVersion(v string) bool {
	if v == "" || v == "dev" {
		return true
	}
	return splitVersion(v) == [3]int{0, 0, 0}
}
