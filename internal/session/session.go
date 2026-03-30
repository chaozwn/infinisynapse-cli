package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type SessionData struct {
	Name      string    `json:"name"`
	TaskID    string    `json:"taskId"`
	ConnID    string    `json:"connId,omitempty"`
	CWD       string    `json:"cwd"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func sessionsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}
	dir := filepath.Join(home, ".agent_infini", "sessions")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("cannot create sessions directory: %w", err)
	}
	return dir, nil
}

func sessionFilePath(name string) (string, error) {
	dir, err := sessionsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".json"), nil
}

func Load(name string) (*SessionData, error) {
	fp, err := sessionFilePath(name)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("session '%s' not found", name)
		}
		return nil, fmt.Errorf("failed to read session '%s': %w", name, err)
	}

	var sess SessionData
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("failed to parse session '%s': %w", name, err)
	}
	return &sess, nil
}

func Save(name, taskID, connID, cwd string) error {
	fp, err := sessionFilePath(name)
	if err != nil {
		return err
	}

	now := time.Now()
	sess := &SessionData{
		Name:      name,
		TaskID:    taskID,
		ConnID:    connID,
		CWD:       cwd,
		UpdatedAt: now,
	}

	existing, loadErr := Load(name)
	if loadErr == nil && existing != nil {
		sess.CreatedAt = existing.CreatedAt
	} else {
		sess.CreatedAt = now
	}

	data, err := json.MarshalIndent(sess, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	return os.WriteFile(fp, data, 0644)
}

func List() ([]SessionData, error) {
	dir, err := sessionsDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read sessions directory: %w", err)
	}

	var sessions []SessionData
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".json")
		sess, err := Load(name)
		if err != nil {
			continue
		}
		sessions = append(sessions, *sess)
	}

	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].UpdatedAt.After(sessions[j].UpdatedAt)
	})

	return sessions, nil
}

func Delete(name string) error {
	fp, err := sessionFilePath(name)
	if err != nil {
		return err
	}

	if err := os.Remove(fp); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("session '%s' not found", name)
		}
		return fmt.Errorf("failed to delete session '%s': %w", name, err)
	}

	if cur, _ := GetCurrent(); cur == name {
		_ = ClearCurrent()
	}
	return nil
}

func Exists(name string) bool {
	fp, err := sessionFilePath(name)
	if err != nil {
		return false
	}
	_, err = os.Stat(fp)
	return err == nil
}

func currentFilePath() (string, error) {
	dir, err := sessionsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, ".current"), nil
}

func SetCurrent(name string) error {
	fp, err := currentFilePath()
	if err != nil {
		return err
	}
	return os.WriteFile(fp, []byte(name), 0644)
}

func GetCurrent() (string, error) {
	fp, err := currentFilePath()
	if err != nil {
		return "", err
	}
	data, err := os.ReadFile(fp)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}

func ClearCurrent() error {
	fp, err := currentFilePath()
	if err != nil {
		return err
	}
	if err := os.Remove(fp); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
