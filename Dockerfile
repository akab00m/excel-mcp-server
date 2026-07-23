# Build from this fork's sources (HTTP-capable binary).
FROM golang:1.24-bookworm AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=docker" -o /out/excel-mcp-server ./cmd/excel-mcp-server

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app
COPY --from=build /out/excel-mcp-server /app/excel-mcp-server

# Container-to-container MCP defaults (override with env if needed).
ENV EXCEL_MCP_TRANSPORT=http \
    EXCEL_MCP_HTTP_ADDR=:8080 \
    EXCEL_MCP_HTTP_PATH=/mcp

EXPOSE 8080

USER nonroot:nonroot
ENTRYPOINT ["/app/excel-mcp-server"]
