package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/chaozwn/infinisynapse-cli/internal/types"
	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage AI tasks",
}

var taskListCmd = &cobra.Command{
	Use:   "list",
	Short: "List tasks with pagination",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		page, _ := cmd.Flags().GetString("page")
		size, _ := cmd.Flags().GetString("size")
		name, _ := cmd.Flags().GetString("name")
		categoryID, _ := cmd.Flags().GetString("category-id")

		params := map[string]string{
			"page":     page,
			"pageSize": size,
		}
		if name != "" {
			params["task_name"] = name
		}
		if categoryID != "" {
			params["category_id"] = categoryID
		}

		data, err := c.Get("/api/ai_task/list", params)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.Print(data,
			[]string{"ID", "Name", "Status", "Created At"},
			func(d interface{}) [][]string {
				raw, ok := d.(json.RawMessage)
				if !ok {
					return nil
				}
				var result struct {
					Items []struct {
						ID        string `json:"id"`
						Name      string `json:"task_name"`
						Status    string `json:"status"`
						CreatedAt string `json:"createdAt"`
					} `json:"items"`
				}
				if err := json.Unmarshal(raw, &result); err != nil {
					return nil
				}
				rows := make([][]string, 0, len(result.Items))
				for _, item := range result.Items {
					name := item.Name
					if len(name) > 50 {
						name = name[:50] + "..."
					}
					rows = append(rows, []string{item.ID, name, item.Status, item.CreatedAt})
				}
				return rows
			},
		)
	},
}

var taskShowCmd = &cobra.Command{
	Use:   "show [id]",
	Short: "Show task details",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/api/ai_task/showTaskWithId/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var taskInfoCmd = &cobra.Command{
	Use:   "info [id]",
	Short: "Get task metadata",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
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

var taskDeleteCmd = &cobra.Command{
	Use:   "delete [ids...]",
	Short: "Delete tasks by IDs",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		body := types.TaskDeleteParams{IDs: args}
		data, err := c.Post("/api/ai_task/deleteTaskWithId", body)
		if err != nil {
			return err
		}

		output.PrintSuccess("Deleted %d task(s)", len(args))
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

var taskCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a running task",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		taskID, _ := cmd.Flags().GetString("task-id")
		if taskID == "" {
			return fmt.Errorf("--task-id is required")
		}

		data, err := c.Post("/api/ai_task/cancelTask?taskId="+taskID, nil)
		if err != nil {
			return err
		}

		output.PrintSuccess("Task %s cancelled", taskID)
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

// --- Category sub-commands ---

var taskCategoryCmd = &cobra.Command{
	Use:   "category",
	Short: "Manage task categories",
}

var taskCategoryListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all task categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		data, err := c.Get("/api/ai_task/category/getAllCategories", nil)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.Print(data,
			[]string{"ID", "Name"},
			func(d interface{}) [][]string {
				raw, ok := d.(json.RawMessage)
				if !ok {
					return nil
				}
				var items []struct {
					ID   string `json:"id"`
					Name string `json:"category_name"`
				}
				if err := json.Unmarshal(raw, &items); err != nil {
					return nil
				}
				rows := make([][]string, 0, len(items))
				for _, item := range items {
					rows = append(rows, []string{item.ID, item.Name})
				}
				return rows
			},
		)
	},
}

var taskCategoryAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a task category",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		body := types.TaskCategoryAddParams{CategoryName: args[0]}
		data, err := c.Post("/api/ai_task/category/add", body)
		if err != nil {
			return err
		}

		output.PrintSuccess("Category '%s' added", args[0])
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

var taskCategoryDeleteCmd = &cobra.Command{
	Use:   "delete [ids...]",
	Short: "Delete task categories",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		body := types.TaskCategoryDeleteParams{IDs: args}
		data, err := c.Post("/api/ai_task/category/delete", body)
		if err != nil {
			return err
		}

		output.PrintSuccess("Deleted %d category(ies)", len(args))
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

func init() {
	taskListCmd.Flags().String("page", "1", "Page number")
	taskListCmd.Flags().String("size", "10", "Page size")
	taskListCmd.Flags().String("name", "", "Filter by task name")
	taskListCmd.Flags().String("category-id", "", "Filter by category ID")

	taskCancelCmd.Flags().String("task-id", "", "Task ID to cancel (required)")

	taskCategoryCmd.AddCommand(taskCategoryListCmd)
	taskCategoryCmd.AddCommand(taskCategoryAddCmd)
	taskCategoryCmd.AddCommand(taskCategoryDeleteCmd)

	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskInfoCmd)
	taskCmd.AddCommand(taskDeleteCmd)
	taskCmd.AddCommand(taskCancelCmd)
	taskCmd.AddCommand(taskCategoryCmd)

	rootCmd.AddCommand(taskCmd)
}
