package model

import (
	"database/sql"
	"fmt"
	"testing"
)

func TestSQLiteFileInfoLifecycle(t *testing.T) {
	dsn := fmt.Sprintf("file:%s/test.db?_pragma=busy_timeout(5000)", t.TempDir())
	if err := InitDB("sqlite", dsn); err != nil {
		t.Fatalf("InitDB returned error: %v", err)
	}
	t.Cleanup(func() {
		if DB != nil {
			_ = DB.Close()
		}
	})

	if err := AddFileEntryInfo("docs", "storage/docs", 0, true); err != nil {
		t.Fatalf("AddFileEntryInfo returned error: %v", err)
	}
	files, err := GetAllFileInfos()
	if err != nil {
		t.Fatalf("GetAllFileInfos returned error: %v", err)
	}
	if len(files) != 1 || files[0].FileName != "docs" || !files[0].IsDir {
		t.Fatalf("unexpected files: %#v", files)
	}

	id := fmt.Sprint(files[0].ID)
	if err := RenameFileInfoByID(id, "documents", "storage/documents"); err != nil {
		t.Fatalf("RenameFileInfoByID returned error: %v", err)
	}
	file, err := GetFileInfoByID(id)
	if err != nil || file.FileName != "documents" {
		t.Fatalf("GetFileInfoByID = %#v, %v", file, err)
	}
	if err := DeleteFileInfoByID(id); err != nil {
		t.Fatalf("DeleteFileInfoByID returned error: %v", err)
	}
}

func TestSQLiteMigratesLegacyFileInfoTable(t *testing.T) {
	dsn := fmt.Sprintf("file:%s/legacy.db?_pragma=busy_timeout(5000)", t.TempDir())
	legacyDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("sql.Open returned error: %v", err)
	}
	_, err = legacyDB.Exec(`CREATE TABLE file_info (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		file_name TEXT NOT NULL,
		file_path TEXT NOT NULL,
		file_size INTEGER,
		update_time DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		t.Fatalf("create legacy table returned error: %v", err)
	}
	if err := legacyDB.Close(); err != nil {
		t.Fatalf("legacy database close returned error: %v", err)
	}

	if err := InitDB("sqlite", dsn); err != nil {
		t.Fatalf("InitDB migration returned error: %v", err)
	}
	t.Cleanup(func() { _ = DB.Close() })
	if err := AddFileEntryInfo("legacy.txt", "storage/legacy.txt", 1, false); err != nil {
		t.Fatalf("insert after migration returned error: %v", err)
	}
}
