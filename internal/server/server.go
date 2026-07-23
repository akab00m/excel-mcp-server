package server

import (
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strings"

	"github.com/mark3labs/mcp-go/server"
	"github.com/negokaz/excel-mcp-server/internal/tools"
)

type ExcelServer struct {
	server *server.MCPServer
}

type HTTPConfig struct {
	// Addr is the listen address, e.g. ":8080" or "0.0.0.0:8080".
	Addr string
	// Path is the MCP endpoint path, e.g. "/mcp".
	Path string
	// Token, if non-empty, requires Authorization: Bearer <token> on MCP requests.
	Token string
}

func New(version string) *ExcelServer {
	s := &ExcelServer{}
	s.server = server.NewMCPServer(
		"excel-mcp-server",
		version,
	)
	tools.AddExcelDescribeSheetsTool(s.server)
	tools.AddExcelReadSheetTool(s.server)
	if runtime.GOOS == "windows" {
		tools.AddExcelScreenCaptureTool(s.server)
	}
	tools.AddExcelWriteToSheetTool(s.server)
	tools.AddExcelCreateTableTool(s.server)
	tools.AddExcelCopySheetTool(s.server)
	tools.AddExcelFormatRangeTool(s.server)
	return s
}

func (s *ExcelServer) StartStdio() error {
	return server.ServeStdio(s.server)
}

// StartHTTP serves the MCP Streamable HTTP transport for container-to-container use.
// Default client URL: http://<host><addr-port><path> (e.g. http://excel-mcp:8080/mcp).
func (s *ExcelServer) StartHTTP(cfg HTTPConfig) error {
	if cfg.Addr == "" {
		cfg.Addr = ":8080"
	}
	if cfg.Path == "" {
		cfg.Path = "/mcp"
	}
	if !strings.HasPrefix(cfg.Path, "/") {
		cfg.Path = "/" + cfg.Path
	}
	cfg.Path = strings.TrimSuffix(cfg.Path, "/")
	if cfg.Path == "" {
		cfg.Path = "/mcp"
	}

	mcpHandler := server.NewStreamableHTTPServer(s.server)

	mux := http.NewServeMux()
	mux.Handle(cfg.Path, mcpHandler)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	var handler http.Handler = mux
	if cfg.Token != "" {
		handler = bearerAuthMiddleware(cfg.Token, mux)
	}

	log.Printf("excel-mcp-server listening on %s (MCP %s, healthz /healthz)", cfg.Addr, cfg.Path)
	return http.ListenAndServe(cfg.Addr, handler)
}

func bearerAuthMiddleware(token string, next http.Handler) http.Handler {
	expected := []byte("Bearer " + token)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Liveness probe stays unauthenticated for orchestrators.
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}
		auth := []byte(r.Header.Get("Authorization"))
		if subtle.ConstantTimeCompare(auth, expected) != 1 {
			w.Header().Set("WWW-Authenticate", `Bearer realm="excel-mcp-server"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Start selects transport: "stdio" (default) or "http".
func (s *ExcelServer) Start(transport string, httpCfg HTTPConfig) error {
	switch strings.ToLower(strings.TrimSpace(transport)) {
	case "", "stdio":
		return s.StartStdio()
	case "http":
		return s.StartHTTP(httpCfg)
	default:
		return fmt.Errorf("invalid transport %q (want stdio or http)", transport)
	}
}
