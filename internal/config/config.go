package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.yaml.in/yaml/v3"
)

const toolName = "agent_infini"

const (
	KeyServer         = "server"
	KeyAPIKey         = "api-key"
	KeyOutput         = "default-output"
	KeyPreferLanguage = "prefer-language"
	KeyConsole        = "console"
	KeyUserID         = "user-id"
)

const DefaultConsoleURL = "https://api.infinisynapse.cn/api"

var SupportedLanguages = []string{"en", "zh_CN", "ar", "ja", "ko", "ru"}

var defaults = map[string]string{
	KeyOutput:         "json",
	KeyPreferLanguage: "zh_CN",
	KeyConsole:        DefaultConsoleURL,
}

// configFile mirrors the WinClaw config.key / config.json structure.
type configFile struct {
	Global map[string]string `yaml:"global" json:"global"`
}

var cfg configFile

type candidate struct {
	path   string
	format string // "yaml" or "json"
}

// Init walks the WinClaw credential chain and loads the first config file found.
//
// Lookup order (per execute_external_tool_resolver.py):
//  1. <binary_dir>/<tool_basename>.key
//  2. <binary_dir>/<tool_filename>.key  (compat, only when filename differs)
//  3. ~/.<tool_basename>/config.key     (YAML)
//  4. ~/.<tool_basename>/config.json    (JSON)
func Init() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot find home directory: %w", err)
	}

	for _, c := range buildCandidates(home) {
		data, err := os.ReadFile(c.path)
		if err != nil {
			continue
		}
		var parsed configFile
		switch c.format {
		case "yaml":
			if err := yaml.Unmarshal(data, &parsed); err != nil {
				continue
			}
		case "json":
			if err := json.Unmarshal(data, &parsed); err != nil {
				continue
			}
		}
		if parsed.Global == nil {
			parsed.Global = make(map[string]string)
		}
		cfg = parsed
		return nil
	}

	cfg = configFile{Global: make(map[string]string)}
	return nil
}

func buildCandidates(home string) []candidate {
	var out []candidate

	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		filename := filepath.Base(exe)
		basename := strings.TrimSuffix(filename, filepath.Ext(filename))

		// 1. <tool_basename>.key
		out = append(out, candidate{
			path:   filepath.Join(dir, basename+".key"),
			format: "yaml",
		})
		// 2. <tool_filename>.key (compat, only when ext exists)
		if filename != basename {
			out = append(out, candidate{
				path:   filepath.Join(dir, filename+".key"),
				format: "yaml",
			})
		}
	}

	// 3. ~/.<tool_basename>/config.key
	// 4. ~/.<tool_basename>/config.json
	iscDir := filepath.Join(home, "."+toolName)
	out = append(out,
		candidate{path: filepath.Join(iscDir, "config.key"), format: "yaml"},
		candidate{path: filepath.Join(iscDir, "config.json"), format: "json"},
	)
	return out
}

func ConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	return filepath.Join(home, "."+toolName), nil
}

func Save(values map[string]string) error {
	dir, err := ConfigDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	if cfg.Global == nil {
		cfg.Global = make(map[string]string)
	}
	for k, v := range values {
		cfg.Global[k] = v
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	path := filepath.Join(dir, "config.key")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("cannot write config file: %w", err)
	}
	return nil
}

// Set overrides a config value at runtime (e.g. from CLI flags).
func Set(key, value string) {
	if cfg.Global == nil {
		cfg.Global = make(map[string]string)
	}
	cfg.Global[key] = value
}

func Get(key string) string {
	if v, ok := cfg.Global[key]; ok && v != "" {
		return v
	}
	return defaults[key]
}

func GetServer() string         { return Get(KeyServer) }
func GetToken() string          { return Get(KeyAPIKey) }
func GetDefaultOutput() string  { return Get(KeyOutput) }
func GetPreferLanguage() string { return Get(KeyPreferLanguage) }
func GetConsole() string        { return Get(KeyConsole) }
func GetUserID() string         { return Get(KeyUserID) }

// IsInitialized reports whether a config file with a non-empty API key exists.
func IsInitialized() bool {
	home, err := os.UserHomeDir()
	if err != nil {
		return false
	}
	for _, c := range buildCandidates(home) {
		data, err := os.ReadFile(c.path)
		if err != nil {
			continue
		}
		var parsed configFile
		switch c.format {
		case "yaml":
			if err := yaml.Unmarshal(data, &parsed); err != nil {
				continue
			}
		case "json":
			if err := json.Unmarshal(data, &parsed); err != nil {
				continue
			}
		}
		if parsed.Global != nil {
			if key, ok := parsed.Global[KeyAPIKey]; ok && key != "" {
				return true
			}
		}
	}
	return false
}
