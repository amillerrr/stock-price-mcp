# Stock Price MCP Server

A Model Context Protocol (MCP) server that provides real-time stock price data from Yahoo Finance. This server integrates with Claude Desktop to allow you to query current stock prices directly in your conversations.

## Features

- Real-time stock price data from Yahoo Finance
- Dockerized for easy deployment
- Compatible with Claude Desktop
- Supports major stock symbols (AAPL, GOOGL, MSFT, TSLA, etc.)
- Returns current price, change, day high/low, and volume

## Quick Start

### 1. Build the Docker Image

```bash
# Clone or navigate to the project directory
cd stock-price-mcp

# Build the Docker image
docker build -t stock-price-mcp:latest .
```

### 2. Test the Server

Test the MCP server directly with Docker:

```bash
# Test initialization
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | docker run -i --rm stock-price-mcp:latest

# Test tools list
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | docker run -i --rm stock-price-mcp:latest

# Test getting Apple stock price
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_stock_price","arguments":{"symbol":"AAPL"}}}' | docker run -i --rm stock-price-mcp:latest

# Test other stocks
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_stock_price","arguments":{"symbol":"GOOGL"}}}' | docker run -i --rm stock-price-mcp:latest
```

### 3. Configure Claude Desktop

#### Option A: Use the provided config file

Copy the configuration from `config/mcp-config.json` to your Claude Desktop configuration:

**Location of Claude Desktop config:**
- **macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows:** `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux:** `~/.config/Claude/claude_desktop_config.json`

**Copy the contents:**
```json
{
  "mcpServers": {
    "stock-price-checker": {
      "command": "docker",
      "args": [
        "run",
        "--rm",
        "-i",
        "stock-price-mcp:latest"
      ],
      "env": {}
    }
  }
}
```

#### Option B: Add to existing config

If you already have other MCP servers configured, add the `stock-price-checker` section to your existing `mcpServers` object.

### 4. Restart Claude Desktop

After updating the configuration, completely quit and restart Claude Desktop for the changes to take effect.

## Usage

Once configured, you can ask Claude for stock prices in natural language:

- "What's the current stock price of Apple?"
- "Get me the price for GOOGL"
- "Show me Tesla's stock price"
- "What's Microsoft trading at right now?"

## Supported Stock Symbols

The server works with any valid stock symbol available on Yahoo Finance, including:

- **Tech:** AAPL, GOOGL, MSFT, TSLA, NVDA, META, AMZN
- **Finance:** JPM, BAC, WFC, GS
- **Healthcare:** JNJ, PFE, UNH, ABBV
- **And many more...**

## Project Structure

```
stock-price-mcp/
├── main.go              # MCP server implementation
├── go.mod               # Go module file
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose configuration
├── config/
│   └── mcp-config.json  # Claude Desktop configuration
└── README.md           # This file
```

## Development

### Local Development (without Docker)

```bash
# Build locally
go build -o stock-price-mcp

# Test locally
echo '{"jsonrpc":"2.0","id":1,"method":"tools/list"}' | ./stock-price-mcp
```

### Using Docker Compose

```bash
# Start the service
docker-compose up -d

# View logs
docker-compose logs -f stock-price-mcp

# Stop the service
docker-compose down
```

## Troubleshooting

### Common Issues

1. **"No tools available" in Claude Desktop**
   - Ensure Docker image is built: `docker build -t stock-price-mcp:latest .`
   - Verify Claude Desktop config is correct
   - Restart Claude Desktop completely

2. **"Unable to fetch data" errors**
   - Check internet connection
   - Verify the stock symbol is valid
   - Yahoo Finance may temporarily block requests

3. **Docker build fails**
   - Ensure you have Go 1.21+ in your Dockerfile
   - Check that all files are in the correct directory

### Testing Configuration

Verify your configuration is working:

```bash
# Test that Docker can run the image
docker run --rm stock-price-mcp:latest --help

# Test JSON-RPC communication
docker run -i --rm stock-price-mcp:latest < config/test-request.json
```

## Technical Details

- **Language:** Go 1.25+
- **Protocol:** JSON-RPC 2.0 over stdio
- **Data Source:** Yahoo Finance API

## License

This project is provided as-is for educational and personal use.

## Contributing

Feel free to submit issues and enhancement requests.
