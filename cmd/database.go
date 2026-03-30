package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/chaozwn/infinisynapse-cli/internal/types"
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Manage data sources",
}

var dbListCmd = &cobra.Command{
	Use:   "list",
	Short: "List databases",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides("", "")
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		dbType, _ := cmd.Flags().GetString("type")
		page, _ := cmd.Flags().GetString("page")
		size, _ := cmd.Flags().GetString("size")

		params := map[string]string{
			"page":     page,
			"pageSize": size,
		}
		if name != "" {
			params["name"] = name
		}
		if dbType != "" {
			params["type"] = dbType
		}

		data, err := c.Get("/api/ai_database/list", params)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.Print(data,
			[]string{"ID", "Name", "Type", "Enabled", "Description"},
			func(d interface{}) [][]string {
				raw, ok := d.(json.RawMessage)
				if !ok {
					return nil
				}
				var result struct {
					Items []struct {
						ID          string `json:"id"`
						Name        string `json:"name"`
						Type        string `json:"type"`
						Enabled     int    `json:"enabled"`
						Description string `json:"description"`
					} `json:"items"`
				}
				if err := json.Unmarshal(raw, &result); err != nil {
					return nil
				}
				rows := make([][]string, 0, len(result.Items))
				for _, item := range result.Items {
					enabled := "No"
					if item.Enabled == 1 {
						enabled = "Yes"
					}
					desc := item.Description
					if len(desc) > 40 {
						desc = desc[:40] + "..."
					}
					rows = append(rows, []string{item.ID, item.Name, item.Type, enabled, desc})
				}
				return rows
			},
		)
	},
}

var dbGetCmd = &cobra.Command{
	Use:   "get [id]",
	Short: "Get database by ID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides("", "")
		if err != nil {
			return err
		}

		data, err := c.Get(fmt.Sprintf("/api/ai_database/getDatabaseById/%s", args[0]), nil)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var dbAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new database",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides("", "")
		if err != nil {
			return err
		}

		name, _ := cmd.Flags().GetString("name")
		dbType, _ := cmd.Flags().GetString("type")
		config, _ := cmd.Flags().GetString("config")
		desc, _ := cmd.Flags().GetString("description")

		if name == "" || dbType == "" || config == "" {
			return fmt.Errorf("--name, --type, and --config are required")
		}

		body := types.DatabaseAddParams{
			Name:        name,
			Type:        dbType,
			Config:      config,
			Enabled:     1,
			Description: desc,
		}

		data, err := c.Post("/api/ai_database/add", body)
		if err != nil {
			return err
		}

		output.PrintSuccess("Database '%s' added", name)
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

var dbUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a database",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides("", "")
		if err != nil {
			return err
		}

		id, _ := cmd.Flags().GetString("id")
		name, _ := cmd.Flags().GetString("name")
		dbType, _ := cmd.Flags().GetString("type")
		cfg, _ := cmd.Flags().GetString("config")
		desc, _ := cmd.Flags().GetString("description")
		enabled, _ := cmd.Flags().GetInt("enabled")

		if id == "" {
			return fmt.Errorf("--id is required")
		}

		body := types.DatabaseEditParams{
			ID:          id,
			Name:        name,
			Type:        dbType,
			Config:      cfg,
			Enabled:     enabled,
			Description: desc,
		}

		data, err := c.Post("/api/ai_database/update", body)
		if err != nil {
			return err
		}

		output.PrintSuccess("Database '%s' updated", id)
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

var dbDeleteCmd = &cobra.Command{
	Use:   "delete [ids...]",
	Short: "Delete databases",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides("", "")
		if err != nil {
			return err
		}

		body := types.DatabaseDeleteParams{IDs: args}
		data, err := c.Post("/api/ai_database/delete", body)
		if err != nil {
			return err
		}

		output.PrintSuccess("Deleted %d database(s)", len(args))
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

var dbTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test database connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides("", "")
		if err != nil {
			return err
		}

		dbType, _ := cmd.Flags().GetString("type")
		cfg, _ := cmd.Flags().GetString("config")

		if dbType == "" || cfg == "" {
			return fmt.Errorf("--type and --config are required")
		}

		body := types.DatabaseTestConnectionParams{
			Type:   dbType,
			Config: cfg,
		}

		data, err := c.Post("/api/ai_database/testConnection", body)
		if err != nil {
			return err
		}

		output.PrintSuccess("Connection test passed")
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

var dbEnableCmd = &cobra.Command{
	Use:   "enable [ids...]",
	Short: "Enable databases",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return toggleDatabase(args, 1)
	},
}

var dbDisableCmd = &cobra.Command{
	Use:   "disable [ids...]",
	Short: "Disable databases",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return toggleDatabase(args, 0)
	},
}

func toggleDatabase(ids []string, enabled int) error {
	c, err := client.NewWithOverrides("", "")
	if err != nil {
		return err
	}

	body := types.DatabaseEnabledParams{
		IDs:     ids,
		Enabled: enabled,
	}

	data, err := c.Post("/api/ai_database/enabled", body)
	if err != nil {
		return err
	}

	action := "enabled"
	if enabled == 0 {
		action = "disabled"
	}
	output.PrintSuccess("%d database(s) %s", len(ids), action)

	if data != nil {
		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	}
	return nil
}

func init() {
	dbListCmd.Flags().String("name", "", "Filter by name")
	dbListCmd.Flags().String("type", "", "Filter by type")
	dbListCmd.Flags().String("page", "1", "Page number")
	dbListCmd.Flags().String("size", "10", "Page size")

	dbAddCmd.Flags().String("name", "", "Database name (required)")
	dbAddCmd.Flags().String("type", "", "Database type (required)")
	dbAddCmd.Flags().String("config", "", "Database config JSON (required)")
	dbAddCmd.Flags().String("description", "", "Description")

	dbUpdateCmd.Flags().String("id", "", "Database ID (required)")
	dbUpdateCmd.Flags().String("name", "", "Database name")
	dbUpdateCmd.Flags().String("type", "", "Database type")
	dbUpdateCmd.Flags().String("config", "", "Database config JSON")
	dbUpdateCmd.Flags().String("description", "", "Description")
	dbUpdateCmd.Flags().Int("enabled", 1, "Enabled (1=yes, 0=no)")

	dbTestCmd.Flags().String("type", "", "Database type (required)")
	dbTestCmd.Flags().String("config", "", "Database config JSON (required)")

	dbCmd.AddCommand(dbListCmd)
	dbCmd.AddCommand(dbGetCmd)
	dbCmd.AddCommand(dbAddCmd)
	dbCmd.AddCommand(dbUpdateCmd)
	dbCmd.AddCommand(dbDeleteCmd)
	dbCmd.AddCommand(dbTestCmd)
	dbCmd.AddCommand(dbEnableCmd)
	dbCmd.AddCommand(dbDisableCmd)

	rootCmd.AddCommand(dbCmd)
}
