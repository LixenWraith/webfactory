package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Storage handles all file system operations for the application
type Storage struct {
	sourcePath string
	targetPath string
}

// New creates a Storage instance with the given root path
func New(sourcePath, targetPath string) *Storage {
	return &Storage{
		sourcePath: sourcePath,
		targetPath: targetPath,
	}
}

// ListBlueprints reads a blueprint file from disk
func (s *Storage) ListBlueprints() (map[string]string, error) {
	blueprints := make(map[string]string)
	blueprintsDir := filepath.Join(s.sourcePath, "blueprints")

	err := filepath.Walk(blueprintsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".blueprint") {
			return err
		}

		rel, err := filepath.Rel(blueprintsDir, path)
		if err != nil {
			return err
		}

		// Get the parent directory as prefix
		dir := filepath.Dir(path)
		dirRel, err := filepath.Rel(s.sourcePath, dir)
		if err != nil {
			return err
		}
		prefix := strings.Split(dirRel, string(filepath.Separator))[0]

		outputPath := filepath.Base(path[:len(path)-len(".blueprint")])
		outputPath = filepath.Join(prefix, outputPath)

		blueprints[rel] = outputPath
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("scanning blueprints: %w", err)
	}

	return blueprints, nil
}

// ReadBlueprint reads a blueprint file from disk
func (s *Storage) ReadBlueprint(path string) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.sourcePath, "blueprints", path))
}

// ReadComponent reads a component file (template, css, js) from disk
func (s *Storage) ReadComponent(componentPath, filename string) ([]byte, error) {
	fullPath := filepath.Join(s.sourcePath, "components", componentPath, filename)
	return os.ReadFile(fullPath)
}

// ListComponentFiles lists all files in a component directory, optionally filtered by extension
func (s *Storage) ListComponentFiles(componentPath string, ext string) ([]string, error) {
	dir := filepath.Join(s.sourcePath, "components", componentPath)
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walking component files: %w", err)
		}
		if !info.IsDir() {
			if ext != "" && filepath.Ext(path) != ext {
				return nil
			}
			rel, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("listing component files: %w", err)
	}

	return files, nil
}

// FindTemplateFile finds the single HTML template file in component directory
func (s *Storage) FindTemplateFile(componentPath string) (string, error) {
	files, err := s.ListComponentFiles(componentPath, ".html")
	if err != nil {
		return "", fmt.Errorf("listing HTML files: %w", err)
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no HTML template found in component %s", componentPath)
	}
	if len(files) > 1 {
		return "", fmt.Errorf("multiple HTML templates found in component %s", componentPath)
	}

	return files[0], nil
}

// GetTargetPath returns the absolute path to target directory
func (s *Storage) GetTargetPath() string {
	return s.targetPath
}

// WriteOutput writes the generated site files
func (s *Storage) WriteOutput(outputPath string, files map[string][]byte) error {
	for path, content := range files {
		fullPath := filepath.Join(outputPath, path)

		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(fullPath, content, 0644); err != nil {
			return err
		}
	}
	return nil
}