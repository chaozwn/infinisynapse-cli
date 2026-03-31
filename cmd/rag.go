package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/chaozwn/infinisynapse-cli/internal/types"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// rag (parent)
// ---------------------------------------------------------------------------

var ragCmd = &cobra.Command{
	Use:   "rag",
	Short: "Manage RAG knowledge bases",
	Long: `Manage RAG (Retrieval-Augmented Generation) knowledge bases that AI tasks
can use for context-aware retrieval.

List knowledge bases:
  agent_infini rag ls
  agent_infini rag ls --enabled
  agent_infini rag ls --keyword sales

Toggle availability:
  agent_infini rag enable <id> [id...]
  agent_infini rag disable <id> [id...]`,
}

// ---------------------------------------------------------------------------
// rag ls
// ---------------------------------------------------------------------------

var ragListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List registered RAG knowledge bases",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		keyword, _ := cmd.Flags().GetString("keyword")
		enabled, _ := cmd.Flags().GetBool("enabled")
		disabled, _ := cmd.Flags().GetBool("disabled")

		params := map[string]string{
			"page":     strconv.Itoa(page),
			"pageSize": strconv.Itoa(pageSize),
			"field":    "updated_at",
			"order":    "desc",
			"source":   "all",
		}
		if keyword != "" {
			params["keyword"] = keyword
		}
		if enabled {
			params["enabled"] = "1"
		} else if disabled {
			params["enabled"] = "0"
		}

		data, err := c.Get("/api/ai_rag_sdk", params)
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		var result types.RagListResponse
		if err := json.Unmarshal(data, &result); err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to parse RAG list: %w", err))
			return nil
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.Print(
			result,
			[]string{"ID", "Name", "Enabled", "Source", "RelatedDBs", "Description"},
			func(v interface{}) [][]string {
				r := v.(types.RagListResponse)
				rows := make([][]string, len(r.Items))
				for i, item := range r.Items {
					enabledStr := "no"
					if item.Enabled == 1 {
						enabledStr = "yes"
					}
					dbNames := make([]string, len(item.DatabaseList))
					for j, db := range item.DatabaseList {
						dbNames[j] = db.Name
					}
					rows[i] = []string{
						item.ID,
						item.Name,
						enabledStr,
						item.Source,
						strings.Join(dbNames, ", "),
						truncate(item.Description, 40),
					}
				}
				return rows
			},
		)
	},
}

// ---------------------------------------------------------------------------
// rag enable
// ---------------------------------------------------------------------------

var ragEnableCmd = &cobra.Command{
	Use:   "enable <id> [id...]",
	Short: "Enable RAG knowledge bases for AI task access",
	Long: `Mark RAG knowledge bases as available so AI tasks can retrieve context from them.

Pass one or more IDs separated by spaces or commas:
  agent_infini rag enable uuid-1 uuid-2
  agent_infini rag enable uuid-1,uuid-2`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids := parseIDArgs(args)
		if len(ids) == 0 {
			output.PrintResult(nil, fmt.Errorf("at least one RAG ID is required"))
			return nil
		}
		return setRagEnabled(ids, 1)
	},
}

// ---------------------------------------------------------------------------
// rag disable
// ---------------------------------------------------------------------------

var ragDisableCmd = &cobra.Command{
	Use:   "disable <id> [id...]",
	Short: "Disable RAG knowledge bases from AI task access",
	Long: `Mark RAG knowledge bases as unavailable so AI tasks will no longer retrieve context from them.

Pass one or more IDs separated by spaces or commas:
  agent_infini rag disable uuid-1 uuid-2
  agent_infini rag disable uuid-1,uuid-2`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids := parseIDArgs(args)
		if len(ids) == 0 {
			output.PrintResult(nil, fmt.Errorf("at least one RAG ID is required"))
			return nil
		}
		return setRagEnabled(ids, 0)
	},
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func setRagEnabled(ids []string, enabled int) error {
	c, err := client.New()
	if err != nil {
		output.PrintResult(nil, err)
		return nil
	}

	body := types.RagEnabledRequest{IDs: ids, Enabled: enabled}
	_, err = c.Post("/api/ai_rag_sdk/enabled", body)
	if err != nil {
		output.PrintResult(nil, err)
		return nil
	}

	action := "enabled"
	if enabled == 0 {
		action = "disabled"
	}
	output.PrintResult(map[string]interface{}{
		"action": action,
		"count":  len(ids),
		"ids":    ids,
	}, nil)
	return nil
}

// ---------------------------------------------------------------------------
// init
// ---------------------------------------------------------------------------

func init() {
	ragListCmd.Flags().Int("page", 1, "Page number (1-based)")
	ragListCmd.Flags().Int("page-size", 10, "Items per page")
	ragListCmd.Flags().String("keyword", "", "Filter by name (substring match, case-insensitive)")
	ragListCmd.Flags().Bool("enabled", false, "Show only enabled RAGs")
	ragListCmd.Flags().Bool("disabled", false, "Show only disabled RAGs")

	ragCmd.AddCommand(ragListCmd)
	ragCmd.AddCommand(ragEnableCmd)
	ragCmd.AddCommand(ragDisableCmd)

	rootCmd.AddCommand(ragCmd)
}
