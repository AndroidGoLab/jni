package main

import "github.com/spf13/cobra"

var locationCmd = &cobra.Command{
	Use:   "location",
	Short: "Location manager operations",
}

var locationGetLastKnownCmd = &cobra.Command{
	Use:   "get-last-known",
	Short: "Get the last known location for a provider",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		provider, err := cmd.Flags().GetString("provider")
		if err != nil {
			return err
		}
		loc, err := grpcClient.Location.GetLastKnownLocation(ctx, provider)
		if err != nil {
			return err
		}
		return printResult(loc)
	},
}

var locationIsProviderEnabledCmd = &cobra.Command{
	Use:   "is-provider-enabled",
	Short: "Check if a location provider is enabled",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		provider, err := cmd.Flags().GetString("provider")
		if err != nil {
			return err
		}
		result, err := grpcClient.Location.IsProviderEnabled(ctx, provider)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

func init() {
	locationGetLastKnownCmd.Flags().String("provider", "gps", "location provider (gps, network, passive)")
	locationIsProviderEnabledCmd.Flags().String("provider", "gps", "location provider (gps, network, passive)")

	locationCmd.AddCommand(locationGetLastKnownCmd)
	locationCmd.AddCommand(locationIsProviderEnabledCmd)
	rootCmd.AddCommand(locationCmd)
}
