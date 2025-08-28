package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/kajidog/aivis-cloud-cli/client/models/domain"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/viper"
)

// SearchModelsParams parameters for search_models tool
type SearchModelsParams struct {
	Query      string   `json:"query,omitempty"`
	Author     string   `json:"author,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Sort       string   `json:"sort,omitempty"`
	PublicOnly bool     `json:"public_only,omitempty"`
}

// GetModelParams parameters for get_model tool
type GetModelParams struct {
	UUID string `json:"uuid,omitempty"`
}

// GetModelSpeakersParams parameters for get_model_speakers tool
type GetModelSpeakersParams struct {
	UUID string `json:"uuid,omitempty"`
}

// getDefaultModelUUID returns the default model UUID from config or fallback
func getDefaultModelUUID() string {
	// Try to get from config first
	if defaultUUID := viper.GetString("default_model_uuid"); defaultUUID != "" {
		return defaultUUID
	}
	
	// Use fallback UUID
	return "a59cb814-0083-4369-8542-f51a29e72af7"
}

// RegisterModelsTools registers all model-related MCP tools
func RegisterModelsTools(server *mcp.Server) {
	// Add search models tool (consolidated)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "search_models",
		Description: "Search AivisCloud voice models with sorting and filtering (replaces popular/recent/top-rated)",
	}, handleSearchModels)

	// Add get model tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_model",
		Description: "Get essential information about a voice model (uses default model if uuid not specified)",
	}, handleGetModel)

	// Add get model speakers tool
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_model_speakers",
		Description: "Get speaker list for a voice model (uses default model if uuid not specified)",
	}, handleGetModelSpeakers)
}

func handleSearchModels(ctx context.Context, req *mcp.CallToolRequest, args SearchModelsParams) (*mcp.CallToolResult, any, error) {
	
	// Use existing client to perform search
	var response *domain.ModelSearchResponse
	var err error

	// Handle special sorting cases first
	if args.Sort == "popularity" || args.Sort == "downloads" {
		limit := args.Limit
		if limit <= 0 {
			limit = 5
		}
		response, err = aivisClient.GetPopularModels(ctx, limit)
	} else if args.Sort == "recent" || args.Sort == "updated" {
		limit := args.Limit
		if limit <= 0 {
			limit = 5
		}
		response, err = aivisClient.GetRecentModels(ctx, limit)
	} else if args.Sort == "rating" || args.Sort == "top-rated" {
		limit := args.Limit
		if limit <= 0 {
			limit = 5
		}
		response, err = aivisClient.GetTopRatedModels(ctx, limit)
	} else if args.Author != "" {
		response, err = aivisClient.SearchModelsByAuthor(ctx, args.Author)
	} else if len(args.Tags) > 0 {
		response, err = aivisClient.SearchModelsByTags(ctx, args.Tags...)
	} else if args.PublicOnly {
		response, err = aivisClient.SearchPublicModels(ctx, args.Query)
	} else {
		// Build search request
		builder := aivisClient.NewModelSearchRequest()
		
		if args.Query != "" {
			builder = builder.WithQuery(args.Query)
		}
		
		// Default limit for token efficiency
		limit := args.Limit
		if limit <= 0 {
			limit = 5 // Default to 5 results to minimize tokens
		}
		builder = builder.WithPageSize(limit)
		
		if args.Sort != "" {
			builder = builder.WithSortBy(args.Sort)
		}

		request := builder.Build()
		response, err = aivisClient.SearchModels(ctx, request)
	}

	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Search failed: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	// Format response for MCP
	resultText := formatSearchResponse(response)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleGetModel(ctx context.Context, req *mcp.CallToolRequest, args GetModelParams) (*mcp.CallToolResult, any, error) {
	
	// Use default UUID if not provided
	uuid := args.UUID
	if uuid == "" {
		uuid = getDefaultModelUUID()
	}

	// Get model details
	model, err := aivisClient.GetModel(ctx, uuid)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get model: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	// Format response for MCP
	resultText := formatModelResponse(model)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func handleGetModelSpeakers(ctx context.Context, req *mcp.CallToolRequest, args GetModelSpeakersParams) (*mcp.CallToolResult, any, error) {
	
	// Use default UUID if not provided
	uuid := args.UUID
	if uuid == "" {
		uuid = getDefaultModelUUID()
	}

	// Get model speakers
	speakers, err := aivisClient.GetModelSpeakers(ctx, uuid)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf("Failed to get model speakers: %v", err)}},
			IsError: true,
		}, nil, nil
	}

	// Format response for MCP
	resultText := formatSpeakersResponse(speakers)
	
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: resultText}},
	}, nil, nil
}

func formatSearchResponse(response *domain.ModelSearchResponse) string {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("Found %d models:\n\n", len(response.Models)))
	
	for _, model := range response.Models {
		result.WriteString(fmt.Sprintf("• %s (UUID: %s)\n", model.Name, model.UUID))
		
		// Author - most important info after name
		if model.User != nil && model.User.Name != "" {
			result.WriteString(fmt.Sprintf("  Author: %s\n", model.User.Name))
		} else if model.Author != "" {
			result.WriteString(fmt.Sprintf("  Author: %s\n", model.Author))
		}
		
		// Downloads - popularity indicator
		if model.TotalDownloadCount > 0 {
			result.WriteString(fmt.Sprintf("  Downloads: %d\n", model.TotalDownloadCount))
		} else if model.DownloadCount > 0 {
			result.WriteString(fmt.Sprintf("  Downloads: %d\n", model.DownloadCount))
		}
		
		result.WriteString("\n")
	}
	
	// Pagination info only if relevant
	if response.Pagination.HasNext {
		result.WriteString(fmt.Sprintf("(Showing %d/%d results)\n", len(response.Models), response.Pagination.TotalResults))
	}
	
	return result.String()
}

func formatModelResponse(model *domain.Model) string {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("%s (UUID: %s)\n", model.Name, model.UUID))
	
	// Author - essential info
	if model.User != nil && model.User.Name != "" {
		result.WriteString(fmt.Sprintf("Author: %s\n", model.User.Name))
	} else if model.Author != "" {
		result.WriteString(fmt.Sprintf("Author: %s\n", model.Author))
	}
	
	// Brief description only if short (under 100 chars)
	if model.Description != "" && len(model.Description) < 100 {
		result.WriteString(fmt.Sprintf("Description: %s\n", model.Description))
	}
	
	// Key metrics
	if model.TotalDownloadCount > 0 {
		result.WriteString(fmt.Sprintf("Downloads: %d\n", model.TotalDownloadCount))
	} else if model.DownloadCount > 0 {
		result.WriteString(fmt.Sprintf("Downloads: %d\n", model.DownloadCount))
	}
	
	// Speaker count (more useful than individual speakers)
	if len(model.Speakers) > 0 {
		result.WriteString(fmt.Sprintf("Speakers: %d\n", len(model.Speakers)))
	}
	
	return result.String()
}

func formatSpeakersResponse(speakers []domain.Speaker) string {
	var result strings.Builder
	
	result.WriteString(fmt.Sprintf("Speakers (%d):\n\n", len(speakers)))
	
	for _, speaker := range speakers {
		result.WriteString(fmt.Sprintf("• %s (ID: %s)\n", speaker.Name, speaker.UUID))
		
		// Only show essential info - language is most important
		if len(speaker.SupportedLanguages) > 0 {
			result.WriteString(fmt.Sprintf("  Languages: %s\n", strings.Join(speaker.SupportedLanguages, ", ")))
		} else if speaker.Language != "" {
			result.WriteString(fmt.Sprintf("  Language: %s\n", speaker.Language))
		}
		
		// Style count is more useful than individual styles
		if len(speaker.Styles) > 0 {
			result.WriteString(fmt.Sprintf("  Styles: %d\n", len(speaker.Styles)))
		}
	}
	
	return result.String()
}