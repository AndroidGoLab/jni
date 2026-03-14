package main

import "github.com/spf13/cobra"

var bluetoothCmd = &cobra.Command{
	Use:   "bluetooth",
	Short: "Bluetooth adapter operations",
}

var bluetoothIsEnabledCmd = &cobra.Command{
	Use:   "is-enabled",
	Short: "Check if Bluetooth is enabled",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Bluetooth.IsEnabled(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

var bluetoothGetNameCmd = &cobra.Command{
	Use:   "get-name",
	Short: "Get the Bluetooth adapter name",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Bluetooth.GetName(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

var bluetoothGetAddressCmd = &cobra.Command{
	Use:   "get-address",
	Short: "Get the Bluetooth adapter MAC address",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Bluetooth.GetAddress(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

var bluetoothGetBondedDevicesCmd = &cobra.Command{
	Use:   "get-bonded-devices",
	Short: "Get bonded devices (returns handle)",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		handle, err := grpcClient.Bluetooth.GetBondedDevices(ctx)
		if err != nil {
			return err
		}
		return printResult(handle)
	},
}

func init() {
	bluetoothCmd.AddCommand(bluetoothIsEnabledCmd)
	bluetoothCmd.AddCommand(bluetoothGetNameCmd)
	bluetoothCmd.AddCommand(bluetoothGetAddressCmd)
	bluetoothCmd.AddCommand(bluetoothGetBondedDevicesCmd)
	rootCmd.AddCommand(bluetoothCmd)
}
