package controller

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestSafeFileNameAcceptsPlainNames(t *testing.T) {
	for _, name := range []string{"report.txt", "archive.tar.gz", "photo-01.jpg"} {
		got, err := safeFileName(name)
		if err != nil {
			t.Fatalf("safeFileName(%q) returned error: %v", name, err)
		}
		if got != name {
			t.Fatalf("safeFileName(%q) = %q", name, got)
		}
	}
}

func TestSafeFileNameRejectsPathLikeNames(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../secret.txt", "nested/file.txt", `nested\file.txt`, "bad\x00name.txt"} {
		if got, err := safeFileName(name); err == nil {
			t.Fatalf("safeFileName(%q) = %q, expected error", name, got)
		}
	}
}

func TestUploadFilePathStaysUnderUploadDirectory(t *testing.T) {
	got, err := uploadFilePath("storage", "report.txt")
	if err != nil {
		t.Fatalf("uploadFilePath returned error: %v", err)
	}

	want := filepath.Join("storage", "report.txt")
	if got != want {
		t.Fatalf("uploadFilePath = %q, want %q", got, want)
	}
}

func TestSafeUploadRelativePathAcceptsFolderPaths(t *testing.T) {
	got, err := safeUploadRelativePath("docs/reports/summary.txt")
	if err != nil {
		t.Fatalf("safeUploadRelativePath returned error: %v", err)
	}
	if got != "docs/reports/summary.txt" {
		t.Fatalf("safeUploadRelativePath = %q", got)
	}
}

func TestSafeUploadRelativePathRejectsTraversal(t *testing.T) {
	for _, name := range []string{
		"../secret.txt",
		"docs/../secret.txt",
		"/absolute.txt",
		"docs//summary.txt",
		`docs\summary.txt`,
		"bad\x00name.txt",
	} {
		if got, err := safeUploadRelativePath(name); err == nil {
			t.Fatalf("safeUploadRelativePath(%q) = %q, expected error", name, got)
		}
	}
}

func TestUploadRelativePathStaysUnderUploadDirectory(t *testing.T) {
	got, err := uploadRelativePath("storage", "docs/summary.txt")
	if err != nil {
		t.Fatalf("uploadRelativePath returned error: %v", err)
	}

	want := filepath.Join("storage", "docs", "summary.txt")
	if got != want {
		t.Fatalf("uploadRelativePath = %q, want %q", got, want)
	}
}

func TestPathWithinUploadDir(t *testing.T) {
	dir := t.TempDir()
	inside := filepath.Join(dir, "docs", "summary.txt")
	outside := filepath.Join(filepath.Dir(dir), "summary.txt")

	if !pathWithinUploadDir(dir, inside) {
		t.Fatalf("expected %q to be inside %q", inside, dir)
	}
	if pathWithinUploadDir(dir, outside) {
		t.Fatalf("expected %q to be outside %q", outside, dir)
	}
	if pathWithinUploadDir(dir, dir) {
		t.Fatalf("upload root itself should not be treated as an uploaded entry")
	}
}

func TestStreamDirectoryAsZipIncludesNestedFiles(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "nested"), 0755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested", "summary.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile returned error: %v", err)
	}

	var buf bytes.Buffer
	if err := streamDirectoryAsZip(&buf, dir, "docs"); err != nil {
		t.Fatalf("streamDirectoryAsZip returned error: %v", err)
	}

	reader, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	if err != nil {
		t.Fatalf("NewReader returned error: %v", err)
	}

	seen := make(map[string]bool)
	for _, file := range reader.File {
		seen[file.Name] = true
	}
	for _, name := range []string{"docs/", "docs/nested/", "docs/nested/summary.txt"} {
		if !seen[name] {
			t.Fatalf("zip missing %q; entries: %#v", name, seen)
		}
	}
}
