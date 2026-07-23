package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/negokaz/excel-mcp-server/internal/server"
)

var (
	version = "dev"
)

func main() {
	transport := envOr("EXCEL_MCP_TRANSPORT", "stdio")
	httpAddr := envOr("EXCEL_MCP_HTTP_ADDR", ":8080")
	httpPath := envOr("EXCEL_MCP_HTTP_PATH", "/mcp")
	httpToken := os.Getenv("EXCEL_MCP_HTTP_TOKEN")

	flag.StringVar(&transport, "transport", transport, "MCP transport: stdio or http (env EXCEL_MCP_TRANSPORT)")
	flag.StringVar(&transport, "t", transport, "MCP transport: stdio or http (shorthand)")
	flag.StringVar(&httpAddr, "addr", httpAddr, "HTTP listen address when transport=http (env EXCEL_MCP_HTTP_ADDR)")
	flag.StringVar(&httpPath, "path", httpPath, "HTTP MCP endpoint path when transport=http (env EXCEL_MCP_HTTP_PATH)")
	flag.Parse()

	s := server.New(version)
	err := s.Start(transport, server.HTTPConfig{
		Addr:  httpAddr,
		Path:  httpPath,
		Token: strings.TrimSpace(httpToken),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start the server: %v\n", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && strings.TrimSpace(v) != "" {
		return v
	}
	return fallback
}
