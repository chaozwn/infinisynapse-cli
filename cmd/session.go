package cmd

import (
	"os"

	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/chaozwn/infinisynapse-cli/internal/session"
	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage chat sessions",
	Long: `Manage session aliases used for multi-turn chat conversations.

Sessions automatically save and restore the task ID between 'isc chat' calls,
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

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionShowCmd)
	sessionCmd.AddCommand(sessionRemoveCmd)

	rootCmd.AddCommand(sessionCmd)
}

func resolveTaskIDFromSession(sessionName, explicitTaskID string) string {
	if explicitTaskID != "" {
		return explicitTaskID
	}
	if sessionName == "" {
		return ""
	}
	sess, err := session.Load(sessionName)
	if err != nil {
		return ""
	}
	return sess.TaskID
}

func saveSession(sessionName, taskID, connID string) {
	if sessionName == "" || taskID == "" {
		return
	}
	cwd, _ := os.Getwd()
	_ = session.Save(sessionName, taskID, connID, cwd)
}
