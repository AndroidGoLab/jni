package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/xaionaro-go/jni/grpc/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	flagAddr     string
	flagToken    string
	flagInsecure bool
	flagTimeout  time.Duration
	flagFormat   string
	flagCert     string
	flagKey      string
	flagCA       string
)

var grpcConn *grpc.ClientConn
var grpcClient *client.Client

var rootCmd = &cobra.Command{
	Use:   "jnicli",
	Short: "CLI for Android API access over gRPC",
	Long:  "jnicli provides command-line access to Android system services via the go-jni gRPC layer.",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		switch cmd.Name() {
		case "help", "completion", "list-commands":
			return nil
		}

		var opts []grpc.DialOption
		// Allow large messages for binary data transfer (camera photos, videos).
		opts = append(opts,
			grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(128*1024*1024)),
		)

		switch {
		case flagCert != "" && flagKey != "":
			tlsCreds, err := buildMTLSCredentials()
			if err != nil {
				return fmt.Errorf("setting up mTLS: %w", err)
			}
			opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
		case flagInsecure:
			// TLS with InsecureSkipVerify — connects to the server over TLS but
			// does not verify the server certificate. Required for self-signed
			// CA (jniservice generates its own CA). This is NOT plaintext.
			opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(
				&tls.Config{InsecureSkipVerify: true})))
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
	rootCmd.PersistentFlags().StringVar(&flagCert, "cert", "", "path to client certificate PEM file (for mTLS)")
	rootCmd.PersistentFlags().StringVar(&flagKey, "key", "", "path to client private key PEM file (for mTLS)")
	rootCmd.PersistentFlags().StringVar(&flagCA, "ca", "", "path to CA certificate PEM file (for mTLS)")
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

func buildMTLSCredentials() (credentials.TransportCredentials, error) {
	cert, err := tls.LoadX509KeyPair(flagCert, flagKey)
	if err != nil {
		return nil, fmt.Errorf("loading client cert/key: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	if flagCA != "" {
		caPEM, err := os.ReadFile(flagCA)
		if err != nil {
			return nil, fmt.Errorf("reading CA cert: %w", err)
		}

		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caPEM) {
			return nil, fmt.Errorf("failed to parse CA certificate from %s", flagCA)
		}
		tlsCfg.RootCAs = caPool
	}

	if flagInsecure {
		tlsCfg.InsecureSkipVerify = true
	}

	return credentials.NewTLS(tlsCfg), nil
}
