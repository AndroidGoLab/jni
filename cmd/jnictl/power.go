package main

import "github.com/spf13/cobra"

var powerCmd = &cobra.Command{
	Use:   "power",
	Short: "Power manager operations",
}

var powerIsInteractiveCmd = &cobra.Command{
	Use:   "is-interactive",
	Short: "Check if the device is interactive (screen on)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Power.IsInteractive(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

var powerIsPowerSaveModeCmd = &cobra.Command{
	Use:   "is-power-save-mode",
	Short: "Check if power save mode is active",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Power.IsPowerSaveMode(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

func init() {
	powerCmd.AddCommand(powerIsInteractiveCmd)
	powerCmd.AddCommand(powerIsPowerSaveModeCmd)
	rootCmd.AddCommand(powerCmd)
}
