package tools

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/negokaz/excel-mcp-server/internal/excel"
)

func TestCreateWorkbook(t *testing.T) {
	path := filepath.Join(t.TempDir(), "out", "book.xlsx")
	res, err := createWorkbook(path, "Report")
	if err != nil {
		t.Fatalf("createWorkbook: %v", err)
	}
	if res == nil || res.IsError {
		t.Fatalf("unexpected tool result: %+v", res)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("file not created: %v", err)
	}

	wb, closeFn, err := excel.OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer closeFn()
	sheet, err := wb.FindSheet("Report")
	if err != nil {
		t.Fatalf("FindSheet Report: %v", err)
	}
	sheet.Release()

	if _, err := createWorkbook(path, ""); err == nil {
		t.Fatal("expected error on existing file")
	}
}

func TestWriteSheetCreatesMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "auto.xlsx")
	res, err := writeSheet(path, "Sheet1", false, "A1:B1", [][]any{{"hello", "world"}})
	if err != nil {
		t.Fatalf("writeSheet: %v", err)
	}
	if res == nil || res.IsError {
		t.Fatalf("unexpected tool result: %+v", res)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected auto-created workbook: %v", err)
	}

	wb, closeFn, err := excel.OpenFile(path)
	if err != nil {
		t.Fatalf("re-open: %v", err)
	}
	defer closeFn()
	sheet, err := wb.FindSheet("Sheet1")
	if err != nil {
		t.Fatalf("FindSheet: %v", err)
	}
	defer sheet.Release()
}

func TestWriteSheetCreatesMissingFileWithCustomSheet(t *testing.T) {
	path := filepath.Join(t.TempDir(), "custom.xlsx")
	res, err := writeSheet(path, "Report", false, "A1:A1", [][]any{{"x"}})
	if err != nil {
		t.Fatalf("writeSheet: %v", err)
	}
	if res == nil || res.IsError {
		t.Fatalf("unexpected tool result: %+v", res)
	}

	wb, closeFn, err := excel.OpenFile(path)
	if err != nil {
		t.Fatalf("re-open: %v", err)
	}
	defer closeFn()
	sheet, err := wb.FindSheet("Report")
	if err != nil {
		t.Fatalf("FindSheet Report: %v", err)
	}
	sheet.Release()
}
