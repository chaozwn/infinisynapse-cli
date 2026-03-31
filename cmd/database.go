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
// db (parent)
// ---------------------------------------------------------------------------

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage databases",
	Long: `Manage database connections for AI tasks.

List databases:
  agent_infini db ls
  agent_infini db ls --enabled
  agent_infini db ls --disabled

Toggle databases:
  agent_infini db enable <id> [id...]
  agent_infini db disable <id> [id...]`,
}

// ---------------------------------------------------------------------------
// db ls
// ---------------------------------------------------------------------------

var dbListCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List databases (paginated)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.New()
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		page, _ := cmd.Flags().GetInt("page")
		pageSize, _ := cmd.Flags().GetInt("page-size")
		name, _ := cmd.Flags().GetString("name")
		dbType, _ := cmd.Flags().GetString("type")
		enabled, _ := cmd.Flags().GetBool("enabled")
		disabled, _ := cmd.Flags().GetBool("disabled")

		params := map[string]string{
			"page":     strconv.Itoa(page),
			"pageSize": strconv.Itoa(pageSize),
			"field":    "updated_at",
			"order":    "desc",
			"source":   "all",
		}
		if name != "" {
			params["name"] = name
		}
		if dbType != "" {
			params["type"] = dbType
		}
		if enabled {
			params["enabled"] = "1"
		} else if disabled {
			params["enabled"] = "0"
		}

		data, err := c.Get("/api/ai_database/list", params)
		if err != nil {
			output.PrintResult(nil, err)
			return nil
		}

		var result types.DatabaseListResponse
		if err := json.Unmarshal(data, &result); err != nil {
			output.PrintResult(nil, fmt.Errorf("failed to parse database list: %w", err))
			return nil
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.Print(
			result,
			[]string{"ID", "Name", "Type", "Enabled", "Source", "RelatedRAGs", "Description"},
			func(v interface{}) [][]string {
				r := v.(types.DatabaseListResponse)
				rows := make([][]string, len(r.Items))
				for i, item := range r.Items {
					enabledStr := "no"
					if item.Enabled == 1 {
						enabledStr = "yes"
					}
					ragNames := make([]string, len(item.RagList))
					for j, rag := range item.RagList {
						ragNames[j] = rag.Name
					}
					rows[i] = []string{
						item.ID,
						item.Name,
						item.Type,
						enabledStr,
						item.Source,
						strings.Join(ragNames, ", "),
						truncate(item.Description, 40),
					}
				}
				return rows
			},
		)
	},
}

// ---------------------------------------------------------------------------
// db enable
// ---------------------------------------------------------------------------

var dbEnableCmd = &cobra.Command{
	Use:   "enable <id> [id...]",
	Short: "Enable one or more databases",
	Long: `Enable databases so they are available for AI tasks.

Multiple IDs can be separated by commas or spaces:
  agent_infini db enable id1 id2 id3
  agent_infini db enable id1,id2,id3`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids := parseIDArgs(args)
		if len(ids) == 0 {
			output.PrintResult(nil, fmt.Errorf("at least one database ID is required"))
			return nil
		}
		return setDatabaseEnabled(ids, 1)
	},
}

// ---------------------------------------------------------------------------
// db disable
// ---------------------------------------------------------------------------

var dbDisableCmd = &cobra.Command{
	Use:   "disable <id> [id...]",
	Short: "Disable one or more databases",
	Long: `Disable databases so they are not used by AI tasks.

Multiple IDs can be separated by commas or spaces:
  agent_infini db disable id1 id2 id3
  agent_infini db disable id1,id2,id3`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ids := parseIDArgs(args)
		if len(ids) == 0 {
			output.PrintResult(nil, fmt.Errorf("at least one database ID is required"))
			return nil
		}
		return setDatabaseEnabled(ids, 0)
	},
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func setDatabaseEnabled(ids []string, enabled int) error {
	c, err := client.New()
	if err != nil {
		output.PrintResult(nil, err)
		return nil
	}

	body := types.DatabaseEnabledRequest{IDs: ids, Enabled: enabled}
	_, err = c.Post("/api/ai_database/enabled", body)
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

func parseIDArgs(args []string) []string {
	var ids []string
	for _, arg := range args {
		for _, id := range strings.Split(arg, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ---------------------------------------------------------------------------
// init
// ---------------------------------------------------------------------------

func init() {
	dbListCmd.Flags().Int("page", 1, "Page number")
	dbListCmd.Flags().Int("page-size", 10, "Number of items per page")
	dbListCmd.Flags().String("name", "", "Filter by database name")
	dbListCmd.Flags().String("type", "", "Filter by database type (mysql, postgres, ...)")
	dbListCmd.Flags().Bool("enabled", false, "Show only enabled databases")
	dbListCmd.Flags().Bool("disabled", false, "Show only disabled databases")

	dbCmd.AddCommand(dbListCmd)
	dbCmd.AddCommand(dbEnableCmd)
	dbCmd.AddCommand(dbDisableCmd)

	rootCmd.AddCommand(dbCmd)
}
