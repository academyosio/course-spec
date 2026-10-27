// Package testcourse copies the test course into a temporary directory and
// edits its files, for tests.
package testcourse

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// Source returns the path of the test course in testdata.
func Source() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("testcourse: no caller information")
	}
	return filepath.Join(filepath.Dir(file), "..", "..", "testdata", "courses", "web-apis")
}

// Copy copies the test course into a new temporary directory and returns
// the course directory.
func Copy(t *testing.T) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "web-apis")
	if err := os.CopyFS(dst, os.DirFS(Source())); err != nil {
		t.Fatalf("copy test course: %v", err)
	}
	return dst
}

// Write writes a file in the course directory.
func Write(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

// Read reads a file in the course directory.
func Read(t *testing.T, dir, rel string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, rel))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// Remove deletes a file or directory in the course directory.
func Remove(t *testing.T, dir, rel string) {
	t.Helper()
	if err := os.RemoveAll(filepath.Join(dir, rel)); err != nil {
		t.Fatal(err)
	}
}

// EditJSON decodes a JSON file into a generic value, lets edit change it,
// and writes it back.
func EditJSON(t *testing.T, dir, rel string, edit func(m map[string]any)) {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(Read(t, dir, rel)), &m); err != nil {
		t.Fatalf("%s: %v", rel, err)
	}
	edit(m)
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	Write(t, dir, rel, string(data))
}

// Questions returns a quiz's questions from a generic quiz value.
func Questions(m map[string]any) []any { return m["questions"].([]any) }

// Question returns question i of a generic quiz value.
func Question(m map[string]any, i int) map[string]any { return Questions(m)[i].(map[string]any) }
