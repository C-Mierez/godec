package envsync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sampleConfig = "package config\n\nimport (\n\t\"github.com/caarlos0/env/v11\"\n)\n\ntype ServerEnv struct {\n\tPort string `env:\"PORT\" envDefault:\"8080\"`\n\tEnv  string `env:\"ENV\" envDefault:\"development\"`\n}\n\ntype DatabaseEnv struct {\n\tURL string `env:\"DATABASE_URL\"`\n}\n\ntype Config struct {\n\tServer struct {\n\t\tServerEnv\n\t}\n\tDatabase struct {\n\t\tDatabaseEnv\n\t}\n}\n"

func TestLoadSchemaReadsEmbeddedStructs(t *testing.T) {
	t.Parallel()

	configPath := writeFile(t, t.TempDir(), "config.go", sampleConfig)
	schema, err := LoadSchema(configPath)
	if err != nil {
		t.Fatalf("LoadSchema() error = %v", err)
	}

	expected := []Entry{
		{Name: "PORT", Default: "8080", HasDefault: true},
		{Name: "ENV", Default: "development", HasDefault: true},
		{Name: "DATABASE_URL"},
	}

	if len(schema.Entries) != len(expected) {
		t.Fatalf("expected %d entries, got %d", len(expected), len(schema.Entries))
	}

	for i, entry := range expected {
		got := schema.Entries[i]
		if got != entry {
			t.Fatalf("entry %d = %#v, want %#v", i, got, entry)
		}
	}
}

func TestFixCreatesAndAppendsMissingKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := writeFile(t, dir, "config.go", sampleConfig)
	envPath := writeFile(t, dir, ".env", "PORT=9000\nLEGACY=1\n")
	examplePath := writeFile(t, dir, ".env.example", "PORT=8080\nLEGACY=1\n")

	report, err := Fix(Paths{ConfigPath: configPath, Files: []string{envPath, examplePath}})
	if err != nil {
		t.Fatalf("Fix() error = %v", err)
	}

	if report == nil || len(report.Files) != 2 {
		t.Fatalf("expected two file reports, got %#v", report)
	}

	envContent := readFile(t, envPath)
	if !strings.Contains(envContent, "PORT=9000") {
		t.Fatalf("expected existing value to be preserved, got %q", envContent)
	}
	if !strings.Contains(envContent, "# LEGACY=1") {
		t.Fatalf("expected stale env key to be commented out, got %q", envContent)
	}
	if !strings.Contains(envContent, "ENV=development") || !strings.Contains(envContent, "DATABASE_URL=") {
		t.Fatalf("expected missing keys to be appended, got %q", envContent)
	}

	exampleContent := readFile(t, examplePath)
	if !strings.Contains(exampleContent, "PORT=8080") || !strings.Contains(exampleContent, "ENV=development") || !strings.Contains(exampleContent, "DATABASE_URL=") {
		t.Fatalf("expected example file to be created from schema, got %q", exampleContent)
	}
	if !strings.Contains(exampleContent, "# LEGACY=1") {
		t.Fatalf("expected stale example key to be commented out, got %q", exampleContent)
	}
}

func TestCheckReportsMissingAndStaleKeys(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	configPath := writeFile(t, dir, "config.go", sampleConfig)
	envPath := writeFile(t, dir, ".env", "PORT=9000\n")
	examplePath := writeFile(t, dir, ".env.example", "PORT=8080\nENV=development\nDATABASE_URL=\nLEGACY=1\n")

	err := Check(Paths{ConfigPath: configPath, Files: []string{envPath, examplePath}})
	if err == nil {
		t.Fatal("expected Check() to fail")
	}

	checkErr, ok := err.(*CheckError)
	if !ok {
		t.Fatalf("expected *CheckError, got %T", err)
	}

	if len(checkErr.Issues) != 2 {
		t.Fatalf("expected issues for both files, got %#v", checkErr.Issues)
	}

	message := checkErr.Error()
	if !strings.Contains(message, "ENV") || !strings.Contains(message, "DATABASE_URL") || !strings.Contains(message, "LEGACY") {
		t.Fatalf("expected message to mention missing and stale keys, got %q", message)
	}
}

func writeFile(t *testing.T, dir, name, content string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}

	return path
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file %s: %v", path, err)
	}

	return string(content)
}
