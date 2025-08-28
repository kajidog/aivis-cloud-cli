package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"
)

// McpCmd is the command for starting MCP server
var McpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server for AivisCloud integration",
	Long: `Start a Model Context Protocol (MCP) server that provides access to AivisCloud 
voice models search and information.

Supports both stdio (default) and HTTP transports for maximum compatibility.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		transport, _ := cmd.Flags().GetString("transport")
		port, _ := cmd.Flags().GetInt("port")

		// Create MCP server
		server := CreateMCPServer()

		// Handle different transport modes
		switch transport {
		case "stdio":
			// Use stdio transport (default)
			fmt.Fprintf(os.Stderr, "Starting AivisCloud MCP server over stdio\n")
			stdioTransport := &mcp.StdioTransport{}
			return server.Run(context.Background(), stdioTransport)

		case "http":
			// Use HTTP transport
			handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
				return server
			}, nil)
			
			fmt.Printf("Starting AivisCloud MCP server on port %d\n", port)
			return http.ListenAndServe(fmt.Sprintf(":%d", port), handler)

		default:
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: Unsupported transport: %s\n", transport)
			fmt.Fprintf(cmd.ErrOrStderr(), "Supported transports: stdio, http\n")
			return fmt.Errorf("unsupported transport: %s", transport)
		}
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
	McpCmd.Flags().String("transport", "stdio", "Transport protocol (stdio or http)")
	McpCmd.Flags().Int("port", 8080, "Port for HTTP transport (ignored for stdio)")
}