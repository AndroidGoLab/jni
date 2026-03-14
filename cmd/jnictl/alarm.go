package main

import "github.com/spf13/cobra"

var alarmCmd = &cobra.Command{
	Use:   "alarm",
	Short: "Alarm manager operations",
}

var alarmCancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "Cancel a scheduled alarm",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		operation, err := cmd.Flags().GetInt64("operation")
		if err != nil {
			return err
		}
		return grpcClient.Alarm.Cancel(ctx, operation)
	},
}

var alarmCanScheduleExactCmd = &cobra.Command{
	Use:   "can-schedule-exact",
	Short: "Check if exact alarms can be scheduled",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Alarm.CanScheduleExactAlarms(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

func init() {
	alarmCancelCmd.Flags().Int64("operation", 0, "operation handle to cancel")
	_ = alarmCancelCmd.MarkFlagRequired("operation")

	alarmCmd.AddCommand(alarmCancelCmd)
	alarmCmd.AddCommand(alarmCanScheduleExactCmd)
	rootCmd.AddCommand(alarmCmd)
}
