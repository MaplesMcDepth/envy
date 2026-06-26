package main

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadEnvFileIgnoresCommentsAndBlankLines(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := strings.Join([]string{
		"",
		"# comment",
		"FOO=bar",
		"EMPTY=",
		"SPACED=value with spaces",
		"",
	}, "\n")

	if err := os.WriteFile(envFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	got := loadEnvFile(envFile)

	if got["FOO"] != "bar" {
		t.Fatalf("expected FOO to be loaded, got %q", got["FOO"])
	}
	if got["EMPTY"] != "" {
		t.Fatalf("expected EMPTY to be preserved, got %q", got["EMPTY"])
	}
	if got["SPACED"] != "value with spaces" {
		t.Fatalf("expected SPACED to keep spaces, got %q", got["SPACED"])
	}
}

func TestSaveEnvFileSortsKeysAndCreatesParentDir(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, "nested", ".env")

	if err := saveEnvFile(envFile, map[string]string{
		"ZED": "last",
		"ALPHA": "first",
	}); err != nil {
		t.Fatalf("save env file: %v", err)
	}

	data, err := os.ReadFile(envFile)
	if err != nil {
		t.Fatalf("read saved env file: %v", err)
	}

	if string(data) != "ALPHA=first\nZED=last\n" {
		t.Fatalf("unexpected file contents: %q", string(data))
	}
}

func TestIsSecretRecognizesSensitiveKeys(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{key: "API_KEY", want: true},
		{key: "dbPassword", want: true},
		{key: "AUTH_TOKEN", want: true},
		{key: "PUBLIC_PORT", want: false},
	}

	for _, tc := range tests {
		if got := isSecret(tc.key); got != tc.want {
			t.Fatalf("isSecret(%q) = %v, want %v", tc.key, got, tc.want)
		}
	}
}

func TestRunLoadEmitsExportStatements(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := strings.Join([]string{
		"# ignored",
		"FOO=bar",
		"QUOTED=value with spaces",
	}, "\n")

	if err := os.WriteFile(envFile, []byte(content), 0o644); err != nil {
		t.Fatalf("write env file: %v", err)
	}

	output := captureStdout(t, func() {
		runLoad(envFile)
	})

	want := "export FOO=\"bar\"\nexport QUOTED=\"value with spaces\"\n"
	if output != want {
		t.Fatalf("unexpected output: %q", output)
	}
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}

	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
	}()

	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		if _, err := io.Copy(&buf, r); err != nil {
			done <- ""
			return
		}
		done <- buf.String()
	}()

	fn()

	if err := w.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}
	out := <-done
	if err := r.Close(); err != nil {
		t.Fatalf("close reader: %v", err)
	}
	return out
}
