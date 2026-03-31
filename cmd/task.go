package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
	"github.com/chaozwn/infinisynapse-cli/internal/config"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/chaozwn/infinisynapse-cli/internal/task"
	"github.com/chaozwn/infinisynapse-cli/internal/types"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func fetchTaskStatus(c *client.Client, taskID string) string {
	data, err := c.Get(fmt.Sprintf("/api/ai_task/getTaskInfo/%s", taskID), nil)
	if err != nil {
		return ""
	}
	var info struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return ""
	}
	return info.Status
}

type workspaceInfo struct {
	Cwd   string   `json:"cwd"`
	Files []string `json:"files"`
}

type filePreviewResponse struct {
	Content  *string `json:"content"`
	FileType string  `json:"fileType"`
}

func (ws *workspaceInfo) FullPaths() []string {
	if ws == nil {
		return nil
	}
	paths := make([]string, len(ws.Files))
	for i, f := range ws.Files {
		paths[i] = ws.Cwd + "/" + f
	}
	return paths
}

func fetchWorkspaceInfo(c *client.Client, taskID string) *workspaceInfo {
	data, err := c.Get(fmt.Sprintf("/api/ai_task/getTaskWorkspace/%s", taskID), nil)
	if err != nil {
		return nil
	}
	var ws workspaceInfo
	if err := json.Unmarshal(data, &ws); err != nil {
		return nil
	}
	return &ws
}

func enrichStreamResult(result *task.StreamResult) map[string]interface{} {
	res := map[string]interface{}{
		"lastMessage": result.LastMessage,
		"taskId":      result.TaskID,
	}
	c, err := client.New()
	if err != nil {
		return res
	}
	res["status"] = fetchTaskStatus(c, result.TaskID)
	if ws := fetchWorkspaceInfo(c, result.TaskID); ws != nil {
		res["workspace"] = map[string]interface{}{"files": ws.FullPaths()}
	}
	return res
}

// ---------------------------------------------------------------------------
// task (parent)
// ---------------------------------------------------------------------------

var taskCmd = &cobra.Command{
	Use:   "task",
	Short: "Manage tasks and chat",
	Long: `Manage tasks for multi-turn chat conversations.

Create a new task:
  agent_infini task new "Analyze sales data"

Continue a conversation:
  agent_infini task ask <taskId> "Focus on revenue"

Manage tasks:
  agent_infini task ls
  agent_infini task show <taskId>
  agent_infini task rm <taskId>
  agent_infini task cancel <taskId>

View enabled resources:
  agent_infini task context

Workspace files:
  agent_infini task file <taskId>
  agent_infini task preview <taskId> <fileName>
  agent_infini task download <taskId> <fileName> -o ./output/`,
}

// ---------------------------------------------------------------------------
// task new
// ---------------------------------------------------------------------------

var taskNewCmd = &cobra.Command{
	Use:   "new [query]",
	Short: "Create a new task",
	Long: `Send a newTask request to the server and stream the response.

Examples:
  agent_infini task new "Analyze sales data"
  agent_infini task new --query "Check stock levels"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query, _ := cmd.Flags().GetString("query")
		if query == "" && len(args) > 0 {
			query = strings.Join(args, " ")
		}
		if query == "" {
			return fmt.Errorf("query is required: provide as argument or via --query")
		}

		result, err := task.RunNewTask(query, jsonOutput)
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}
		if !jsonOutput {
			output.PrintResult(enrichStreamResult(result), nil)
		}
		return nil
	},
}

// ---------------------------------------------------------------------------
// task ask
// ---------------------------------------------------------------------------

var taskAskCmd = &cobra.Command{
	Use:   "ask <taskId> [query]",
	Short: "Continue a conversation in an existing task",
	Long: `Send an askResponse to continue the conversation in an existing task.

Examples:
  agent_infini task ask <taskId> "Focus on revenue"
  agent_infini task ask <taskId> --query "Show me the trends"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		query, _ := cmd.Flags().GetString("query")
		if query == "" && len(args) > 1 {
			query = strings.Join(args[1:], " ")
		}
		if query == "" {
			return fmt.Errorf("query is required: provide as argument or via --query")
		}

		result, err := task.RunAskResponse(taskID, query, jsonOutput)
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}
		if !jsonOutput {
			output.PrintResult(enrichStreamResult(result), nil)
		}
		return nil
	},
}

// ---------------------------------------------------------------------------
// task ls
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// task show
// ---------------------------------------------------------------------------

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

		taskID := args[0]

		type chanResult struct {
			data json.RawMessage
			err  error
		}

		taskInfoCh := make(chan chanResult, 1)
		uiMsgCh := make(chan chanResult, 1)
		wsCh := make(chan chanResult, 1)

		go func() {
			d, e := c.Get(fmt.Sprintf("/api/ai_task/getTaskInfo/%s", taskID), nil)
			taskInfoCh <- chanResult{d, e}
		}()
		go func() {
			d, e := c.Get("/api/ai_task/getUiMessageById", map[string]string{"id": taskID})
			uiMsgCh <- chanResult{d, e}
		}()
		go func() {
			d, e := c.Get(fmt.Sprintf("/api/ai_task/getTaskWorkspace/%s", taskID), nil)
			wsCh <- chanResult{d, e}
		}()

		tiRes := <-taskInfoCh
		if tiRes.err != nil {
			output.PrintResult(nil, tiRes.err)
			return nil
		}

		var parsed map[string]interface{}
		if err := json.Unmarshal(tiRes.data, &parsed); err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to parse task info: %w", err))
			return nil
		}

		umRes := <-uiMsgCh
		if umRes.err == nil && umRes.data != nil {
			var messages []json.RawMessage
			if err := json.Unmarshal(umRes.data, &messages); err == nil && len(messages) > 0 {
				for i := len(messages) - 1; i >= 0; i-- {
					var msg map[string]interface{}
					if err := json.Unmarshal(messages[i], &msg); err != nil {
						continue
					}
					if msg["ask"] == "completion_result" {
						continue
					}
					parsed["lastInfiniMessage"] = msg
					break
				}
			}
		}

		wsRes := <-wsCh
		if wsRes.err == nil && wsRes.data != nil {
			var ws workspaceInfo
			if err := json.Unmarshal(wsRes.data, &ws); err == nil {
				parsed["workspace"] = map[string]interface{}{"files": ws.FullPaths()}
			}
		}

		output.PrintResult(parsed, nil)
		return nil
	},
}

// ---------------------------------------------------------------------------
// task rm
// ---------------------------------------------------------------------------

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

// ---------------------------------------------------------------------------
// task cancel
// ---------------------------------------------------------------------------

var taskCancelCmd = &cobra.Command{
	Use:   "cancel <taskId>",
	Short: "Cancel a running task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]

		c, err := client.New()
		if err != nil {
			return err
		}

		msg := types.WebviewMessage{
			Type:   "cancelTask",
			TaskID: taskID,
		}

		data, err := c.Post("/api/ai/message", msg)
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

// ---------------------------------------------------------------------------
// task context
// ---------------------------------------------------------------------------

var taskContextCmd = &cobra.Command{
	Use:     "context",
	Aliases: []string{"ctx"},
	Short:   "Show enabled databases and RAGs for tasks",
	Long: `Display all currently enabled databases and RAG knowledge bases,
along with their related resources and actual enabled status.

Examples:
  agent_infini task context
  agent_infini task ctx`,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		params := map[string]string{
			"enabled":  "1",
			"pageSize": "100",
			"field":    "updated_at",
			"order":    "desc",
			"source":   "all",
		}

		type chanResult struct {
			data json.RawMessage
			err  error
		}

		dbCh := make(chan chanResult, 1)
		ragCh := make(chan chanResult, 1)

		go func() {
			d, e := c.Get("/api/ai_database/list", params)
			dbCh <- chanResult{d, e}
		}()
		go func() {
			d, e := c.Get("/api/ai_rag_sdk", params)
			ragCh <- chanResult{d, e}
		}()

		dbRes := <-dbCh
		ragRes := <-ragCh

		if dbRes.err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to fetch databases: %w", dbRes.err))
			return nil
		}
		if ragRes.err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to fetch RAGs: %w", ragRes.err))
			return nil
		}

		var dbResult types.DatabaseListResponse
		if err := json.Unmarshal(dbRes.data, &dbResult); err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to parse database list: %w", err))
			return nil
		}

		var ragResult types.RagListResponse
		if err := json.Unmarshal(ragRes.data, &ragResult); err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to parse RAG list: %w", err))
			return nil
		}

		printer := output.NewPrinter(getOutputFormat())

		if getOutputFormat() == output.FormatTable {
			fmt.Printf("Enabled Databases (%d):\n", len(dbResult.Items))
			dbRows := make([][]string, len(dbResult.Items))
			for i, item := range dbResult.Items {
				related := formatRelatedRags(item.RagList)
				dbRows[i] = []string{item.ID, item.Name, item.Type, item.Source, related}
			}
			printer.PrintTable([]string{"ID", "Name", "Type", "Source", "RelatedRAGs"}, dbRows)

			fmt.Printf("\nEnabled RAGs (%d):\n", len(ragResult.Items))
			ragRows := make([][]string, len(ragResult.Items))
			for i, item := range ragResult.Items {
				related := formatRelatedDBs(item.DatabaseList)
				ragRows[i] = []string{item.ID, item.Name, item.Source, related}
			}
			printer.PrintTable([]string{"ID", "Name", "Source", "RelatedDBs"}, ragRows)
			return nil
		}

		return printer.PrintJSON(map[string]interface{}{
			"databases": dbResult.Items,
			"rags":      ragResult.Items,
			"summary": map[string]int{
				"databases": len(dbResult.Items),
				"rags":      len(ragResult.Items),
			},
		})
	},
}

func formatRelatedRags(rags []types.RelatedRag) string {
	if len(rags) == 0 {
		return "-"
	}
	parts := make([]string, len(rags))
	for i, r := range rags {
		status := "disabled"
		if r.Enabled == 1 {
			status = "enabled"
		}
		parts[i] = fmt.Sprintf("%s (%s)", r.Name, status)
	}
	return strings.Join(parts, ", ")
}

func formatRelatedDBs(dbs []types.RelatedDatabase) string {
	if len(dbs) == 0 {
		return "-"
	}
	parts := make([]string, len(dbs))
	for i, d := range dbs {
		status := "disabled"
		if d.Enabled == 1 {
			status = "enabled"
		}
		parts[i] = fmt.Sprintf("%s (%s)", d.Name, status)
	}
	return strings.Join(parts, ", ")
}

func printContextSummary() {
	config.Init()
	c, err := client.New()
	if err != nil {
		return
	}

	params := map[string]string{
		"enabled":  "1",
		"pageSize": "100",
		"field":    "updated_at",
		"order":    "desc",
		"source":   "all",
	}

	type fetchResult struct {
		data json.RawMessage
		err  error
	}

	dbCh := make(chan fetchResult, 1)
	ragCh := make(chan fetchResult, 1)

	go func() {
		d, e := c.Get("/api/ai_database/list", params)
		dbCh <- fetchResult{d, e}
	}()
	go func() {
		d, e := c.Get("/api/ai_rag_sdk", params)
		ragCh <- fetchResult{d, e}
	}()

	dbRes := <-dbCh
	ragRes := <-ragCh

	var dbItems []types.DatabaseItem
	if dbRes.err == nil {
		var dbResult types.DatabaseListResponse
		if json.Unmarshal(dbRes.data, &dbResult) == nil {
			dbItems = dbResult.Items
		}
	}

	var ragItems []types.RagItem
	if ragRes.err == nil {
		var ragResult types.RagListResponse
		if json.Unmarshal(ragRes.data, &ragResult) == nil {
			ragItems = ragResult.Items
		}
	}

	fmt.Println("\nCurrent Context:")

	if getOutputFormat() == output.FormatJSON {
		out, _ := json.MarshalIndent(map[string]interface{}{
			"databases": dbItems,
			"rags":      ragItems,
			"summary": map[string]int{
				"databases": len(dbItems),
				"rags":      len(ragItems),
			},
		}, "", "  ")
		fmt.Println(string(out))
		return
	}

	if len(dbItems) > 0 {
		names := make([]string, len(dbItems))
		for i, item := range dbItems {
			names[i] = item.Name
		}
		fmt.Printf("  Databases: %d enabled (%s)\n", len(dbItems), strings.Join(names, ", "))
	} else {
		fmt.Println("  Databases: none enabled")
	}
	if len(ragItems) > 0 {
		names := make([]string, len(ragItems))
		for i, item := range ragItems {
			names[i] = item.Name
		}
		fmt.Printf("  RAGs:      %d enabled (%s)\n", len(ragItems), strings.Join(names, ", "))
	} else {
		fmt.Println("  RAGs:      none enabled")
	}
}

// ---------------------------------------------------------------------------
// task file
// ---------------------------------------------------------------------------

var taskFileCmd = &cobra.Command{
	Use:   "file <taskId>",
	Short: "List workspace files for a task",
	Long: `Show all files in the workspace of a task.

Examples:
  agent_infini task file <taskId>`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		ws := fetchWorkspaceInfo(c, args[0])
		if ws == nil {
			output.PrintResult(nil, fmt.Errorf("failed to fetch workspace info"))
			return nil
		}
		output.PrintResult(ws, nil)
		return nil
	},
}

// ---------------------------------------------------------------------------
// task preview
// ---------------------------------------------------------------------------

var taskPreviewCmd = &cobra.Command{
	Use:   "preview <taskId> <fileName>",
	Short: "Preview a workspace file content",
	Long: `Display the content of a workspace file to stdout.

Examples:
  agent_infini task preview <taskId> test_file.txt`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		fileName := args[1]

		c, err := client.New()
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		data, err := c.Post("/api/ai_task/previewFile", map[string]string{
			"taskId":   taskID,
			"fileName": fileName,
		})
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		var resp filePreviewResponse
		if err := json.Unmarshal(data, &resp); err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to parse response: %w", err))
			return nil
		}
		if resp.Content == nil {
			output.PrintResult(nil, fmt.Errorf("file not found: %s", fileName))
			return nil
		}
		fmt.Print(*resp.Content)
		return nil
	},
}

// ---------------------------------------------------------------------------
// task download
// ---------------------------------------------------------------------------

var taskDownloadCmd = &cobra.Command{
	Use:   "download <taskId> <fileName>",
	Short: "Download a workspace file to local disk",
	Long: `Download a file from the task workspace to the local disk.

Examples:
  agent_infini task download <taskId> report.csv
  agent_infini task download <taskId> report.csv -o ./output/`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		taskID := args[0]
		fileName := args[1]
		outDir, _ := cmd.Flags().GetString("output")

		c, err := client.New()
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		endpoint := fmt.Sprintf("/api/tools/storage/downloadTaskFile/%s?path=%s",
			taskID, url.QueryEscape(fileName))
		resp, err := c.RawRequest("GET", endpoint, nil)
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(resp.Body)
			output.PrintResult(nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body)))
			return nil
		}

		dest := outDir
		if info, err := os.Stat(outDir); err == nil && info.IsDir() {
			dest = filepath.Join(outDir, filepath.Base(fileName))
		}
		f, err := os.Create(dest)
		if err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to create file %s: %w", dest, err))
			return nil
		}
		defer f.Close()

		n, err := io.Copy(f, resp.Body)
		if err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to write file: %w", err))
			return nil
		}

		output.PrintResult(map[string]interface{}{"file": dest, "size": n}, nil)
		return nil
	},
}

// ---------------------------------------------------------------------------
// init
// ---------------------------------------------------------------------------

func init() {
	defaultHelp := taskCmd.HelpFunc()
	taskCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		defaultHelp(cmd, args)
		printContextSummary()
	})

	taskNewCmd.Flags().StringP("query", "q", "", "Initial message/query")
	taskAskCmd.Flags().StringP("query", "q", "", "Message to continue the conversation")

	taskListCmd.Flags().Int("page", 1, "Page number")
	taskListCmd.Flags().Int("page-size", 10, "Number of items per page")
	taskListCmd.Flags().String("search", "", "Search tasks by name")

	taskDownloadCmd.Flags().StringP("output", "o", ".", "Output file path or directory")

	taskCmd.AddCommand(taskNewCmd)
	taskCmd.AddCommand(taskAskCmd)
	taskCmd.AddCommand(taskListCmd)
	taskCmd.AddCommand(taskShowCmd)
	taskCmd.AddCommand(taskRemoveCmd)
	taskCmd.AddCommand(taskCancelCmd)
	taskCmd.AddCommand(taskContextCmd)
	taskCmd.AddCommand(taskFileCmd)
	taskCmd.AddCommand(taskPreviewCmd)
	taskCmd.AddCommand(taskDownloadCmd)

	rootCmd.AddCommand(taskCmd)
}
