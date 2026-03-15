package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var microphoneCmd = &cobra.Command{
	Use:   "microphone",
	Short: "Microphone operations (record)",
}

var microphoneRecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Record audio from the microphone (not yet implemented)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("not yet implemented; requires AudioRecord Java helper")
	},
}

func init() {
	microphoneRecordCmd.Flags().StringP("output", "o", "", "output file (default: stdout)")
	microphoneRecordCmd.Flags().DurationP("duration", "d", 0, "recording duration (0 = until interrupted)")
	microphoneCmd.AddCommand(microphoneRecordCmd)
	rootCmd.AddCommand(microphoneCmd)
}
