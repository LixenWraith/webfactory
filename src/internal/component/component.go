package component

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"webfactory/src/internal/storage"
)

// Component represents a parsed and loaded component
type Component struct {
	Path     string            // Dot-separated path (e.g., "simple" or "composite.layout")
	Template []byte            // Raw template content
	Styles   []byte            // Combined CSS content
	Scripts  map[string][]byte // JS content for each file
	Children map[string]*Component
}

// Registry manages all loaded components
type Registry struct {
	store  *storage.Storage
	loaded map[string]*Component // key is "path.name"
}

// New creates a new component registry
func New(store *storage.Storage) *Registry {
	return &Registry{
		store:  store,
		loaded: make(map[string]*Component),
	}
}

// Load loads a component and its assets
func (r *Registry) Load(path string) (*Component, error) {
	if comp, exists := r.loaded[path]; exists {
		return comp, nil
	}

	comp := &Component{
		Path:     path,
		Children: make(map[string]*Component),
		Scripts:  make(map[string][]byte),
	}

	parts := strings.Split(path, ".")
	fsPath := filepath.Join(parts...)

	// Find and load HTML template
	templateFile, err := r.store.FindTemplateFile(fsPath)
	if err != nil {
		return nil, fmt.Errorf("finding template: %w", err)
	}
	template, err := r.store.ReadComponent(fsPath, templateFile)
	if err != nil {
		return nil, err
	}
	comp.Template = template

	// Load all CSS files and combine
	cssFiles, err := r.store.ListComponentFiles(fsPath, ".css")
	if err != nil {
		return nil, fmt.Errorf("listing CSS files: %w", err)
	}
	var cssContent bytes.Buffer
	for _, file := range cssFiles {
		content, err := r.store.ReadComponent(fsPath, file)
		if err != nil {
			return nil, fmt.Errorf("reading CSS %s: %w", file, err)
		}
		cssContent.Write(content)
		cssContent.WriteByte('\n')
	}
	comp.Styles = cssContent.Bytes()

	// Load all JS files
	jsFiles, err := r.store.ListComponentFiles(fsPath, ".js")
	if err != nil {
		return nil, fmt.Errorf("listing JS files: %w", err)
	}
	for _, file := range jsFiles {
		content, err := r.store.ReadComponent(fsPath, file)
		if err != nil {
			return nil, fmt.Errorf("reading JS %s: %w", file, err)
		}
		comp.Scripts[file] = content
	}

	r.loaded[path] = comp
	return comp, nil
}

// Get returns a loaded component
func (r *Registry) Get(path string) *Component {
	return r.loaded[path]
}

// Each iterates over all loaded components
func (r *Registry) Each(fn func(comp *Component)) {
	for _, comp := range r.loaded {
		fn(comp)
	}
}

// func (r *Registry) Cleanup() {
// 	r.loaded = nil
// }