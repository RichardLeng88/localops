package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInspectDirectoryMarker(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}

	if got.ProjectPath != filepath.Clean(root) {
		t.Fatalf("ProjectPath = %q", got.ProjectPath)
	}
	if got.GitMarkerPath != filepath.Join(root, ".git") {
		t.Fatalf("GitMarkerPath = %q", got.GitMarkerPath)
	}
	if got.GitMarkerType != "directory" {
		t.Fatalf("GitMarkerType = %q", got.GitMarkerType)
	}
}

func TestInspectRegularFileMarker(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".git"), []byte("gitdir: elsewhere"), 0o644); err != nil {
		t.Fatal(err)
	}

	got, err := Inspect(root)
	if err != nil {
		t.Fatal(err)
	}
	if got.GitMarkerType != "regular-file" {
		t.Fatalf("GitMarkerType = %q", got.GitMarkerType)
	}
}

func TestInspectRejectsEmptyAndRelativePaths(t *testing.T) {
	for _, path := range []string{"", "."} {
		t.Run(path, func(t *testing.T) {
			got, err := Inspect(path)
			if err == nil {
				t.Fatal("Inspect succeeded")
			}
			if got != (Inspection{}) {
				t.Fatalf("Inspection = %#v", got)
			}
		})
	}
}

func TestInspectRejectsInvalidSelectedPath(t *testing.T) {
	parent := t.TempDir()
	filePath := filepath.Join(parent, "file")
	if err := os.WriteFile(filePath, nil, 0o644); err != nil {
		t.Fatal(err)
	}

	for name, path := range map[string]string{
		"missing":       filepath.Join(parent, "missing"),
		"file":          filePath,
		"no Git marker": parent,
	} {
		t.Run(name, func(t *testing.T) {
			got, err := Inspect(path)
			if err == nil {
				t.Fatal("Inspect succeeded")
			}
			if got != (Inspection{}) {
				t.Fatalf("Inspection = %#v", got)
			}
		})
	}
}

func TestInspectDoesNotSearchSiblingDirectories(t *testing.T) {
	parent := t.TempDir()
	selected := filepath.Join(parent, "selected")
	sibling := filepath.Join(parent, "sibling")
	if err := os.MkdirAll(filepath.Join(sibling, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(selected, 0o755); err != nil {
		t.Fatal(err)
	}

	got, err := Inspect(selected)
	if err == nil {
		t.Fatal("Inspect succeeded")
	}
	if got != (Inspection{}) {
		t.Fatalf("Inspection = %#v", got)
	}
}
