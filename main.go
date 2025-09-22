package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "strings"
    "time"
)

type MCPServer struct{}

type JSONRPCRequest struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Method  string      `json:"method"`
    Params  interface{} `json:"params,omitempty"`
}

type JSONRPCResponse struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      interface{} `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
    Code    int    `json:"code"`
    Message string `json:"message"`
}

func NewMCPServer() *MCPServer {
    return &MCPServer{}
}

func (s *MCPServer) HandleRequest(req JSONRPCRequest) JSONRPCResponse {
    // Ensure ID is never null - use 0 if not provided
    id := req.ID
    if id == nil {
        id = 0
    }

    switch req.Method {
    case "initialize":
        return s.handleInitialize(req, id)
    case "tools/list":
        return s.handleToolsList(req, id)
    case "tools/call":
        return s.handleToolsCall(req, id)
    default:
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      id,
            Error:   &RPCError{Code: -32601, Message: "Method not found"},
        }
    }
}

func (s *MCPServer) handleInitialize(req JSONRPCRequest, id interface{}) JSONRPCResponse {
    return JSONRPCResponse{
        JSONRPC: "2.0",
        ID:      id,
        Result: map[string]interface{}{
            "protocolVersion": "2024-11-05",
            "capabilities": map[string]interface{}{
                "tools": map[string]interface{}{},
            },
            "serverInfo": map[string]interface{}{
                "name":    "stock-price-checker",
                "version": "1.0.0",
            },
        },
    }
}

func (s *MCPServer) handleToolsList(req JSONRPCRequest, id interface{}) JSONRPCResponse {
    tools := []map[string]interface{}{
        {
            "name":        "get_stock_price",
            "description": "Get current stock price and basic info for a company using Yahoo Finance",
            "inputSchema": map[string]interface{}{
                "type": "object",
                "properties": map[string]interface{}{
                    "symbol": map[string]interface{}{
                        "type":        "string",
                        "description": "Stock symbol (e.g., AAPL, GOOGL, MSFT, TSLA)",
                    },
                },
                "required": []string{"symbol"},
            },
        },
    }

    return JSONRPCResponse{
        JSONRPC: "2.0",
        ID:      id,
        Result:  map[string]interface{}{"tools": tools},
    }
}

func (s *MCPServer) handleToolsCall(req JSONRPCRequest, id interface{}) JSONRPCResponse {
    // Safely extract params
    if req.Params == nil {
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      id,
            Error:   &RPCError{Code: -32602, Message: "Missing params"},
        }
    }

    params, ok := req.Params.(map[string]interface{})
    if !ok {
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      id,
            Error:   &RPCError{Code: -32602, Message: "Invalid params"},
        }
    }

    name, ok := params["name"].(string)
    if !ok {
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      id,
            Error:   &RPCError{Code: -32602, Message: "Missing tool name"},
        }
    }

    arguments, ok := params["arguments"].(map[string]interface{})
    if !ok {
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      id,
            Error:   &RPCError{Code: -32602, Message: "Missing arguments"},
        }
    }

    switch name {
    case "get_stock_price":
        return s.getStockPrice(id, arguments)
    default:
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      id,
            Error:   &RPCError{Code: -32602, Message: "Unknown tool"},
        }
    }
}

func (s *MCPServer) getStockPrice(id interface{}, args map[string]interface{}) JSONRPCResponse {
    // Safely extract symbol
    symbolInterface, ok := args["symbol"]
    if !ok {
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      id,
            Error:   &RPCError{Code: -32602, Message: "Missing symbol parameter"},
        }
    }

    symbol, ok := symbolInterface.(string)
    if !ok {
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      id,
            Error:   &RPCError{Code: -32602, Message: "Symbol must be a string"},
        }
    }

    if symbol == "" {
        return JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      id,
            Error:   &RPCError{Code: -32602, Message: "Symbol cannot be empty"},
        }
    }

    symbol = strings.ToUpper(symbol)
    
    // Try multiple Yahoo Finance endpoints with proper headers
    client := &http.Client{Timeout: 10 * time.Second}
    
    urls := []string{
        fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s", symbol),
        fmt.Sprintf("https://query2.finance.yahoo.com/v1/finance/quoteResponse?symbols=%s", symbol),
    }
    
    for _, url := range urls {
        if result := s.tryEndpoint(client, url, symbol, id); result != nil {
            return *result
        }
    }
    
    return JSONRPCResponse{
        JSONRPC: "2.0",
        ID:      id,
        Error:   &RPCError{Code: -32603, Message: fmt.Sprintf("Unable to fetch data for symbol: %s", symbol)},
    }
}

func (s *MCPServer) tryEndpoint(client *http.Client, url, symbol string, id interface{}) *JSONRPCResponse {
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        return nil
    }
    
    // Critical: Add User-Agent to avoid being blocked
    req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
    
    resp, err := client.Do(req)
    if err != nil {
        return nil
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil
    }

    var data map[string]interface{}
    if err := json.Unmarshal(body, &data); err != nil {
        return nil
    }
    
    if stockInfo := s.extractStockInfo(data, symbol); stockInfo != "" {
        result := &JSONRPCResponse{
            JSONRPC: "2.0",
            ID:      id,
            Result: map[string]interface{}{
                "content": []map[string]interface{}{
                    {
                        "type": "text",
                        "text": stockInfo,
                    },
                },
            },
        }
        return result
    }
    
    return nil
}

func (s *MCPServer) extractStockInfo(data map[string]interface{}, symbol string) string {
    // Try chart API format
    if chart, ok := data["chart"].(map[string]interface{}); ok {
        if results, ok := chart["result"].([]interface{}); ok && len(results) > 0 {
            if result, ok := results[0].(map[string]interface{}); ok {
                if meta, ok := result["meta"].(map[string]interface{}); ok {
                    return s.formatStockData(meta, symbol)
                }
            }
        }
    }
    
    // Try quote response format
    if quoteResponse, ok := data["quoteResponse"].(map[string]interface{}); ok {
        if results, ok := quoteResponse["result"].([]interface{}); ok && len(results) > 0 {
            if result, ok := results[0].(map[string]interface{}); ok {
                return s.formatQuoteData(result, symbol)
            }
        }
    }
    
    return ""
}

func (s *MCPServer) formatStockData(meta map[string]interface{}, symbol string) string {
    currentPrice, _ := meta["regularMarketPrice"].(float64)
    previousClose, _ := meta["previousClose"].(float64)
    dayHigh, _ := meta["regularMarketDayHigh"].(float64)
    dayLow, _ := meta["regularMarketDayLow"].(float64)
    volume, _ := meta["regularMarketVolume"].(float64)
    
    if currentPrice == 0 {
        return ""
    }
    
    change := currentPrice - previousClose
    changePercent := 0.0
    if previousClose != 0 {
        changePercent = (change / previousClose) * 100
    }
    
    result := fmt.Sprintf(`Stock: %s
Current Price: $%.2f
Previous Close: $%.2f
Change: $%.2f (%.2f%%)`,
        symbol, currentPrice, previousClose, change, changePercent)
    
    if dayHigh > 0 {
        result += fmt.Sprintf("\nDay High: $%.2f", dayHigh)
    }
    if dayLow > 0 {
        result += fmt.Sprintf("\nDay Low: $%.2f", dayLow)
    }
    if volume > 0 {
        result += fmt.Sprintf("\nVolume: %.0f", volume)
    }
    
    return result
}

func (s *MCPServer) formatQuoteData(quote map[string]interface{}, symbol string) string {
    currentPrice, _ := quote["regularMarketPrice"].(float64)
    previousClose, _ := quote["regularMarketPreviousClose"].(float64)
    dayHigh, _ := quote["regularMarketDayHigh"].(float64)
    dayLow, _ := quote["regularMarketDayLow"].(float64)
    volume, _ := quote["regularMarketVolume"].(float64)
    
    if currentPrice == 0 {
        return ""
    }
    
    change := currentPrice - previousClose
    changePercent := 0.0
    if previousClose != 0 {
        changePercent = (change / previousClose) * 100
    }
    
    result := fmt.Sprintf(`Stock: %s
Current Price: $%.2f
Previous Close: $%.2f
Change: $%.2f (%.2f%%)`,
        symbol, currentPrice, previousClose, change, changePercent)
    
    if dayHigh > 0 {
        result += fmt.Sprintf("\nDay High: $%.2f", dayHigh)
    }
    if dayLow > 0 {
        result += fmt.Sprintf("\nDay Low: $%.2f", dayLow)
    }
    if volume > 0 {
        result += fmt.Sprintf("\nVolume: %.0f", volume)
    }
    
    return result
}

func main() {
    server := NewMCPServer()
    
    // Stdio mode for Claude Desktop
    decoder := json.NewDecoder(os.Stdin)
    encoder := json.NewEncoder(os.Stdout)
    
    for {
        var req JSONRPCRequest
        if err := decoder.Decode(&req); err != nil {
            if err == io.EOF {
                break
            }
            // Log error but continue (don't crash on malformed input)
            log.Printf("JSON decode error: %v", err)
            continue
        }
        
        // Validate basic request structure
        if req.JSONRPC != "2.0" {
            req.JSONRPC = "2.0" // Set default
        }
        if req.ID == nil {
            req.ID = 0 // Set default ID
        }
        if req.Method == "" {
            // Send error response for missing method
            errorResp := JSONRPCResponse{
                JSONRPC: "2.0",
                ID:      req.ID,
                Error:   &RPCError{Code: -32600, Message: "Invalid Request - missing method"},
            }
            encoder.Encode(errorResp)
            continue
        }
        
        resp := server.HandleRequest(req)
        if err := encoder.Encode(resp); err != nil {
            log.Printf("Failed to encode response: %v", err)
        }
    }
}
