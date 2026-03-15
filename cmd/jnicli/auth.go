package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/xaionaro-go/jni/grpc/server/certauth"
	authpb "github.com/xaionaro-go/jni/proto/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "client registration and permission management",
}

var authRegisterCmd = &cobra.Command{
	Use:   "register",
	Short: "register a new client by generating a key and CSR",
	RunE: func(cmd *cobra.Command, args []string) error {
		cn, err := cmd.Flags().GetString("cn")
		if err != nil {
			return fmt.Errorf("reading --cn flag: %w", err)
		}
		if cn == "" {
			return fmt.Errorf("--cn is required")
		}

		keyOut, err := cmd.Flags().GetString("key-out")
		if err != nil {
			return fmt.Errorf("reading --key-out flag: %w", err)
		}

		certOut, err := cmd.Flags().GetString("cert-out")
		if err != nil {
			return fmt.Errorf("reading --cert-out flag: %w", err)
		}

		caOut, err := cmd.Flags().GetString("ca-out")
		if err != nil {
			return fmt.Errorf("reading --ca-out flag: %w", err)
		}

		csrPEM, keyPEM, err := certauth.GenerateCSR(cn)
		if err != nil {
			return fmt.Errorf("generating CSR: %w", err)
		}

		ctx, cancel := requestContext(cmd)
		defer cancel()

		client := authpb.NewAuthServiceClient(grpcConn)
		resp, err := client.Register(ctx, &authpb.RegisterRequest{
			CsrPem: string(csrPEM),
		})
		if err != nil {
			return fmt.Errorf("registering: %w", err)
		}

		if err := os.WriteFile(keyOut, keyPEM, 0600); err != nil {
			return fmt.Errorf("writing key to %s: %w", keyOut, err)
		}

		if err := os.WriteFile(certOut, []byte(resp.GetClientCertPem()), 0644); err != nil {
			return fmt.Errorf("writing cert to %s: %w", certOut, err)
		}

		if err := os.WriteFile(caOut, []byte(resp.GetCaCertPem()), 0644); err != nil {
			return fmt.Errorf("writing CA cert to %s: %w", caOut, err)
		}

		fmt.Printf("registered successfully:\n  key:  %s\n  cert: %s\n  ca:   %s\n", keyOut, certOut, caOut)
		return nil
	},
}

var authRequestPermissionCmd = &cobra.Command{
	Use:   "request-permission",
	Short: "request permission to call specific gRPC methods",
	RunE: func(cmd *cobra.Command, args []string) error {
		methodsRaw, err := cmd.Flags().GetString("methods")
		if err != nil {
			return fmt.Errorf("reading --methods flag: %w", err)
		}

		methods := strings.Split(methodsRaw, ",")
		for i := range methods {
			methods[i] = strings.TrimSpace(methods[i])
		}

		ctx, cancel := requestContext(cmd)
		defer cancel()

		client := authpb.NewAuthServiceClient(grpcConn)
		resp, err := client.RequestPermission(ctx, &authpb.RequestPermissionRequest{
			Methods: methods,
		})
		if err != nil {
			return err
		}

		return printProtoMessage(resp)
	},
}

var authListPermissionsCmd = &cobra.Command{
	Use:   "list-permissions",
	Short: "list methods this client is permitted to call",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := requestContext(cmd)
		defer cancel()

		client := authpb.NewAuthServiceClient(grpcConn)
		resp, err := client.ListMyPermissions(ctx, &authpb.ListMyPermissionsRequest{})
		if err != nil {
			return err
		}

		return printProtoMessage(resp)
	},
}

func init() {
	authRegisterCmd.Flags().String("cn", "", "common name for the client certificate (required)")
	authRegisterCmd.Flags().String("key-out", "client.key", "path to write the private key")
	authRegisterCmd.Flags().String("cert-out", "client.crt", "path to write the signed certificate")
	authRegisterCmd.Flags().String("ca-out", "ca.crt", "path to write the CA certificate")

	authRequestPermissionCmd.Flags().String("methods", "", "comma-separated gRPC method patterns to request")

	authCmd.AddCommand(authRegisterCmd)
	authCmd.AddCommand(authRequestPermissionCmd)
	authCmd.AddCommand(authListPermissionsCmd)
	rootCmd.AddCommand(authCmd)
}
