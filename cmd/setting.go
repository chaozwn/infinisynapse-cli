package cmd

import (
	"fmt"

	"github.com/chaozwn/infinisynapse-cli/internal/client"
	"github.com/chaozwn/infinisynapse-cli/internal/output"
	"github.com/chaozwn/infinisynapse-cli/internal/types"
	"github.com/spf13/cobra"
)

var settingCmd = &cobra.Command{
	Use:   "setting",
	Short: "Manage system settings",
}

var settingGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a setting value by key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		data, err := c.Get("/api/ai_setting/getKey", map[string]string{"key": args[0]})
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var settingSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set a setting value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		params := map[string]string{
			"key":   args[0],
			"value": args[1],
		}

		data, err := c.Post(fmt.Sprintf("/api/ai_setting/setKey?key=%s&value=%s", args[0], args[1]), nil)
		if err != nil {
			return err
		}

		output.PrintSuccess("Setting '%s' = '%s'", params["key"], params["value"])
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

var settingLanguageCmd = &cobra.Command{
	Use:   "language",
	Short: "Get or set preferred language",
}

var settingLanguageGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get preferred language",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		data, err := c.Get("/api/ai_setting/getPreferLanguage", nil)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var settingLanguageSetCmd = &cobra.Command{
	Use:   "set [language]",
	Short: "Set preferred language (e.g. zh-CN, en-US)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		data, err := c.Post(fmt.Sprintf("/api/ai_setting/updatePreferLanguage?language=%s", args[0]), nil)
		if err != nil {
			return err
		}

		output.PrintSuccess("Language set to '%s'", args[0])
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

var settingEngineConfigCmd = &cobra.Command{
	Use:   "engine-config",
	Short: "Get or update engine configuration",
}

var settingEngineConfigGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get engine configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		data, err := c.Get("/api/ai_setting/getEngineConfig", nil)
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

var settingEngineConfigSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Update engine configuration",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		body := types.UpdateEngineConfigParams{
			Config: []types.EngineConfigItem{
				{Key: args[0], Value: args[1]},
			},
		}

		data, err := c.Post("/api/ai_setting/updateEngineConfig", body)
		if err != nil {
			return err
		}

		output.PrintSuccess("Engine config '%s' = '%s'", args[0], args[1])
		if data != nil {
			printer := output.NewPrinter(getOutputFormat())
			return printer.PrintJSON(data)
		}
		return nil
	},
}

var settingModelInfoCmd = &cobra.Command{
	Use:   "model-info [model-id]",
	Short: "Get model information",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := client.NewWithOverrides(flagServer, flagToken)
		if err != nil {
			return err
		}

		data, err := c.Get("/api/ai_setting/getModelInfo", map[string]string{"modelId": args[0]})
		if err != nil {
			return err
		}

		printer := output.NewPrinter(getOutputFormat())
		return printer.PrintJSON(data)
	},
}

func init() {
	settingLanguageCmd.AddCommand(settingLanguageGetCmd)
	settingLanguageCmd.AddCommand(settingLanguageSetCmd)

	settingEngineConfigCmd.AddCommand(settingEngineConfigGetCmd)
	settingEngineConfigCmd.AddCommand(settingEngineConfigSetCmd)

	settingCmd.AddCommand(settingGetCmd)
	settingCmd.AddCommand(settingSetCmd)
	settingCmd.AddCommand(settingLanguageCmd)
	settingCmd.AddCommand(settingEngineConfigCmd)
	settingCmd.AddCommand(settingModelInfoCmd)

	rootCmd.AddCommand(settingCmd)
}
