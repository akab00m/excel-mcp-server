#!/bin/bash
# Smoke: create workbook + write-autocreate against local excel-mcp (:3003).
# Reads EXCEL_MCP_HTTP_TOKEN from repo .env — does not print it.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TOKEN=$(grep -E '^EXCEL_MCP_HTTP_TOKEN=' "$ROOT/.env" | cut -d= -f2-)
AUTH="Authorization: Bearer ${TOKEN}"
WS="${EXCEL_E2E_ROOT:-/home/cursor-bridge/projects/cursor-cli-bridge/workspace}"
E2E="$WS/e2e"
mkdir -p "$E2E"

curl -sS -D /tmp/mcp_hdrs.txt -o /tmp/mcp_body.txt \
  -H "$AUTH" \
  -H 'Content-Type: application/json' \
  -H 'Accept: application/json, text/event-stream' \
  -d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"smoke","version":"0.0.1"}}}' \
  http://127.0.0.1:3003/mcp >/dev/null
SESSION=$(awk -F': ' 'tolower($1)=="mcp-session-id"{print $2}' /tmp/mcp_hdrs.txt | tr -d '\r')
hdr=(-H "$AUTH" -H 'Content-Type: application/json' -H 'Accept: application/json, text/event-stream' -H "Mcp-Session-Id: $SESSION")
curl -sS "${hdr[@]}" -d '{"jsonrpc":"2.0","method":"notifications/initialized"}' http://127.0.0.1:3003/mcp >/dev/null || true

LIST=$(curl -sS "${hdr[@]}" -d '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' http://127.0.0.1:3003/mcp)
echo "$LIST" | grep -q excel_create_workbook && echo TOOLS_HAS_CREATE=yes || { echo TOOLS_HAS_CREATE=no; exit 1; }

rm -f "$E2E/new_deploy_smoke.xlsx" "$E2E/probe_autocreate.xlsx"
curl -sS "${hdr[@]}" -d "{\"jsonrpc\":\"2.0\",\"id\":3,\"method\":\"tools/call\",\"params\":{\"name\":\"excel_create_workbook\",\"arguments\":{\"fileAbsolutePath\":\"$E2E/new_deploy_smoke.xlsx\",\"sheetName\":\"Report\"}}}" http://127.0.0.1:3003/mcp >/dev/null
test -f "$E2E/new_deploy_smoke.xlsx" && echo CREATE_OK=yes || { echo CREATE_OK=no; exit 1; }

curl -sS "${hdr[@]}" -d "{\"jsonrpc\":\"2.0\",\"id\":4,\"method\":\"tools/call\",\"params\":{\"name\":\"excel_write_to_sheet\",\"arguments\":{\"fileAbsolutePath\":\"$E2E/probe_autocreate.xlsx\",\"sheetName\":\"Report\",\"newSheet\":false,\"range\":\"A1:A1\",\"values\":[[\"E2E_AUTOCREATE_OK\"]]}}}" http://127.0.0.1:3003/mcp >/dev/null
test -f "$E2E/probe_autocreate.xlsx" && echo WRITE_OK=yes || { echo WRITE_OK=no; exit 1; }

DESC=$(curl -sS "${hdr[@]}" -d "{\"jsonrpc\":\"2.0\",\"id\":5,\"method\":\"tools/call\",\"params\":{\"name\":\"excel_describe_sheets\",\"arguments\":{\"fileAbsolutePath\":\"$E2E/probe_autocreate.xlsx\"}}}")
echo "$DESC" | grep -q Report && echo DESC_HAS_Report=yes || { echo DESC_HAS_Report=no; exit 1; }
echo SMOKE_PASS
