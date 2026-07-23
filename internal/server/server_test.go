package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNormalizeHTTPConfigRequiresToken(t *testing.T) {
	_, err := normalizeHTTPConfig(HTTPConfig{})
	if err == nil {
		t.Fatal("expected error when token is empty")
	}
	if !strings.Contains(err.Error(), "EXCEL_MCP_HTTP_TOKEN") {
		t.Fatalf("unexpected error: %v", err)
	}

	cfg, err := normalizeHTTPConfig(HTTPConfig{Token: "  secret  "})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "secret" {
		t.Fatalf("token not trimmed: %q", cfg.Token)
	}
	if cfg.Addr != ":8080" || cfg.Path != "/mcp" {
		t.Fatalf("unexpected defaults: addr=%q path=%q", cfg.Addr, cfg.Path)
	}
}

func TestStartHTTPRequiresToken(t *testing.T) {
	s := New("test")
	err := s.StartHTTP(HTTPConfig{Addr: "127.0.0.1:0"})
	if err == nil {
		t.Fatal("expected StartHTTP to fail without token")
	}
	if !strings.Contains(err.Error(), "EXCEL_MCP_HTTP_TOKEN") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBearerAuthMiddleware(t *testing.T) {
	s := New("test")
	cfg, err := normalizeHTTPConfig(HTTPConfig{Token: "test-token", Path: "/mcp"})
	if err != nil {
		t.Fatal(err)
	}
	handler := s.newHTTPHandler(cfg)

	t.Run("healthz without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
		if body := rec.Body.String(); body != "ok" {
			t.Fatalf("body=%q", body)
		}
	})

	t.Run("mcp without token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("mcp wrong token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Authorization", "Bearer wrong")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
		}
	})

	t.Run("mcp with valid token reaches MCP handler", func(t *testing.T) {
		body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`
		req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json, text/event-stream")
		req.Header.Set("Authorization", "Bearer test-token")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code == http.StatusUnauthorized {
			t.Fatal("valid token must not return 401")
		}
		respBody, _ := io.ReadAll(rec.Body)
		if !strings.Contains(string(respBody), "excel-mcp-server") && rec.Code != http.StatusOK && rec.Code != http.StatusAccepted {
			// Accept any non-401 MCP response; initialize should succeed with 200 JSON.
			t.Fatalf("unexpected MCP response status=%d body=%s", rec.Code, string(respBody))
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200 from initialize, got %d body=%s", rec.Code, string(respBody))
		}
		if !strings.Contains(string(respBody), `"name":"excel-mcp-server"`) {
			t.Fatalf("expected serverInfo in body: %s", string(respBody))
		}
	})
}

func TestStartRejectsUnknownTransport(t *testing.T) {
	s := New("test")
	err := s.Start("udp", HTTPConfig{Token: "x"})
	if err == nil || !strings.Contains(err.Error(), "invalid transport") {
		t.Fatalf("unexpected error: %v", err)
	}
}
