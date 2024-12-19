package builder

import (
	"fmt"
	"path/filepath"
	"strings"

	"webfactory/src/internal/blueprint"
	"webfactory/src/internal/component"
	"webfactory/src/internal/storage"
	"webfactory/src/internal/template"
)

// Builder orchestrates the site generation process
type Builder struct {
	store *storage.Storage
}

// New creates a new Builder instance
func New(sourcePath, outputPath string) *Builder {
	store := storage.New(sourcePath, outputPath)

	return &Builder{
		store: store,
	}
}

// Build processes all blueprints and generates the site
func (b *Builder) Build() error {
	// Get list of blueprints
	blueprints, err := b.store.ListBlueprints()
	if err != nil {
		return fmt.Errorf("finding blueprints: %w", err)
	}

	// Process each blueprint
	for path, outputRel := range blueprints {
		if err := b.processBlueprint(path, outputRel); err != nil {
			return fmt.Errorf("processing blueprint %s: %w", path, err)
		}
	}

	return nil
}

// processBlueprint handles a single blueprint file
func (b *Builder) processBlueprint(path, outputRel string) error {
	// Read and parse blueprint
	content, err := b.store.ReadBlueprint(path)
	if err != nil {
		return fmt.Errorf("reading blueprint: %w", err)
	}

	registry := component.New(b.store)
	processor := template.New(registry)

	tree, err := blueprint.New(string(content))
	if err != nil {
		return fmt.Errorf("parsing blueprint: %w", err)
	}

	// Load components referenced in blueprint
	var loadComponents func(*blueprint.Node) error
	loadComponents = func(node *blueprint.Node) error {
		if node == nil {
			return nil
		}

		if node.Block.ID != -1 {
			_, err := registry.Load(node.Block.Path)
			if err != nil {
				return fmt.Errorf("loading component %s: %w", node.Block.Path, err)
			}
		}

		for _, child := range node.Children {
			if err := loadComponents(child); err != nil {
				return err
			}
		}
		return nil
	}

	if err := loadComponents(tree); err != nil {
		return fmt.Errorf("loading components: %w", err)
	}

	// Process template
	result, err := processor.Assembler(tree)
	if err != nil {
		return fmt.Errorf("processing template: %w", err)
	}

	// Write output files
	if err := b.writeOutput(outputRel, result); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	// processor.Cleanup()
	// registry.Cleanup()
	return nil
}

// writeOutput writes all generated files to disk
func (b *Builder) writeOutput(outputRel string, result *template.ProcessResult) error {
	files := make(map[string][]byte)

	// Strip the blueprints/ prefix if present and get base output path
	outputPath := strings.TrimPrefix(outputRel, "blueprints/")

	// Add main HTML file
	files[outputPath+".html"] = result.HTML

	// Add asset files to appropriate directories
	for name, content := range result.Files {
		var dir string
		switch filepath.Ext(name) {
		case ".css":
			dir = "css"
		case ".js":
			dir = "js"
		default:
			dir = "assets"
		}
		files[filepath.Join(dir, name)] = content
	}

	// Write all files
	targetPath := b.store.GetTargetPath()
	return b.store.WriteOutput(targetPath, files)
}