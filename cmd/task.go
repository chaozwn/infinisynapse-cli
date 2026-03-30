package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/chaozwn/infinisynapse-cli/internal/task"
	"github.com/chaozwn/infinisynapse-cli/internal/types"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks and chat",
	Long: `Manage tasks for multi-turn chat conversations.

Create a new task:
  agent_infini task new --query "Analyze sales data"

Continue a conversation in an existing task:
  agent_infini task -t <taskId> --query "Focus on revenue"

Manage tasks:
  agent_infini task ls
  agent_infini task ls --search "analysis" --page 2
  agent_infini task show <taskId>
  agent_infini task rm <taskId>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query, _ := cmd.Flags().GetString("query")
		if query == "" {
			return cmd.Help()
		}
		if globalTaskID == "" {
			return fmt.Errorf("--task-id / -t is required when using --query")
		}

		_, err := task.RunAskResponse(globalTaskID, query)
		return err
	},
}

var taskNewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create a new task",
	Long: `Send a newTask request to the server and stream the response.

Examples:
  agent_infini task new --query "Analyze sales data"
  agent_infini task new --query "Check stock levels"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query, _ := cmd.Flags().GetString("query")
		if query == "" {
			return fmt.Errorf("--query is required")
		}

		_, err := task.RunNewTask(query)
		return err
	},
}

var taskListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List tasks (paginated)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		search, _ := cmd.Flags().GetString("search")

		params := map[string]string{
			"page":     strconv.Itoa(page),
			"pageSize": strconv.Itoa(pageSize),
			"field":    "updated_at",
			"order":    "desc",
		}
		if search != "" {
			params["task_name"] = search
		}

		data, err := c.Get("/api/ai_task/list", params)
		if err != nil {
			return err
		}

		var result types.TaskListResponse
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("failed to parse task list: %w", err)
		}

		if len(result.Items) == 0 {
			output.PrintSuccess("No tasks found.")
			return nil
		}

		printer := output.NewPrinter(getOutputFormat())
		headers := []string{"ID", "User ID", "Task Name", "Status", "Updated At"}
		rows := make([][]string, 0, len(result.Items))
		for _, item := range result.Items {
			id := item.ID
			if len(id) > 16 {
				id = id[:16] + "..."
			}
			userID := item.UserID
			if len(userID) > 16 {
				userID = userID[:16] + "..."
			}
			name := item.TaskName
			if len(name) > 40 {
				name = name[:40] + "..."
			}
			status := item.Status
			if status == "" {
				status = "-"
			}
			rows = append(rows, []string{id, userID, name, status, item.UpdatedAt})
		}
		printer.PrintTable(headers, rows)

		fmt.Fprintf(cmd.OutOrStdout(), "\nPage %d/%d, %d total tasks\n",
			result.Meta.CurrentPage, result.Meta.TotalPages, result.Meta.TotalItems)

		return nil
	},
}

var taskShowCmd = &cobra.Command{
	Use:   "show [taskId]",
	Short: "Show task details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/api/ai_task/getTaskInfo/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var taskRemoveCmd = &cobra.Command{
	Use:     "rm [taskId...]",
	Aliases: []string{"delete", "remove"},
	Short:   "Delete one or more tasks",
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		body := map[string][]string{"ids": args}
		_, err = c.Post("/api/ai_task/deleteTaskWithId", body)
		if err != nil {
			return err
		}

		output.PrintSuccess("Deleted %d task(s)", len(args))
		return nil
	},
}

var taskStateCmd = &cobra.Command{
	Use:   "state",
	Short: "Get AI state for a task",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		tid := globalTaskID
		if flagTID, _ := cmd.Flags().GetString("task-id"); flagTID != "" {
			tid = flagTID
		}

		params := map[string]string{}
		if tid != "" {
			params["taskId"] = tid
		}

		data, err := c.Get("/api/ai/state", params)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var taskCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a running task",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			return err
		}

		tid := globalTaskID
		if flagTID, _ := cmd.Flags().GetString("task-id"); flagTID != "" {
			tid = flagTID
		}
		if tid == "" {
			return fmt.Errorf("--task-id is required")
		}

		msg := types.WebviewMessage{
			Type:   "cancelTask",
			TaskID: tid,
		}

		data, err := c.Post("/api/ai/message", msg)
		if err != nil {
			return err
		}

		output.PrintSuccess("Task %s cancelled", tid)

		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

func init() {
	taskCmd.Flags().StringP("query", "q", "", "Send a message to continue the conversation (askResponse)")

	taskNewCmd.Flags().StringP("query", "q", "", "Initial message/query (required)")
	_ = taskNewCmd.MarkFlagRequired("query")

	taskListCmd.Flags().Int("page", 1, "Page number")
	taskListCmd.Flags().Int("page-size", 10, "Number of items per page")
	taskListCmd.Flags().String("search", "", "Search tasks by name")

	taskStateCmd.Flags().String("task-id", "", "Get state for specific task")
	taskCancelCmd.Flags().String("task-id", "", "Task ID to cancel")

	taskCmd.AddCommand(taskNewCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskRemoveCmd)
	taskCmd.AddCommand(taskStateCmd)
	taskCmd.AddCommand(taskCancelCmd)

	rootCmd.AddCommand(taskCmd)
}
