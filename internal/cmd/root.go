package cmd

import (
	"github.com/noot-app/foundation-foods-mcp-server/internal/auth"
	"github.com/noot-app/foundation-foods-mcp-server/internal/config"
	"github.com/noot-app/foundation-foods-mcp-server/internal/mcpgo"
	"github.com/noot-app/foundation-foods-mcp-server/internal/query"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "foundation-foods-mcp-server",
	Short: "FoundationFoods MCP Server with DuckDB",
	Long: `FoundationFoods MCP Server provides access to the Foundation Foods dataset
via a remote MCP server.

The server operates in three modes:

1. STDIO Mode (--stdio): For local Claude Desktop integration
   - Uses stdio pipes for communication
   - No authentication required
   - Perfect for local development and Claude Desktop

2. HTTP Mode (default): For remote deployment over the internet
   - Exposes HTTP endpoints with JSON-RPC 2.0
   - Requires Bearer token authentication (except /health)
   - Ideal for shared/remote MCP server deployments

Available MCP Tools:
- search_foundation_foods_by_name: Search foundation foods by name

Authentication (HTTP Mode Only):
Bearer token authentication is required for all MCP endpoints except /health.
Use the FOUNDATIONFOODS_MCP_TOKEN environment variable to set the token.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if we should run in stdio mode (for Claude Desktop)
		stdio, _ := cmd.Flags().GetBool("stdio")

		if stdio {
			return runStdioMode(cmd, args)
		} else {
			return runHTTPMode(cmd, args)
		}
	},
}

func init() {
	rootCmd.Flags().Bool("stdio", false, "Run in stdio mode for local Claude Desktop integration (default: HTTP mode for remote deployment)")
}

// runStdioMode runs the MCP server in stdio mode for Claude Desktop
func runStdioMode(cmd *cobra.Command, args []string) error {
	// Use a logger that writes to stderr to avoid interfering with stdio MCP communication
	logger := config.NewLogger(true) // true for stdio mode

	// Load configuration
	cfg := config.Load()

	logger.Info("🔌 Starting FoundationFoods MCP Server in STDIO mode",
		"mode", "stdio",
		"description", "Local MCP server for Claude Desktop integration",
		"auth", "not required for stdio mode",
		"transport", "stdio pipes")

	// Load Foundation Foods data
	queryEngine, err := query.NewEngine(cfg.FoundationFoodsJsonFile, logger)
	if err != nil {
		logger.Error("Failed to initialize query engine", "error", err)
		return err
	}

	// Create auth (not needed for stdio but required by constructor)
	authenticator := auth.NewBearerTokenAuth(cfg.AuthToken)

	// Create MCP server
	mcpSrv := mcpgo.NewServer(queryEngine, authenticator, logger)

	// Run the MCP server on stdio transport (no auth needed for local use)
	return mcpSrv.ServeStdio()
}

// runHTTPMode runs the MCP server in HTTP mode for remote deployment
func runHTTPMode(cmd *cobra.Command, args []string) error {
	// Setup structured logging for HTTP mode
	logger := config.NewLogger(false) // false for HTTP mode

	// Load configuration
	cfg := config.Load()

	logger.Info("🌐 Starting FoundationFoods MCP Server in HTTP mode",
		"mode", "http",
		"description", "Remote MCP server with API key authentication",
		"auth", "Bearer token required (except /health endpoint)",
		"transport", "HTTP/JSON-RPC 2.0",
		"port", cfg.Port)

	// Load Foundation Foods data
	queryEngine, err := query.NewEngine(cfg.FoundationFoodsJsonFile, logger)
	if err != nil {
		logger.Error("Failed to initialize query engine", "error", err)
		return err
	}

	// Create auth
	authenticator := auth.NewBearerTokenAuth(cfg.AuthToken)

	// Create MCP server
	mcpSrv := mcpgo.NewServer(queryEngine, authenticator, logger)

	// Run the MCP server on HTTP transport with auth
	return mcpSrv.ServeHTTP(":" + cfg.Port)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// Run is the main entry point for the CLI application
func Run() error {
	return Execute()
}
