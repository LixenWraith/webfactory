package assets

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"webfactory/src/internal/component"
)

type Manager struct {
	css     map[string][]byte  // content hash -> content
	cssKeys []string           // ordered list of css content hashes
	js      map[string]jsAsset // content hash -> {content, files}
	jsKeys  []string           // ordered list of js content hashes
}

type jsAsset struct {
	content []byte
	files   []string // list of "component-filename.js"
}

func New() *Manager {
	return &Manager{
		css:     make(map[string][]byte),
		cssKeys: make([]string, 0),
		js:      make(map[string]jsAsset),
		jsKeys:  make([]string, 0),
	}
}

// ProcessComponent handles all assets for a component
func (m *Manager) ProcessComponent(comp *component.Component) error {
	if comp == nil {
		return nil
	}

	// Handle CSS - hash based deduplication with order preservation
	if len(comp.Styles) > 0 {
		hash := generateHash(comp.Styles)
		if _, exists := m.css[hash]; !exists {
			m.css[hash] = comp.Styles
			m.cssKeys = append(m.cssKeys, hash)
		}
	}

	// Handle JS - content based deduplication with filename tracking and order preservation
	for origName, content := range comp.Scripts {
		hash := generateHash(content)
		baseName := strings.TrimSuffix(origName, ".js")
		outName := fmt.Sprintf("%s-%s", sanitizeFileName(comp.Path), baseName)

		if asset, exists := m.js[hash]; exists {
			// Add new filename to existing content
			asset.files = append(asset.files, outName)
			m.js[hash] = asset
		} else {
			// Store new content with filename
			m.js[hash] = jsAsset{
				content: content,
				files:   []string{outName},
			}
			m.jsKeys = append(m.jsKeys, hash)
		}
	}

	return nil
}

// GetAssetTags returns both style and script tags
func (m *Manager) GetAssetTags(prefix string) (styles, scripts string) {
	// All CSS is merged into one file
	if len(m.css) > 0 {
		styles = fmt.Sprintf(`<link rel="stylesheet" href="%s">`,
			filepath.Join(prefix, "css", "styles.css"))
	}

	// Generate script tags for each unique JS file
	var jsB bytes.Buffer
	for _, asset := range m.js {
		for _, filename := range asset.files {
			jsName := sanitizeFileName(filename) + ".js"
			jsB.WriteString(fmt.Sprintf(`<script src="%s"></script>`,
				filepath.Join(prefix, "js", jsName)))
			jsB.WriteByte('\n')
		}
	}
	scripts = strings.TrimSpace(jsB.String())

	return styles, scripts
}

// GetFiles returns all CSS and JS files for output
func (m *Manager) GetFiles() map[string][]byte {
	files := make(map[string][]byte)

	// Merge all CSS in order
	if len(m.css) > 0 {
		var merged bytes.Buffer
		for _, hash := range m.cssKeys {
			if content, exists := m.css[hash]; exists {
				merged.Write(content)
				merged.WriteByte('\n')
			}
		}
		files["styles.css"] = bytes.TrimSuffix(merged.Bytes(), []byte{'\n'})
	}

	// Keep JS files separate but ordered
	for _, hash := range m.jsKeys {
		if asset, exists := m.js[hash]; exists {
			for _, filename := range asset.files {
				jsName := sanitizeFileName(filename) + ".js"
				files[jsName] = asset.content
			}
		}
	}

	return files
}

// generateHash creates a hash of content for deduplication
func generateHash(content []byte) string {
	h := sha256.New()
	h.Write(content)
	return hex.EncodeToString(h.Sum(nil))
}

// sanitizeFileName creates a safe filename from component path
func sanitizeFileName(path string) string {
	// Replace dots and any non-alphanumeric with dash
	name := strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' ||
			r >= 'A' && r <= 'Z' ||
			r >= '0' && r <= '9' {
			return r
		}
		return '-'
	}, path)

	// Remove consecutive dashes
	for strings.Contains(name, "--") {
		name = strings.ReplaceAll(name, "--", "-")
	}

	// Trim dashes from ends
	return strings.Trim(name, "-")
}