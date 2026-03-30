package cmd

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/chaozwn/infinisynapse-cli/internal/session"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage chat sessions",
	Long: `Manage session aliases used for multi-turn chat conversations.

Sessions automatically save and restore the task ID between 'agent_infini chat' calls,
enabling seamless multi-turn workflows without manually tracking task IDs.`,
}

var sessionListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List all sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := session.List()
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			output.PrintSuccess("No sessions found.")
			return nil
		}

		printer := output.NewPrinter(getOutputFormat())
		headers := []string{"Name", "Task ID", "CWD", "Updated At"}
		rows := make([][]string, 0, len(sessions))
		for _, s := range sessions {
			taskID := s.TaskID
			if len(taskID) > 16 {
				taskID = taskID[:16] + "..."
			}
			cwd := s.CWD
			if len(cwd) > 40 {
				cwd = "..." + cwd[len(cwd)-37:]
			}
			rows = append(rows, []string{s.Name, taskID, cwd, s.UpdatedAt.Format("2006-01-02 15:04:05")})
		}
		printer.PrintTable(headers, rows)
		return nil
	},
}

var sessionShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show session details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sess, err := session.Load(args[0])
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(sess)
	},
}

var sessionRemoveCmd = &cobra.Command{
	Use:     "rm [name]",
	Aliases: []string{"delete", "remove"},
	Short:   "Delete a session",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := session.Delete(args[0]); err != nil {
			return err
		}
		output.PrintSuccess("Session '%s' deleted", args[0])
		return nil
	},
}

var sessionUseCmd = &cobra.Command{
	Use:   "use [name]",
	Short: "Create (if needed) and set a session as current",
	Long: `Set the named session as the current active session.
If the session does not exist, it will be created.
Subsequent 'chat' calls without --session will use this session automatically.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		if !session.Exists(name) {
			taskID := strconv.FormatInt(time.Now().UnixMilli(), 10)
			cwd, _ := os.Getwd()
			if err := session.Save(name, taskID, "", cwd); err != nil {
				return fmt.Errorf("failed to create session '%s': %w", name, err)
			}
			output.PrintSuccess("Session '%s' created (taskId: %s)", name, taskID)
		}

		if err := session.SetCurrent(name); err != nil {
			return fmt.Errorf("failed to set current session: %w", err)
		}
		output.PrintSuccess("Current session set to '%s'", name)
		return nil
	},
}

var sessionCurrentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show the current active session",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, err := session.GetCurrent()
		if err != nil {
			return fmt.Errorf("failed to read current session: %w", err)
		}
		if name == "" {
			output.PrintSuccess("No current session set. Use 'agent_infini session use <name>' to set one.")
			return nil
		}

		sess, err := session.Load(name)
		if err != nil {
			return fmt.Errorf("current session '%s' is set but cannot be loaded: %w", name, err)
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(sess)
	},
}

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionShowCmd)
	sessionCmd.AddCommand(sessionRemoveCmd)
	sessionCmd.AddCommand(sessionUseCmd)
	sessionCmd.AddCommand(sessionCurrentCmd)

	rootCmd.AddCommand(sessionCmd)
}

func resolveSessionName(sessionName string) string {
	if sessionName != "" {
		return sessionName
	}
	cur, _ := session.GetCurrent()
	return cur
}

func resolveTaskIDFromSession(sessionName, explicitTaskID string) string {
	if explicitTaskID != "" {
		return explicitTaskID
	}
	name := resolveSessionName(sessionName)
	if name == "" {
		return ""
	}
	sess, err := session.Load(name)
	if err != nil {
		return ""
	}
	return sess.TaskID
}

func saveSession(sessionName, taskID, connID string) {
	name := resolveSessionName(sessionName)
	if name == "" || taskID == "" {
		return
	}
	cwd, _ := os.Getwd()
	_ = session.Save(name, taskID, connID, cwd)
}
