package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

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
	Long: `Manage RAG (Retrieval-Augmented Generation) knowledge bases.

List knowledge bases:
  agent_infini rag ls
  agent_infini rag ls --enabled
  agent_infini rag ls --disabled

Toggle knowledge bases:
  agent_infini rag enable <id> [id...]
  agent_infini rag disable <id> [id...]`,
}

// ---------------------------------------------------------------------------
// rag ls
// ---------------------------------------------------------------------------

var ragListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List RAG knowledge bases (paginated)",
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
			[]string{"ID", "Name", "Enabled", "Source", "Description"},
			func(v interface{}) [][]string {
				r := v.(types.RagListResponse)
				rows := make([][]string, len(r.Items))
				for i, item := range r.Items {
					enabledStr := "no"
					if item.Enabled == 1 {
						enabledStr = "yes"
					}
					rows[i] = []string{
						item.ID,
						item.Name,
						enabledStr,
						item.Source,
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
	Short: "Enable one or more RAG knowledge bases",
	Long: `Enable RAG knowledge bases so they are available for AI tasks.

Multiple IDs can be separated by commas or spaces:
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
	Short: "Disable one or more RAG knowledge bases",
	Long: `Disable RAG knowledge bases so they are not used by AI tasks.

Multiple IDs can be separated by commas or spaces:
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
	ragListCmd.Flags().Int("page", 1, "Page number")
	ragListCmd.Flags().Int("page-size", 10, "Number of items per page")
	ragListCmd.Flags().String("keyword", "", "Filter by name (case-insensitive)")
	ragListCmd.Flags().Bool("enabled", false, "Show only enabled RAGs")
	ragListCmd.Flags().Bool("disabled", false, "Show only disabled RAGs")

	ragCmd.AddCommand(ragListCmd)
	ragCmd.AddCommand(ragEnableCmd)
	ragCmd.AddCommand(ragDisableCmd)

	rootCmd.AddCommand(ragCmd)
}
