# Build from this fork's sources (HTTP-capable binary).
FROM golang:1.24-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=docker" -o /out/excel-mcp-server ./cmd/excel-mcp-server

# Numeric non-root default; override with compose / docker run --user so the
# process can read/write a shared volume owned by the agent user.
FROM gcr.io/distroless/static-debian12

WORKDIR /app
COPY --from=build /out/excel-mcp-server /app/excel-mcp-server

# Container-to-container MCP defaults.
# EXCEL_MCP_HTTP_TOKEN must be provided at runtime (required for HTTP).
ENV EXCEL_MCP_TRANSPORT=http \
    EXCEL_MCP_HTTP_ADDR=:8080 \
    EXCEL_MCP_HTTP_PATH=/mcp

EXPOSE 8080

USER 1000:1000
ENTRYPOINT ["/app/excel-mcp-server"]
