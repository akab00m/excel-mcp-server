# AGENTS.md

## Cursor Cloud specific instructions

This repo is the **Excel MCP Server**: a Go binary that speaks the MCP protocol over **stdio** (default) or **Streamable HTTP** (`--transport http` / `EXCEL_MCP_TRANSPORT=http`), plus a thin TypeScript launcher (`dist/launcher.js`) that spawns the correct prebuilt Go binary. Standard commands live in `README.md` and `CLAUDE.md`; notes below are only the non-obvious caveats.

### Services
- **stdio**: local MCP clients spawn the process; no network port.
- **http**: listens on `EXCEL_MCP_HTTP_ADDR` (default `:8080`), MCP at `EXCEL_MCP_HTTP_PATH` (default `/mcp`), liveness at `/healthz`. Optional `EXCEL_MCP_HTTP_TOKEN` (Bearer). Docker image defaults to HTTP for container-to-container use. On Linux it uses the cross-platform `excelize` backend; the Windows-only OLE live-editing and `excel_screen_capture` features cannot run here.

### Build / run caveats
- `npm run build` runs `goreleaser build --snapshot --clean && tsc`. `goreleaser` (v2) is required and is installed at `~/go/bin` (on `PATH` via `~/.bashrc`), not via `npm -g` (the npm global prefix is root-owned `/` and fails). If `goreleaser` is missing, install with `go install github.com/goreleaser/goreleaser/v2@latest`.
- The launcher resolves the binary at `dist/excel-mcp-server_<os>_<arch>/` (Linux: `dist/excel-mcp-server_linux_amd64_v1/excel-mcp-server`). You must run `npm run build` before using `dist/launcher.js` / `npm run debug`, otherwise the launcher throws because the binary dir doesn't exist.
- For fast Go-only iteration, bypass goreleaser and the launcher: `go run ./cmd/excel-mcp-server` serves MCP directly over stdio.

### Testing caveats
- There are currently **no Go test files**, so `go test ./...` passes trivially with `[no test files]`.
- `gofmt -l .` flags `internal/excel/pagination.go` (pre-existing); leave it as-is.
- `excel.OpenFile` uses `excelize.OpenFile`, which requires the file to **already exist**. To exercise `excel_write_to_sheet`, first create a blank workbook (e.g. a small Go program using the vendored `github.com/xuri/excelize/v2` `NewFile().SaveAs(path)`), then write/read against it.
- End-to-end smoke test: pipe line-delimited JSON-RPC into `node dist/launcher.js` (or the raw binary). Keep stdin open until responses arrive — closing stdin immediately (a plain `cat file | ...` pipe) can EOF the server before `tools/call` responses flush. A Node child-process driver that awaits each response works reliably.
