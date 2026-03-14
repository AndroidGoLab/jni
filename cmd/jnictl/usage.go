package main

import "github.com/spf13/cobra"

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Usage stats manager operations",
}

var usageIsAppInactiveCmd = &cobra.Command{
	Use:   "is-app-inactive",
	Short: "Check if an app is considered inactive",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		pkg, err := cmd.Flags().GetString("package")
		if err != nil {
			return err
		}
		result, err := grpcClient.Usage.IsAppInactive(ctx, pkg)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

var usageGetStandbyBucketCmd = &cobra.Command{
	Use:   "get-standby-bucket",
	Short: "Get the app standby bucket",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Usage.GetAppStandbyBucket(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

func init() {
	usageIsAppInactiveCmd.Flags().String("package", "", "package name to check")
	_ = usageIsAppInactiveCmd.MarkFlagRequired("package")

	usageCmd.AddCommand(usageIsAppInactiveCmd)
	usageCmd.AddCommand(usageGetStandbyBucketCmd)
	rootCmd.AddCommand(usageCmd)
}
