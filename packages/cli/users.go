package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "User management commands",
	Long:  "Commands for managing users and viewing profiles",
}

var getUserMeCmd = &cobra.Command{
	Use:   "me",
	Short: "Get your account information",
	Run: func(cmd *cobra.Command, args []string) {
		client := aivisClient

		ctx := context.Background()
		userProfile, err := client.GetMe(ctx)
		if err != nil {
			fmt.Printf("Error getting user profile: %v\n", err)
			return
		}

		output, _ := json.MarshalIndent(userProfile, "", "  ")
		fmt.Println(string(output))
	},
}

var getUserByHandleCmd = &cobra.Command{
	Use:   "handle [handle]",
	Short: "Get user profile by handle",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		handle := args[0]
		client := aivisClient

		ctx := context.Background()
		userProfile, err := client.GetUserByHandle(ctx, handle)
		if err != nil {
			fmt.Printf("Error getting user: %v\n", err)
			return
		}

		output, _ := json.MarshalIndent(userProfile, "", "  ")
		fmt.Println(string(output))
	},
}

func init() {
	// Command registration is handled in main.go
	usersCmd.AddCommand(getUserMeCmd)
	usersCmd.AddCommand(getUserByHandleCmd)
}