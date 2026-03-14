package main

import "github.com/spf13/cobra"

var wifiCmd = &cobra.Command{
	Use:   "wifi",
	Short: "WiFi manager operations",
}

var wifiIsEnabledCmd = &cobra.Command{
	Use:   "is-enabled",
	Short: "Check if WiFi is enabled",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Wifi.IsEnabled(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

func init() {
	wifiCmd.AddCommand(wifiIsEnabledCmd)
	rootCmd.AddCommand(wifiCmd)
}
