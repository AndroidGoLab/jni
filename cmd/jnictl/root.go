package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/xaionaro-go/jni/grpc/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	flagAddr     string
	flagToken    string
	flagInsecure bool
	flagTimeout  time.Duration
	flagFormat   string
)

var grpcConn *grpc.ClientConn
var grpcClient *client.Client

var rootCmd = &cobra.Command{
	Use:   "jnictl",
	Short: "CLI for Android API access over gRPC",
	Long:  "jnictl provides command-line access to Android system services via the go-jni gRPC layer.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		switch cmd.Name() {
		case "help", "completion", "list-commands":
			return nil
		}

		var opts []grpc.DialOption
		if flagInsecure {
			opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
		}
		if flagToken != "" {
			opts = append(opts, grpc.WithPerRPCCredentials(tokenCredentials{token: flagToken}))
		}

		conn, err := grpc.NewClient(flagAddr, opts...)
		if err != nil {
			return fmt.Errorf("connect to %s: %w", flagAddr, err)
		}
		grpcConn = conn
		grpcClient = client.NewClient(conn)
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if grpcConn != nil {
			return grpcConn.Close()
		}
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagAddr, "addr", "a", "localhost:50051", "gRPC server address")
	rootCmd.PersistentFlags().StringVarP(&flagToken, "token", "t", "", "authentication token")
	rootCmd.PersistentFlags().BoolVar(&flagInsecure, "insecure", false, "use insecure connection (no TLS)")
	rootCmd.PersistentFlags().DurationVar(&flagTimeout, "timeout", 10*time.Second, "request timeout")
	rootCmd.PersistentFlags().StringVarP(&flagFormat, "format", "f", "json", "output format (json|text)")
}

func requestContext(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	return context.WithTimeout(cmd.Context(), flagTimeout)
}

// tokenCredentials implements grpc.PerRPCCredentials for bearer token auth.
type tokenCredentials struct {
	token string
}

func (t tokenCredentials) GetRequestMetadata(
	_ context.Context,
	_ ...string,
) (map[string]string, error) {
	return map[string]string{"authorization": "Bearer " + t.token}, nil
}

func (t tokenCredentials) RequireTransportSecurity() bool {
	return false
}
