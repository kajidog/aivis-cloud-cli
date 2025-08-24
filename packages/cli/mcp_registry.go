package main

import (
	"github.com/kajidog/aivis-cloud-cli/client"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SetClient sets the global Aivis Cloud CLIent for MCP tools
func SetClient(client *client.Client) {
	aivisClient = client
}

// RegisterAllTools registers all MCP tools from different categories
func RegisterAllTools(server *mcp.Server) {
	// Register model-related tools
	RegisterModelsTools(server)

	// Register TTS-related tools
	RegisterTTSTools(server)

	// Future tool categories can be added here:
	// RegisterUserTools(server)
	// RegisterPaymentTools(server)
}
