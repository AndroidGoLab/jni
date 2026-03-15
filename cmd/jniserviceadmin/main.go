package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/xaionaro-go/jni/grpc/server/acl"
)

var (
	flagDB string
	store  *acl.Store
)

var rootCmd = &cobra.Command{
	Use:   "jniserviceadmin",
	Short: "Server-side ACL management for jniservice",
	Long:  "jniserviceadmin provides direct SQLite-based management of clients, grants, and access requests for jniservice.",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		switch cmd.Name() {
		case "help", "completion":
			return nil
		}

		s, err := acl.OpenStore(flagDB)
		if err != nil {
			return fmt.Errorf("opening ACL store at %s: %w", flagDB, err)
		}
		store = s
		return nil
	},
	PersistentPostRunE: func(_ *cobra.Command, _ []string) error {
		if store != nil {
			return store.Close()
		}
		return nil
	},
}

// clients

var clientsCmd = &cobra.Command{
	Use:   "clients",
	Short: "Manage registered clients",
}

var clientsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all registered clients",
	RunE: func(_ *cobra.Command, _ []string) error {
		clients, err := store.ListClients()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "CLIENT_ID\tFINGERPRINT\tREGISTERED_AT")
		for _, c := range clients {
			fmt.Fprintf(w, "%s\t%s\t%s\n", c.ClientID, c.Fingerprint, c.RegisteredAt.Format("2006-01-02T15:04:05Z"))
		}
		return w.Flush()
	},
}

var clientsRevokeCmd = &cobra.Command{
	Use:   "revoke <client-id>",
	Short: "Revoke a client and all its grants",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		clientID := args[0]
		if err := store.RevokeClient(clientID); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Revoked client %q and all associated grants.\n", clientID)
		return nil
	},
}

// grants

var (
	flagGrantsClient string
)

var grantsCmd = &cobra.Command{
	Use:   "grants",
	Short: "Manage method-access grants",
}

var grantsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List grants (optionally filtered by client)",
	RunE: func(_ *cobra.Command, _ []string) error {
		grants, err := store.ListGrants(flagGrantsClient)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "CLIENT_ID\tMETHOD_PATTERN\tGRANTED_AT\tGRANTED_BY")
		for _, g := range grants {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", g.ClientID, g.MethodPattern, g.GrantedAt.Format("2006-01-02T15:04:05Z"), g.GrantedBy)
		}
		return w.Flush()
	},
}

var grantsApproveCmd = &cobra.Command{
	Use:   "approve <client-id> <method-pattern>",
	Short: "Grant a client access to a method pattern",
	Args:  cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		clientID := args[0]
		pattern := args[1]
		if err := store.Grant(clientID, pattern, "cli:admin"); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Granted %q to client %q.\n", pattern, clientID)
		return nil
	},
}

var grantsRevokeCmd = &cobra.Command{
	Use:   "revoke <client-id> <method-pattern>",
	Short: "Revoke a client's access to a method pattern",
	Args:  cobra.ExactArgs(2),
	RunE: func(_ *cobra.Command, args []string) error {
		clientID := args[0]
		pattern := args[1]
		if err := store.Revoke(clientID, pattern); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Revoked %q from client %q.\n", pattern, clientID)
		return nil
	},
}

// requests

var requestsCmd = &cobra.Command{
	Use:   "requests",
	Short: "Manage pending access requests",
}

var requestsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List pending access requests",
	RunE: func(_ *cobra.Command, _ []string) error {
		requests, err := store.ListPendingRequests()
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCLIENT_ID\tMETHODS\tREQUESTED_AT")
		for _, r := range requests {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", r.ID, r.ClientID, strings.Join(r.Methods, ","), r.RequestedAt.Format("2006-01-02T15:04:05Z"))
		}
		return w.Flush()
	},
}

var requestsApproveCmd = &cobra.Command{
	Use:   "approve <request-id>",
	Short: "Approve a pending access request",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("parsing request-id %q: %w", args[0], err)
		}
		if err := store.ApproveRequest(id, "cli:admin"); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Approved request %d.\n", id)
		return nil
	},
}

var requestsDenyCmd = &cobra.Command{
	Use:   "deny <request-id>",
	Short: "Deny a pending access request",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("parsing request-id %q: %w", args[0], err)
		}
		if err := store.DenyRequest(id); err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "Denied request %d.\n", id)
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagDB, "db", "/data/local/tmp/jniservice.db", "path to SQLite database")

	grantsListCmd.Flags().StringVar(&flagGrantsClient, "client", "", "filter grants by client ID")

	clientsCmd.AddCommand(clientsListCmd, clientsRevokeCmd)
	grantsCmd.AddCommand(grantsListCmd, grantsApproveCmd, grantsRevokeCmd)
	requestsCmd.AddCommand(requestsListCmd, requestsApproveCmd, requestsDenyCmd)

	rootCmd.AddCommand(clientsCmd, grantsCmd, requestsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
