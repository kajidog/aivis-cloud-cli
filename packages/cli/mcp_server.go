package main

import (
	"fmt"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// McpCmd is the command for starting MCP server
var McpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for AivisCloud integration",
	Long: `Start a Model Context Protocol (MCP) server that provides access to AivisCloud 
voice models search and information.

HTTP mode only: Due to a known bug in MCP Go SDK v0.2.0, stdio mode is temporarily disabled.
Use HTTP transport for both testing and production use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		transport, _ := cmd.Flags().GetString("transport")
		port, _ := cmd.Flags().GetInt("port")

		// Check for stdio mode and warn
		if transport == "stdio" {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: stdio mode is temporarily disabled due to a known bug in MCP Go SDK v0.2.0\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "Please use HTTP mode instead: --transport http --port %d\n", port)
			return fmt.Errorf("stdio transport is disabled")
		}

		// Only HTTP mode is supported for now
		if transport != "http" {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: Only HTTP transport is supported currently\n")
			fmt.Fprintf(cmd.ErrOrStderr(), "Usage: --transport http --port %d\n", port)
			return fmt.Errorf("unsupported transport: %s. Only 'http' is supported", transport)
		}

		// Create MCP server
		server := CreateMCPServer()

		// Run server over HTTP only
		handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
			return server
		}, nil)
		
		fmt.Printf("Starting AivisCloud MCP server on port %d\n", port)
		return http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
	},
}

// CreateMCPServer creates a new MCP server with all tools registered
func CreateMCPServer() *mcp.Server {
	// Create server with AivisCloud implementation info
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "aivis-cloud-cli",
		Version: "1.0.0",
	}, nil)

	// Register all tools
	RegisterAllTools(server)

	return server
}

func init() {
	McpCmd.Flags().String("transport", "http", "Transport protocol (http only - stdio disabled)")
	McpCmd.Flags().Int("port", 8080, "Port for HTTP transport")
}