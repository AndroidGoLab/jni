package main

import "github.com/spf13/cobra"

var jobCmd = &cobra.Command{
	Use:   "job",
	Short: "Job scheduler operations",
}

var jobCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a scheduled job",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		jobId, err := cmd.Flags().GetInt32("id")
		if err != nil {
			return err
		}
		return grpcClient.Job.Cancel(ctx, jobId)
	},
}

var jobCancelAllCmd = &cobra.Command{
	Use:   "cancel-all",
	Short: "Cancel all scheduled jobs",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		return grpcClient.Job.CancelAll(ctx)
	},
}

func init() {
	jobCancelCmd.Flags().Int32("id", 0, "job ID to cancel")
	_ = jobCancelCmd.MarkFlagRequired("id")

	jobCmd.AddCommand(jobCancelCmd)
	jobCmd.AddCommand(jobCancelAllCmd)
	rootCmd.AddCommand(jobCmd)
}
