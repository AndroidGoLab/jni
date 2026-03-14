package main

import "github.com/spf13/cobra"

var vibratorCmd = &cobra.Command{
	Use:   "vibrator",
	Short: "Vibrator operations",
}

var vibratorHasVibratorCmd = &cobra.Command{
	Use:   "has-vibrator",
	Short: "Check if the device has a vibrator",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Vibrator.HasVibrator(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

var vibratorCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel any ongoing vibration",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		return grpcClient.Vibrator.Cancel(ctx)
	},
}

func init() {
	vibratorCmd.AddCommand(vibratorHasVibratorCmd)
	vibratorCmd.AddCommand(vibratorCancelCmd)
	rootCmd.AddCommand(vibratorCmd)
}
