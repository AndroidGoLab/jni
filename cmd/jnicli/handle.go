package main

import (
	"github.com/spf13/cobra"
	pb "github.com/AndroidGoLab/jni/proto/handlestore"
)

var handleCmd = &cobra.Command{
	Use:   "handle",
	Short: "Object handle management",
}

var handleReleaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Release a server-side JNI object handle",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()
		client := pb.NewHandleStoreServiceClient(grpcConn)
		handle, err := cmd.Flags().GetInt64("handle")
		if err != nil {
			return err
		}
		resp, err := client.ReleaseHandle(ctx, &pb.ReleaseHandleRequest{Handle: handle})
		if err != nil {
			return err
		}
		return printProtoMessage(resp)
	},
}

func init() {
	handleReleaseCmd.Flags().Int64("handle", 0, "handle ID to release")
	_ = handleReleaseCmd.MarkFlagRequired("handle")

	handleCmd.AddCommand(handleReleaseCmd)
	rootCmd.AddCommand(handleCmd)
}
