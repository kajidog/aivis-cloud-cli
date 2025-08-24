package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/kajidog/aivis-cloud-cli/client/payment/domain"
)

var paymentCmd = &cobra.Command{
	Use:   "payment",
	Short: "Payment and billing management commands",
	Long:  "Commands for managing payments, subscriptions, and billing information",
}

var getSubscriptionsCmd = &cobra.Command{
	Use:   "subscriptions",
	Short: "Get subscriptions",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		client := aivisClient

		ctx := context.Background()
		subscriptions, err := client.GetSubscriptions(ctx, limit, offset)
		if err != nil {
			fmt.Printf("Error getting subscriptions: %v\n", err)
			return
		}

		output, _ := json.MarshalIndent(subscriptions, "", "  ")
		fmt.Println(string(output))
	},
}

var getCreditTransactionsCmd = &cobra.Command{
	Use:   "transactions",
	Short: "Get credit transaction history",
	Run: func(cmd *cobra.Command, args []string) {
		transactionType, _ := cmd.Flags().GetString("type")
		status, _ := cmd.Flags().GetString("status")
		startDateStr, _ := cmd.Flags().GetString("start-date")
		endDateStr, _ := cmd.Flags().GetString("end-date")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		var startDate, endDate *time.Time
		if startDateStr != "" {
			if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
				startDate = &parsed
			}
		}
		if endDateStr != "" {
			if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
				endDate = &parsed
			}
		}

		client := aivisClient

		ctx := context.Background()
		response, err := client.GetCreditTransactions(ctx, 
			domain.TransactionType(transactionType), 
			domain.TransactionStatus(status), 
			startDate, endDate, limit, offset)
		if err != nil {
			fmt.Printf("Error getting transactions: %v\n", err)
			return
		}

		output, _ := json.MarshalIndent(response, "", "  ")
		fmt.Println(string(output))
	},
}

var getAPIKeysCmd = &cobra.Command{
	Use:   "api-keys",
	Short: "Get API keys",
	Run: func(cmd *cobra.Command, args []string) {
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		client := aivisClient

		ctx := context.Background()
		apiKeys, err := client.GetAPIKeys(ctx, limit, offset)
		if err != nil {
			fmt.Printf("Error getting API keys: %v\n", err)
			return
		}

		output, _ := json.MarshalIndent(apiKeys, "", "  ")
		fmt.Println(string(output))
	},
}

var createAPIKeyCmd = &cobra.Command{
	Use:   "create-api-key [name]",
	Short: "Create a new API key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		client := aivisClient

		ctx := context.Background()
		apiKey, err := client.CreateAPIKey(ctx, name)
		if err != nil {
			fmt.Printf("Error creating API key: %v\n", err)
			return
		}

		output, _ := json.MarshalIndent(apiKey, "", "  ")
		fmt.Println(string(output))
	},
}

var deleteAPIKeyCmd = &cobra.Command{
	Use:   "delete-api-key [key-id]",
	Short: "Delete an API key",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keyID := args[0]
		client := aivisClient

		ctx := context.Background()
		err := client.DeleteAPIKey(ctx, keyID)
		if err != nil {
			fmt.Printf("Error deleting API key: %v\n", err)
			return
		}

		fmt.Println("API key deleted successfully")
	},
}

var getUsageSummariesCmd = &cobra.Command{
	Use:   "usage",
	Short: "Get usage statistics",
	Run: func(cmd *cobra.Command, args []string) {
		period, _ := cmd.Flags().GetString("period")
		startDateStr, _ := cmd.Flags().GetString("start-date")
		endDateStr, _ := cmd.Flags().GetString("end-date")
		modelID, _ := cmd.Flags().GetString("model-id")

		var startDate, endDate *time.Time
		if startDateStr != "" {
			if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
				startDate = &parsed
			}
		}
		if endDateStr != "" {
			if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
				endDate = &parsed
			}
		}

		client := aivisClient

		ctx := context.Background()
		stats, err := client.GetUsageSummaries(ctx, period, startDate, endDate, modelID)
		if err != nil {
			fmt.Printf("Error getting usage stats: %v\n", err)
			return
		}

		output, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Println(string(output))
	},
}

func init() {
	// Command registration is handled in main.go
	
	paymentCmd.AddCommand(getSubscriptionsCmd)
	paymentCmd.AddCommand(getCreditTransactionsCmd)
	paymentCmd.AddCommand(getAPIKeysCmd)
	paymentCmd.AddCommand(createAPIKeyCmd)
	paymentCmd.AddCommand(deleteAPIKeyCmd)
	paymentCmd.AddCommand(getUsageSummariesCmd)

	// Pagination flags
	getSubscriptionsCmd.Flags().IntP("limit", "l", 20, "Number of results to return")
	getSubscriptionsCmd.Flags().IntP("offset", "o", 0, "Offset for pagination")

	// Transaction filters
	getCreditTransactionsCmd.Flags().String("type", "", "Filter by transaction type (credit, debit, refund)")
	getCreditTransactionsCmd.Flags().String("status", "", "Filter by status (pending, completed, failed, canceled)")
	getCreditTransactionsCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	getCreditTransactionsCmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	getCreditTransactionsCmd.Flags().IntP("limit", "l", 20, "Number of results to return")
	getCreditTransactionsCmd.Flags().IntP("offset", "o", 0, "Offset for pagination")

	// API keys pagination
	getAPIKeysCmd.Flags().IntP("limit", "l", 20, "Number of results to return")
	getAPIKeysCmd.Flags().IntP("offset", "o", 0, "Offset for pagination")

	// Usage stats filters
	getUsageSummariesCmd.Flags().String("period", "month", "Period (day, week, month, year)")
	getUsageSummariesCmd.Flags().String("start-date", "", "Start date (YYYY-MM-DD)")
	getUsageSummariesCmd.Flags().String("end-date", "", "End date (YYYY-MM-DD)")
	getUsageSummariesCmd.Flags().String("model-id", "", "Filter by model ID")
}