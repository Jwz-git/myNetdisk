package config

import (
	"strings"
	"testing"
)

func validTestConfig() *Config {
	return &Config{
		Database: DatabaseConfig{Driver: "sqlite", Path: "data/test.db"},
		Server:   ServerConfig{Port: "8080"},
		Admin:    AdminConfig{Username: "admin", Password: "password"},
		Upload:   UploadConfig{Directory: "storage"},
	}
}

func TestValidateConfigSupportsSQLiteWithoutMySQLFields(t *testing.T) {
	cfg := validTestConfig()
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig returned error: %v", err)
	}
	if got := cfg.GetDatabaseDriver(); got != "sqlite" {
		t.Fatalf("GetDatabaseDriver() = %q, want sqlite", got)
	}
	if got := cfg.GetDSN(); !strings.HasPrefix(got, "file:data/test.db?") {
		t.Fatalf("GetDSN() = %q", got)
	}
}

func TestValidateConfigDefaultsToSQLite(t *testing.T) {
	cfg := validTestConfig()
	cfg.Database = DatabaseConfig{}
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig returned error: %v", err)
	}
	if cfg.Database.Driver != "sqlite" || cfg.Database.Path != "data/personal_disk.db" {
		t.Fatalf("database defaults = %#v", cfg.Database)
	}
}

func TestValidateConfigSupportsMySQL(t *testing.T) {
	cfg := validTestConfig()
	cfg.Database = DatabaseConfig{
		Driver: "MySQL", Host: "127.0.0.1", Port: "3306", User: "root",
		Name: "personal_disk", Charset: "utf8mb4", ParseTime: true, Location: "Local",
	}
	if err := validateConfig(cfg); err != nil {
		t.Fatalf("validateConfig returned error: %v", err)
	}
	if got := cfg.GetDatabaseDriver(); got != "mysql" {
		t.Fatalf("GetDatabaseDriver() = %q, want mysql", got)
	}
	if got := cfg.GetDSN(); !strings.Contains(got, "tcp(127.0.0.1:3306)/personal_disk") {
		t.Fatalf("GetDSN() = %q", got)
	}
}

func TestValidateConfigRejectsUnknownDatabase(t *testing.T) {
	cfg := validTestConfig()
	cfg.Database.Driver = "postgres"
	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig accepted unsupported database")
	}
}
