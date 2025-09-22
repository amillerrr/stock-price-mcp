# Use official Go image as build stage
FROM golang:1.25.1-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o stock-price-mcp .

# Use minimal alpine image for final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/stock-price-mcp .

# Make it executable
RUN chmod +x ./stock-price-mcp

# Expose any ports if needed (not required for stdio MCP)
# EXPOSE 8080

# Run the MCP server
CMD ["./stock-price-mcp"]
