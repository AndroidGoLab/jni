package main

import "github.com/spf13/cobra"

var irCmd = &cobra.Command{
	Use:   "ir",
	Short: "Infrared emitter operations",
}

var irHasEmitterCmd = &cobra.Command{
	Use:   "has-emitter",
	Short: "Check if the device has an IR emitter",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		result, err := grpcClient.Ir.HasIrEmitter(ctx)
		if err != nil {
			return err
		}
		return printResult(result)
	},
}

var irTransmitCmd = &cobra.Command{
	Use:   "transmit",
	Short: "Transmit an IR pattern",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		freq, err := cmd.Flags().GetInt32("frequency")
		if err != nil {
			return err
		}
		pattern, err := cmd.Flags().GetInt64("pattern")
		if err != nil {
			return err
		}
		return grpcClient.Ir.Transmit(ctx, freq, pattern)
	},
}

func init() {
	irTransmitCmd.Flags().Int32("frequency", 0, "carrier frequency in Hz")
	_ = irTransmitCmd.MarkFlagRequired("frequency")
	irTransmitCmd.Flags().Int64("pattern", 0, "pattern handle")
	_ = irTransmitCmd.MarkFlagRequired("pattern")

	irCmd.AddCommand(irHasEmitterCmd)
	irCmd.AddCommand(irTransmitCmd)
	rootCmd.AddCommand(irCmd)
}
