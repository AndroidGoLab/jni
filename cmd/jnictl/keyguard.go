package main

import "github.com/spf13/cobra"

var keyguardCmd = &cobra.Command{
	Use:   "keyguard",
	Short: "Keyguard manager operations",
}

var keyguardIsLockedCmd = &cobra.Command{
	Use:   "is-locked",
	Short: "Check if the keyguard is locked",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Keyguard.IsKeyguardLocked(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

var keyguardIsSecureCmd = &cobra.Command{
	Use:   "is-secure",
	Short: "Check if the keyguard is secure",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Keyguard.IsKeyguardSecure(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

var keyguardIsDeviceLockedCmd = &cobra.Command{
	Use:   "is-device-locked",
	Short: "Check if the device is locked",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Keyguard.IsDeviceLocked(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

var keyguardIsDeviceSecureCmd = &cobra.Command{
	Use:   "is-device-secure",
	Short: "Check if the device is secure",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Keyguard.IsDeviceSecure(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

func init() {
	keyguardCmd.AddCommand(keyguardIsLockedCmd)
	keyguardCmd.AddCommand(keyguardIsSecureCmd)
	keyguardCmd.AddCommand(keyguardIsDeviceLockedCmd)
	keyguardCmd.AddCommand(keyguardIsDeviceSecureCmd)
	rootCmd.AddCommand(keyguardCmd)
}
