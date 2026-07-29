// Package project inspects one explicitly selected project directory.
package project

import (
	"fmt"
	"os"
	"path/filepath"
)

// Inspection identifies the selected project and its direct Git marker.
type Inspection struct {
	ProjectPath   string
	GitMarkerPath string
	GitMarkerType string
}

// Inspect checks one absolute path without searching or reading Git metadata.
func Inspect(projectPath string) (Inspection, error) {
	if projectPath == "" {
		return Inspection{}, fmt.Errorf("project path is required")
	}
	if !filepath.IsAbs(projectPath) {
		return Inspection{}, fmt.Errorf("project path must be absolute")
	}

	cleanPath := filepath.Clean(projectPath)
	info, err := os.Stat(cleanPath)
	if err != nil {
		return Inspection{}, fmt.Errorf("inspect project path: %w", err)
	}
	if !info.IsDir() {
		return Inspection{}, fmt.Errorf("project path is not a directory")
	}

	markerPath := filepath.Join(cleanPath, ".git")
	markerInfo, err := os.Lstat(markerPath)
	if err != nil {
		return Inspection{}, fmt.Errorf("inspect Git marker: %w", err)
	}

	var markerType string
	switch {
	case markerInfo.IsDir():
		markerType = "directory"
	case markerInfo.Mode().IsRegular():
		markerType = "regular-file"
	default:
		return Inspection{}, fmt.Errorf("Git marker is not a directory or regular file")
	}

	return Inspection{
		ProjectPath:   cleanPath,
		GitMarkerPath: markerPath,
		GitMarkerType: markerType,
	}, nil
}
