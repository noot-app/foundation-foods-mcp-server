package mcpgo

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/noot-app/foundation-foods-mcp-server/internal/auth"
	"github.com/noot-app/foundation-foods-mcp-server/internal/query"
)

// responseRecorder wraps http.ResponseWriter to capture response details
type responseRecorder struct {
	http.ResponseWriter
	statusCode    int
	bytesWritten  int
	headerWritten bool
}

func (r *responseRecorder) WriteHeader(code int) {
	if r.headerWritten {
		return // Prevent duplicate WriteHeader calls
	}
	r.statusCode = code
	r.headerWritten = true
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(data []byte) (int, error) {
	if !r.headerWritten {
		r.WriteHeader(http.StatusOK)
	}
	n, err := r.ResponseWriter.Write(data)
	r.bytesWritten += n
	return n, err
}

// Server wraps the mark3labs MCP server with authentication
type Server struct {
	mcpServer   *server.MCPServer
	queryEngine query.QueryEngine
	auth        *auth.BearerTokenAuth
	log         *slog.Logger
}

// NewServer creates a new MCP server with the mark3labs SDK
func NewServer(queryEngine query.QueryEngine, authenticator *auth.BearerTokenAuth, logger *slog.Logger) *Server {
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"FoundationFoods MCP Server",
		"1.0.0",
		server.WithToolCapabilities(false), // Tools don't change dynamically
		server.WithRecovery(),              // Recover from panics
		server.WithLogging(),               // Enable logging
	)

	s := &Server{
		mcpServer:   mcpServer,
		queryEngine: queryEngine,
		auth:        authenticator,
		log:         logger,
	}

	// Add tools
	s.addTools()

	return s
}

func (s *Server) addTools() {
	// Search products by brand and name tool
	searchTool := mcp.NewTool("search_foundation_foods_by_name",
		mcp.WithDescription("Search USDA foundation foods by name. This tool is only meant to be used for generic product searches like 'milk', 'eggs', 'Cheese, cheddar', 'Broccoli, raw', etc."),
		mcp.WithString("name",
			mcp.Required(),
			mcp.MinLength(1), // must be at least 1 char
			mcp.Description("Food items/name to search for. Required and must be a non-empty string."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results (default: 3, max: 10)"),
			mcp.DefaultNumber(3),
			mcp.Min(1),
			mcp.Max(10),
		),
		mcp.WithOutputSchema[query.SearchProductsResponse](),
		mcp.WithIdempotentHintAnnotation(true),
	)

	s.mcpServer.AddTool(searchTool, s.handleFoodSearch)

	// Simplified nutrients search tool
	simplifiedTool := mcp.NewTool("search_foundation_foods_and_return_nutrients_simplified",
		mcp.WithDescription("Search USDA foundation foods by name and return simplified nutrient information. Returns only essential nutrient data (name, amount, unit) for each food match."),
		mcp.WithString("name",
			mcp.Required(),
			mcp.MinLength(1),
			mcp.Description("Food items/name to search for. Required and must be a non-empty string."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results (default: 3, max: 10)"),
			mcp.DefaultNumber(3),
			mcp.Min(1),
			mcp.Max(10),
		),
		mcp.WithOutputSchema[query.SimplifiedNutrientResponse](),
		mcp.WithIdempotentHintAnnotation(true),
	)

	s.mcpServer.AddTool(simplifiedTool, s.handleSimplifiedFoodSearch)
}

// ServeHTTP serves the MCP server over HTTP with authentication
func (s *Server) ServeHTTP(addr string) error {
	// Create a custom HTTP handler that includes authentication
	mux := http.NewServeMux()

	// Health endpoint (no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "healthy",
		})
	})

	// Create the streamable HTTP server
	streamableServer := server.NewStreamableHTTPServer(
		s.mcpServer,
		server.WithEndpointPath("/mcp"),
		server.WithStateLess(true), // Stateless for better OpenAI compatibility
	)

	// MCP endpoint with authentication and enhanced error logging
	mux.HandleFunc("/mcp", func(w http.ResponseWriter, r *http.Request) {
		// Add recovery middleware for better error handling
		defer func() {
			if recovery := recover(); recovery != nil {
				s.log.Error("MCP endpoint panic recovered",
					"panic", recovery,
					"method", r.Method,
					"url", r.URL.String(),
					"remote_addr", r.RemoteAddr)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			}
		}()

		s.log.Debug("MCP request received",
			"method", r.Method,
			"url", r.URL.String(),
			"content_type", r.Header.Get("Content-Type"),
			"content_length", r.ContentLength,
			"remote_addr", r.RemoteAddr)

		// Check authentication for all non-health endpoints
		if !s.auth.IsAuthorized(r) {
			s.auth.SetUnauthorizedHeaders(w)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			s.log.Warn("Unauthorized MCP request", "remote_addr", r.RemoteAddr, "user_agent", r.UserAgent())
			return
		}

		// Create a custom ResponseWriter to capture response details
		recorder := &responseRecorder{ResponseWriter: w}

		// Forward to the streamable HTTP server
		streamableServer.ServeHTTP(recorder, r)

		s.log.Debug("MCP response sent",
			"status_code", recorder.statusCode,
			"response_size", recorder.bytesWritten,
			"content_type", recorder.Header().Get("Content-Type"))
	})

	s.log.Info("Starting MCP server", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

// ServeStdio serves the MCP server over stdio (no auth required for local use)
func (s *Server) ServeStdio() error {
	s.log.Info("Starting MCP server in stdio mode")
	return server.ServeStdio(s.mcpServer)
}

func (s *Server) handleFoodSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.log.Debug("handleFoodSearch: Starting tool call",
		"arguments", request.GetArguments())

	// Extract arguments
	name, err := request.RequireString("name")
	if err != nil {
		s.log.Warn("handleFoodSearch: Missing 'name' parameter", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Missing required parameter 'name': %v", err)), nil
	}

	// Validate minimum lengths
	if len(name) < 1 {
		s.log.Warn("handleFoodSearch: Invalid 'name' parameter", "length", len(name))
		return mcp.NewToolResultError("Parameter 'name' must be at least 1 character long"), nil
	}

	limitFloat := request.GetFloat("limit", 3.0)
	limit := int(limitFloat)
	if limit <= 0 {
		limit = 3
	}
	if limit > 10 {
		limit = 10
	}

	s.log.Debug("MCP search_foundation_foods_by_name called",
		"name", name,
		"limit", limit)

	// Execute search
	products, err := s.queryEngine.SearchFoodsByName(ctx, name, limit)
	if err != nil {
		s.log.Error("Food search failed", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	// Prepare structured response
	response := query.SearchProductsResponse{
		Found:    len(products) > 0,
		Count:    len(products),
		Products: products,
	}

	// Create fallback text for backwards compatibility
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		s.log.Error("handleFoodSearch: Failed to marshal response", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}

	s.log.Debug("handleFoodSearch: Returning structured result",
		"found", response.Found,
		"count", response.Count,
		"response_size", len(responseJSON))

	// Return both structured content and text fallback for maximum compatibility
	return mcp.NewToolResultStructured(response, string(responseJSON)), nil
}

func (s *Server) handleSimplifiedFoodSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	s.log.Debug("handleSimplifiedFoodSearch: Starting tool call",
		"arguments", request.GetArguments())

	// Extract arguments
	name, err := request.RequireString("name")
	if err != nil {
		s.log.Warn("handleSimplifiedFoodSearch: Missing 'name' parameter", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Missing required parameter 'name': %v", err)), nil
	}

	// Validate minimum lengths
	if len(name) < 1 {
		s.log.Warn("handleSimplifiedFoodSearch: Invalid 'name' parameter", "length", len(name))
		return mcp.NewToolResultError("Parameter 'name' must be at least 1 character long"), nil
	}

	limitFloat := request.GetFloat("limit", 3.0)
	limit := int(limitFloat)
	if limit <= 0 {
		limit = 3
	}
	if limit > 10 {
		limit = 10
	}

	s.log.Debug("MCP search_foundation_foods_and_return_nutrients_simplified called",
		"name", name,
		"limit", limit)

	// Execute simplified search
	response, err := s.queryEngine.SearchFoodsByNameSimplified(ctx, name, limit)
	if err != nil {
		s.log.Error("Simplified food search failed", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Search failed: %v", err)), nil
	}

	// Create fallback text for backwards compatibility
	responseJSON, err := json.MarshalIndent(response, "", "  ")
	if err != nil {
		s.log.Error("handleSimplifiedFoodSearch: Failed to marshal response", "error", err)
		return mcp.NewToolResultError(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}

	s.log.Debug("handleSimplifiedFoodSearch: Returning structured result",
		"found", response.Found,
		"count", response.Count,
		"foods_count", len(response.Foods),
		"response_size", len(responseJSON))

	// Return both structured content and text fallback for maximum compatibility
	return mcp.NewToolResultStructured(response, string(responseJSON)), nil
}
