package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kajidog/aiviscloud-mcp/client/models/domain"
	"github.com/spf13/cobra"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Model management operations",
	Long:  "Search, list, and manage voice models",
}

var modelsSearchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search for voice models",
	Long:  "Search for available voice models with optional filters",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags
		author, _ := cmd.Flags().GetString("author")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")
		sort, _ := cmd.Flags().GetString("sort")
		public, _ := cmd.Flags().GetBool("public")
		outputFormat, _ := cmd.Flags().GetString("output")

		ctx := context.Background()
		var response *domain.ModelSearchResponse
		var err error

		// Handle specific search types
		if author != "" {
			response, err = aivisClient.SearchModelsByAuthor(ctx, author)
		} else if len(tags) > 0 {
			response, err = aivisClient.SearchModelsByTags(ctx, tags...)
		} else if public {
			query := ""
			if len(args) > 0 {
				query = args[0]
			}
			response, err = aivisClient.SearchPublicModels(ctx, query)
		} else {
			// Build search request
			builder := aivisClient.NewModelSearchRequest()
			
			if len(args) > 0 {
				builder = builder.WithQuery(args[0])
			}
			if limit > 0 {
				builder = builder.WithPageSize(limit)
			}
			if offset > 0 && limit > 0 {
				page := (offset / limit) + 1
				builder = builder.WithPage(page)
			}
			if sort != "" {
				builder = builder.WithSortBy(sort)
			}

			request := builder.Build()
			response, err = aivisClient.SearchModels(ctx, request)
		}

		if err != nil {
			return fmt.Errorf("failed to search models: %v", err)
		}

		return outputModels(response, outputFormat)
	},
}

var modelsGetCmd = &cobra.Command{
	Use:   "get [model-uuid]",
	Short: "Get model details",
	Long:  "Retrieve detailed information about a specific model",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		modelUUID := args[0]
		outputFormat, _ := cmd.Flags().GetString("output")
		speakers, _ := cmd.Flags().GetBool("speakers")

		ctx := context.Background()

		if speakers {
			speakerList, err := aivisClient.GetModelSpeakers(ctx, modelUUID)
			if err != nil {
				return fmt.Errorf("failed to get model speakers: %v", err)
			}

			return outputSpeakers(speakerList, outputFormat)
		}

		model, err := aivisClient.GetModel(ctx, modelUUID)
		if err != nil {
			return fmt.Errorf("failed to get model: %v", err)
		}

		return outputModel(model, outputFormat)
	},
}

var modelsPopularCmd = &cobra.Command{
	Use:   "popular",
	Short: "Get popular models",
	Long:  "Retrieve popular models sorted by download count",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		outputFormat, _ := cmd.Flags().GetString("output")

		if limit <= 0 {
			limit = 10
		}

		ctx := context.Background()
		response, err := aivisClient.GetPopularModels(ctx, limit)
		if err != nil {
			return fmt.Errorf("failed to get popular models: %v", err)
		}

		return outputModels(response, outputFormat)
	},
}

var modelsRecentCmd = &cobra.Command{
	Use:   "recent",
	Short: "Get recent models",
	Long:  "Retrieve recently updated models",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		outputFormat, _ := cmd.Flags().GetString("output")

		if limit <= 0 {
			limit = 10
		}

		ctx := context.Background()
		response, err := aivisClient.GetRecentModels(ctx, limit)
		if err != nil {
			return fmt.Errorf("failed to get recent models: %v", err)
		}

		return outputModels(response, outputFormat)
	},
}

var modelsTopRatedCmd = &cobra.Command{
	Use:   "top-rated",
	Short: "Get top-rated models",
	Long:  "Retrieve highest-rated models",
	RunE: func(cmd *cobra.Command, args []string) error {
		limit, _ := cmd.Flags().GetInt("limit")
		outputFormat, _ := cmd.Flags().GetString("output")

		if limit <= 0 {
			limit = 10
		}

		ctx := context.Background()
		response, err := aivisClient.GetTopRatedModels(ctx, limit)
		if err != nil {
			return fmt.Errorf("failed to get top-rated models: %v", err)
		}

		return outputModels(response, outputFormat)
	},
}

func outputModels(response *domain.ModelSearchResponse, format string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(response)
	
	case "table", "":
		// Table format
		fmt.Printf("Found %d models (Total: %d)\n\n", len(response.Models), response.Pagination.TotalResults)
		
		for _, model := range response.Models {
			fmt.Printf("UUID: %s\n", model.UUID)
			fmt.Printf("Name: %s\n", model.Name)
			if model.Author != "" {
				fmt.Printf("Author: %s\n", model.Author)
			}
			if model.Description != "" {
				fmt.Printf("Description: %s\n", model.Description)
			}
			if len(model.Tags) > 0 {
				fmt.Printf("Tags: %s\n", strings.Join(model.Tags, ", "))
			}
			if model.DownloadCount > 0 {
				fmt.Printf("Downloads: %d\n", model.DownloadCount)
			}
			if model.Rating > 0 {
				fmt.Printf("Rating: %.2f\n", model.Rating)
			}
			fmt.Println("---")
		}
		
		if response.Pagination.HasNext {
			fmt.Printf("\nMore results available. Use --offset %d to see more.\n", len(response.Models))
		}
		
	default:
		return fmt.Errorf("unsupported output format: %s. Supported formats: json, table", format)
	}
	
	return nil
}

func outputModel(model *domain.Model, format string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(model)
	
	case "table", "":
		fmt.Printf("Model Details\n")
		fmt.Printf("=============\n")
		fmt.Printf("UUID: %s\n", model.UUID)
		fmt.Printf("Name: %s\n", model.Name)
		if model.Author != "" {
			fmt.Printf("Author: %s\n", model.Author)
		}
		if model.Description != "" {
			fmt.Printf("Description: %s\n", model.Description)
		}
		if len(model.Tags) > 0 {
			fmt.Printf("Tags: %s\n", strings.Join(model.Tags, ", "))
		}
		if model.DownloadCount > 0 {
			fmt.Printf("Downloads: %d\n", model.DownloadCount)
		}
		if model.Rating > 0 {
			fmt.Printf("Rating: %.2f\n", model.Rating)
		}
		if !model.CreatedAt.IsZero() {
			fmt.Printf("Created: %s\n", model.CreatedAt.Format("2006-01-02 15:04:05"))
		}
		if !model.UpdatedAt.IsZero() {
			fmt.Printf("Updated: %s\n", model.UpdatedAt.Format("2006-01-02 15:04:05"))
		}
		
	default:
		return fmt.Errorf("unsupported output format: %s. Supported formats: json, table", format)
	}
	
	return nil
}

func outputSpeakers(speakers []domain.Speaker, format string) error {
	switch format {
	case "json":
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(speakers)
	
	case "table", "":
		fmt.Printf("Model Speakers\n")
		fmt.Printf("==============\n")
		
		for i, speaker := range speakers {
			fmt.Printf("Speaker %d:\n", i+1)
			fmt.Printf("  ID: %s\n", speaker.UUID)
			fmt.Printf("  Name: %s\n", speaker.Name)
			if speaker.Language != "" {
				fmt.Printf("  Language: %s\n", speaker.Language)
			}
			if speaker.Gender != "" {
				fmt.Printf("  Gender: %s\n", speaker.Gender)
			}
			fmt.Println()
		}
		
	default:
		return fmt.Errorf("unsupported output format: %s. Supported formats: json, table", format)
	}
	
	return nil
}

func init() {
	// Models search command flags
	modelsSearchCmd.Flags().String("author", "", "Filter by author")
	modelsSearchCmd.Flags().StringSlice("tags", nil, "Filter by tags")
	modelsSearchCmd.Flags().Int("limit", 0, "Maximum number of results")
	modelsSearchCmd.Flags().Int("offset", 0, "Offset for pagination")
	modelsSearchCmd.Flags().String("sort", "", "Sort field (name, created_at, updated_at, download_count, rating)")
	modelsSearchCmd.Flags().Bool("public", false, "Search only public models")
	modelsSearchCmd.Flags().String("output", "table", "Output format: table, json")

	// Models get command flags
	modelsGetCmd.Flags().String("output", "table", "Output format: table, json")
	modelsGetCmd.Flags().Bool("speakers", false, "Get model speakers instead of model details")

	// Models list commands flags
	modelsPopularCmd.Flags().Int("limit", 10, "Maximum number of results")
	modelsPopularCmd.Flags().String("output", "table", "Output format: table, json")
	
	modelsRecentCmd.Flags().Int("limit", 10, "Maximum number of results")
	modelsRecentCmd.Flags().String("output", "table", "Output format: table, json")
	
	modelsTopRatedCmd.Flags().Int("limit", 10, "Maximum number of results")
	modelsTopRatedCmd.Flags().String("output", "table", "Output format: table, json")

	// Add subcommands to models command
	modelsCmd.AddCommand(modelsSearchCmd)
	modelsCmd.AddCommand(modelsGetCmd)
	modelsCmd.AddCommand(modelsPopularCmd)
	modelsCmd.AddCommand(modelsRecentCmd)
	modelsCmd.AddCommand(modelsTopRatedCmd)
}