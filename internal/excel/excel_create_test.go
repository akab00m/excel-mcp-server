package excel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCreateFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "book.xlsx")

	if err := CreateFile(path, "Data"); err != nil {
		t.Fatalf("CreateFile: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file on disk: %v", err)
	}

	wb, closeFn, err := OpenFile(path)
	if err != nil {
		t.Fatalf("OpenFile: %v", err)
	}
	defer closeFn()

	sheets, err := wb.GetSheets()
	if err != nil {
		t.Fatalf("GetSheets: %v", err)
	}
	if len(sheets) != 1 {
		t.Fatalf("expected 1 sheet, got %d", len(sheets))
	}
	name, err := sheets[0].Name()
	sheets[0].Release()
	if err != nil {
		t.Fatalf("Name: %v", err)
	}
	if name != "Data" {
		t.Fatalf("sheet name = %q, want Data", name)
	}

	if err := CreateFile(path, ""); err == nil {
		t.Fatal("expected error when file already exists")
	}
}

func TestOpenFileMissingPathHint(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "missing.xlsx")
	_, _, err := OpenFile(missing)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
	msg := err.Error()
	if !strings.Contains(msg, "shared volume") || !strings.Contains(msg, missing) {
		t.Fatalf("error should mention shared volume and path, got: %v", err)
	}
}

func TestOpenFileOrCreate(t *testing.T) {
	path := filepath.Join(t.TempDir(), "new.xlsx")
	wb, closeFn, err := OpenFileOrCreate(path)
	if err != nil {
		t.Fatalf("OpenFileOrCreate: %v", err)
	}
	defer closeFn()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected created file: %v", err)
	}
	_ = wb
}
