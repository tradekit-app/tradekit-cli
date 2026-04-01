package cmd

import (
	"fmt"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"github.com/tradekit-dev/tradekit-cli/internal/auth"
	"github.com/tradekit-dev/tradekit-cli/internal/config"
	"github.com/tradekit-dev/tradekit-cli/pkg/types"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authentication commands",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to TradeKit",
	RunE: func(cmd *cobra.Command, args []string) error {
		email, _ := cmd.Flags().GetString("email")
		if email == "" {
			fmt.Print("Email: ")
			fmt.Scanln(&email)
		}
		email = strings.TrimSpace(email)

		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password := string(passwordBytes)

		c := getClient(cmd)
		loginReq := types.LoginRequest{
			Email:    email,
			Password: password,
		}

		resp, err := c.Login(cmd.Context(), loginReq)
		if err != nil {
			errMsg := err.Error()
			// Check if 2FA is required
			if strings.Contains(errMsg, "TWO_FACTOR") || strings.Contains(errMsg, "two_factor") || strings.Contains(errMsg, "2FA") {
				fmt.Print("2FA Code: ")
				var code string
				fmt.Scanln(&code)
				loginReq.TwoFactorCode = strings.TrimSpace(code)

				resp, err = c.Login(cmd.Context(), loginReq)
				if err != nil {
					return fmt.Errorf("login failed: %w", err)
				}
			} else {
				return fmt.Errorf("login failed: %w", err)
			}
		}

		// Save credentials
		store := auth.NewStore(config.Dir())
		creds := &auth.Credentials{
			AccessToken:  resp.AccessToken,
			RefreshToken: resp.RefreshToken,
			ExpiresAt:    resp.ExpiresAt,
			UserID:       resp.User.ID,
			Email:        resp.User.Email,
			Plan:         resp.User.SubscriptionPlan,
		}
		if err := store.Save(creds); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		fmt.Printf("Logged in as %s (%s plan)\n", resp.User.Email, resp.User.SubscriptionPlan)

		if resp.User.SubscriptionPlan != "pro" {
			fmt.Println("\nNote: CLI features require a Pro plan and API key.")
			fmt.Println("  Upgrade: https://tradekit.com.br/pricing")
			fmt.Println("  Create API key: tradekit auth apikey create")
		}
		return nil
	},
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Log out and clear stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := auth.NewStore(config.Dir())

		// Try server-side logout (ignore errors)
		if store.IsLoggedIn() {
			c := getClient(cmd)
			_ = c.Logout(cmd.Context())
		}

		if err := store.Clear(); err != nil {
			return fmt.Errorf("failed to clear credentials: %w", err)
		}
		fmt.Println("Logged out successfully.")
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := auth.NewStore(config.Dir())
		creds, err := store.Load()
		if err != nil {
			return err
		}

		if creds.AccessToken == "" && creds.APIKey == "" {
			fmt.Println("Not logged in. Run: tradekit auth login")
			return nil
		}

		if creds.APIKey != "" {
			fmt.Println("Authenticated via API key")
		}

		// Fetch fresh user info
		c := getClient(cmd)
		user, err := c.GetMe(cmd.Context())
		if err != nil {
			// Fall back to cached info
			fmt.Printf("Email: %s\n", creds.Email)
			fmt.Printf("Plan:  %s\n", creds.Plan)
			if !creds.ExpiresAt.IsZero() {
				fmt.Printf("Token expires: %s\n", creds.ExpiresAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		}

		return printResult(cmd, user)
	},
}

var apikeyCmd = &cobra.Command{
	Use:   "apikey",
	Short: "Manage API keys",
}

var apikeyCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		scopes, _ := cmd.Flags().GetStringSlice("scopes")

		if name == "" {
			fmt.Print("Key name: ")
			fmt.Scanln(&name)
		}
		if len(scopes) == 0 {
			scopes = []string{"read", "write"}
		}

		c := getClient(cmd)
		resp, err := c.CreateAPIKey(cmd.Context(), types.CreateAPIKeyRequest{
			Name:   name,
			Scopes: scopes,
		})
		if err != nil {
			return err
		}

		return printResult(cmd, resp)
	},
}

var apikeyListCmd = &cobra.Command{
	Use:   "list",
	Short: "List API keys",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		keys, err := c.ListAPIKeys(cmd.Context())
		if err != nil {
			return err
		}
		return printResult(cmd, keys)
	},
}

var apikeyRevokeCmd = &cobra.Command{
	Use:   "revoke <id>",
	Short: "Revoke an API key",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := getClient(cmd)
		if err := c.RevokeAPIKey(cmd.Context(), args[0]); err != nil {
			return err
		}
		fmt.Println("API key revoked.")
		return nil
	},
}

func init() {
	loginCmd.Flags().StringP("email", "e", "", "Email address")

	apikeyCreateCmd.Flags().StringP("name", "n", "", "API key name")
	apikeyCreateCmd.Flags().StringSlice("scopes", nil, "Scopes (read, write, trade)")

	apikeyCmd.AddCommand(apikeyCreateCmd)
	apikeyCmd.AddCommand(apikeyListCmd)
	apikeyCmd.AddCommand(apikeyRevokeCmd)

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)
	authCmd.AddCommand(apikeyCmd)
}

