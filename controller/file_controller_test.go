package controller

import (
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
	got, err := uploadFilePath("uploads", "report.txt")
	if err != nil {
		t.Fatalf("uploadFilePath returned error: %v", err)
	}

	want := filepath.Join("uploads", "report.txt")
	if got != want {
		t.Fatalf("uploadFilePath = %q, want %q", got, want)
	}
}
