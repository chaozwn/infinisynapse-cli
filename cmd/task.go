package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

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

		_, err := task.RunAskResponse(globalTaskID, query, jsonOutput)
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

		_, err := task.RunNewTask(query, jsonOutput)
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
			output.PrintResult(nil, err)
			return nil
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
			output.PrintResult(nil, err)
			return nil
		}

		var result types.TaskListResponse
		if err := json.Unmarshal(data, &result); err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to parse task list: %w", err))
			return nil
		}

		output.PrintResult(result, nil)
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
			output.PrintResult(nil, err)
			return nil
		}

		data, err := c.Get(fmt.Sprintf("/api/ai_task/getTaskInfo/%s", args[0]), nil)
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		var parsed interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to parse task info: %w", err))
			return nil
		}

		output.PrintResult(parsed, nil)
		return nil
	},
}

var taskRemoveCmd = &cobra.Command{
	Use:   "rm <taskId>[,taskId...]",
	Short: "Delete one or more tasks (comma or space separated)",
	Long: `Delete one or more tasks by ID.

Multiple IDs can be separated by commas or spaces:
  agent_infini task rm id1,id2,id3
  agent_infini task rm id1 id2 id3`,
	Aliases: []string{"delete", "remove"},
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var ids []string
		for _, arg := range args {
			for _, id := range strings.Split(arg, ",") {
				id = strings.TrimSpace(id)
				if id != "" {
					ids = append(ids, id)
				}
			}
		}
		if len(ids) == 0 {
			output.PrintResult(nil, fmt.Errorf("at least one task ID is required"))
			return nil
		}

		c, err := client.New()
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		body := map[string][]string{"ids": ids}
		_, err = c.Post("/api/ai_task/deleteTaskWithId", body)
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		output.PrintResult(map[string]int{"deleted": len(ids)}, nil)
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
