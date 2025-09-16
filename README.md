# Foundation Foods MCP Server 🔌

[![lint](https://github.com/noot-app/foundation-foods-mcp-server/actions/workflows/lint.yml/badge.svg)](https://github.com/noot-app/foundation-foods-mcp-server/actions/workflows/lint.yml)
[![test](https://github.com/noot-app/foundation-foods-mcp-server/actions/workflows/test.yml/badge.svg)](https://github.com/noot-app/foundation-foods-mcp-server/actions/workflows/test.yml)
[![build](https://github.com/noot-app/foundation-foods-mcp-server/actions/workflows/build.yml/badge.svg)](https://github.com/noot-app/foundation-foods-mcp-server/actions/workflows/build.yml)
[![docker](https://github.com/noot-app/foundation-foods-mcp-server/actions/workflows/docker.yml/badge.svg)](https://github.com/noot-app/foundation-foods-mcp-server/actions/workflows/docker.yml)

MCP server to provide models context around the USDA's Foundation Foods database

![logo](./docs/assets/logo.png)

## Usage 💻

This MCP server can operate in two distinct modes:

### 1. **STDIO Mode** (Local Claude Desktop Integration)

- **Use case**: Local development and Claude Desktop integration
- **Command**: `./foundation-foods-mcp-server --stdio`
- **Transport**: stdio pipes
- **Authentication**: None required
- **Perfect for**: Claude Desktop, local development, testing

### 2. **HTTP Mode** (Remote Deployment)

- **Use case**: Remote MCP server accessible over the internet
- **Command**: `./foundation-foods-mcp-server` (default mode)
- **Transport**: HTTP with JSON-RPC 2.0
- **Authentication**: Bearer token required (except `/health` endpoint)
- **Perfect for**: Shared deployments, cloud hosting, team access, mcp as a service

## Demo 📹

Check out the related project [noot-app/openfoodfacts-mcp-server](https://github.com/noot-app/openfoodfacts-mcp-server?tab=readme-ov-file#demo-) for a near identical MCP server with a demo video showcasing its capabilities and how to use it with Claude Desktop.

## How It Works 💡

This MCP server provides access to the USDA Foundation Foods database through three specialized tools, each designed for different use cases and levels of detail.

## Available Tools 🛠️

### 1. `search_foundation_foods_by_name`

Basic food search with complete data

- **Purpose**: Search for foods and get comprehensive food information
- **Returns**: Complete food details including all available nutrients
- **Best for**: Detailed nutritional analysis, research, when you need all available data
- **Example**: Get complete nutritional profile for "milk" including every measured nutrient

### 2. `search_foundation_foods_and_return_nutrients`

Customizable nutrient filtering

- **Purpose**: Search for foods with customizable nutrient selection
- **Returns**: Essential nutrient data (name, amount, unit) for specified nutrients only
- **Customization**: Accepts `nutrients_to_include` parameter to filter which nutrients to return
- **Best for**: Targeted nutritional queries, meal planning, when you want specific nutrients
- **Example**: Get only protein, calcium, and vitamin D data for "milk"

### 3. `search_foundation_foods_and_return_nutrients_simplified`

Fixed high-value nutrients (no customization)

- **Purpose**: Search with a fixed, optimized set of essential nutrients
- **Returns**: Default nutrient set
- **Optimization**: Pre-selected nutrients based on comprehensive data analysis
- **Best for**: Consistent results, general nutrition tracking, when you want the "best" nutrients without customization
- **Example**: Get the top nutrients for "milk" - always the same essential nutrients

## Local Setup for Claude Desktop (STDIO Mode)

This setup uses **STDIO mode** for local Claude Desktop integration.

### 1. Build the Binary

```bash
script/build --simple
```

### 2. Configure Claude Desktop

Add this to your Claude Desktop MCP settings (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS):

```json
{
  "mcpServers": {
    "foundationfoods": {
      "command": "/path/to/foundation-foods-mcp-server",
      "args": ["--stdio"],
      "env": {
        "FOUNDATIONFOODS_MCP_TOKEN": "your-secret-token",
        "DATA_DIR": "/full/path/to/foundation-foods-mcp-server/data",
        "ENV": "development"
      }
    }
  }
}
```

### 3. Try it Out

Restart Claude Desktop. The mcp server will automatically start and be ready for food product queries.

## Remote Deployment (HTTP Mode)

This setup uses **HTTP mode** for remote deployment with authentication.

### Environment Variables

For production deployment (HTTP mode), configure these environment variables:

```bash
# Required: Authentication
FOUNDATIONFOODS_MCP_TOKEN=your-production-secret-token

# Optional: Server configuration  
PORT=8080
ENV=production
```

### Running in HTTP Mode

For remote deployment, run **without** the `--stdio` flag (HTTP mode is the default):

```bash
./foundation-foods-mcp-server
```

This will start an HTTP server on the configured port (default 8080) with:

- `/health` endpoint (no authentication required)
- `/mcp` endpoint (Bearer token authentication required)

## Quick Reference

### Command Options

| Mode | Command | Use Case | Authentication | Transport |
|------|---------|----------|----------------|-----------|
| **STDIO** | `./foundation-foods-mcp-server --stdio` | Claude Desktop, local development | None | stdio pipes |
| **HTTP** | `./foundation-foods-mcp-server` | Remote deployment, shared access | Bearer token | HTTP/JSON-RPC |

### Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `FOUNDATIONFOODS_MCP_TOKEN` | Yes (HTTP mode) | - | Bearer token for authentication |
| `PORT` | No | `8080` | HTTP server port (HTTP mode only) |
| `ENV` | No | `production` | Environment (development/production) |
| `LOG_LEVEL` | No | `INFO` | The log level |

### HTTP Endpoints (HTTP Mode Only)

| Endpoint | Authentication | Description |
|----------|----------------|-------------|
| `/health` | None | Health check endpoint |
| `/mcp` | Bearer token | MCP JSON-RPC 2.0 endpoint |

## STDIO Mode (Local Development)

A cool tip for developing locally, you can actually do this and it will return a result from the MCP server:

```bash
echo '{"jsonrpc": "2.0", "method": "tools/call", "params": {"name": "search_foundation_foods_and_return_nutrients_simplified", "arguments": {"name": "milk", "limit": 2}}, "id": 1}' | go run ./cmd/foundation-foods-mcp-server --stdio
```
