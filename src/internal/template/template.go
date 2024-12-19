package template

import (
	"bytes"
	"fmt"
	"strings"
	"webfactory/src/internal/assets"
	"webfactory/src/internal/blueprint"
	"webfactory/src/internal/component"
)

// ProcessResult contains all processed template outputs
type ProcessResult struct {
	HTML       []byte
	Files      map[string][]byte // Combined CSS and individual JS files from GetFiles()
	Components map[string]string
}

// Assembler wraps Process() to return all template outputs
func (p *Processor) Assembler(node *blueprint.Node) (*ProcessResult, error) {
	html, err := p.Process(node)
	if err != nil {
		return nil, fmt.Errorf("processing template: %w", err)
	}

	stylesTag, scriptTags := p.assets.GetAssetTags("")
	var finalBuf bytes.Buffer

	if p.hasStyles {
		html = bytes.ReplaceAll(html, []byte("{{styles}}"), []byte(stylesTag))
	} else if stylesTag != "" {
		finalBuf.WriteString(stylesTag)
	}

	finalBuf.Write(html)

	if p.hasScripts {
		finalBuf = *bytes.NewBuffer(bytes.ReplaceAll(finalBuf.Bytes(), []byte("{{script}}"), []byte(scriptTags)))
	} else if scriptTags != "" {
		finalBuf.WriteString(scriptTags)
	}

	result := &ProcessResult{
		HTML:       finalBuf.Bytes(),
		Files:      p.assets.GetFiles(),
		Components: p.GetUsedComponents(),
	}

	return result, nil
}

type Processor struct {
	registry   *component.Registry
	assets     *assets.Manager
	vars       map[string][]string
	errLines   []processError
	hasStyles  bool
	hasScripts bool
}

type processError struct {
	line      int
	directive string
	msg       string
}

func New(registry *component.Registry) *Processor {
	return &Processor{
		registry: registry,
		assets:   assets.New(),
		vars:     make(map[string][]string),
		errLines: make([]processError, 0),
	}
}

// Process handles template processing from root node
func (p *Processor) Process(node *blueprint.Node) ([]byte, error) {
	if node == nil {
		return nil, nil
	}

	var output []byte

	// Process root's children as it's a virtual node
	if node.Block.ID == -1 {
		output = p.processChildren(node)
	} else {
		comp := p.registry.Get(node.Block.Path)
		if comp == nil {
			p.addError(0, node.Block.Path, fmt.Sprintf("component not found: %s", node.Block.Path))
			return []byte(fmt.Sprintf("{{%s}}", node.Block.Path)), nil
		}

		// Process html and assets
		p.processAssets(comp, node.Block.Path)
		output = p.processTemplate(comp.Template, node.Block.Vars, node.Children)
	}

	if len(p.errLines) > 0 {
		var msgs []string
		for _, err := range p.errLines {
			msgs = append(msgs, fmt.Sprintf("line %d [%s]: %s", err.line, err.directive, err.msg))
		}
		return output, fmt.Errorf("template processing errors: %s", strings.Join(msgs, "; "))
	}

	return output, nil
}

func (p *Processor) GetUsedComponents() map[string]string {
	paths := make(map[string]string)
	p.registry.Each(func(comp *component.Component) {
		fsPath := strings.ReplaceAll(comp.Path, ".", "/")
		paths[comp.Path] = fsPath
	})
	return paths
}

// func (p *Processor) Cleanup() {
// 	p.registry = nil
// 	p.assets = nil
// 	p.vars = nil
// 	p.errLines = nil
// }

func (p *Processor) addError(line int, directive string, msg string) {
	// Check for duplicate
	for _, err := range p.errLines {
		if err.line == line && err.directive == directive {
			return
		}
	}
	p.errLines = append(p.errLines, processError{
		line:      line,
		directive: directive,
		msg:       msg,
	})
}

func (p *Processor) processAssets(comp *component.Component, path string) {
	if err := p.assets.ProcessComponent(comp); err != nil {
		p.addError(0, path, fmt.Sprintf("asset error in %s: %v", path, err))
	}
}

// processChildren handles child components recursively
func (p *Processor) processChildren(node *blueprint.Node) []byte {
	var buf bytes.Buffer
	for _, child := range node.Children {
		childContent, _ := p.Process(child)
		if len(childContent) > 0 {
			buf.Write(childContent)
		}
	}
	return buf.Bytes()
}

// processTemplate handles template substitution
func (p *Processor) processTemplate(tmpl []byte, vars map[string][]string, children []*blueprint.Node) []byte {
	tokenizer := NewTokenizer(tmpl)
	tokens := tokenizer.Tokenize()

	var buf bytes.Buffer
	var rangeVar string
	var rangeContent []string
	inRange := false
	rangeStart := -1

	for i, token := range tokens {
		switch token.Type {
		case TextToken:
			if !inRange {
				buf.WriteString(token.Content)
			}
		case StyleToken:
			p.hasStyles = true

		case ScriptToken:
			p.hasScripts = true

		case ComponentToken:
			if !inRange {
				for _, child := range children {
					childContent, _ := p.Process(child)
					if len(childContent) > 0 {
						buf.Write(childContent)
					}
				}
			}

		case RangeStartToken:
			if !inRange {
				inRange = true
				rangeVar = token.Content
				rangeStart = i
				// Get array of values for range variable
				if values, ok := vars[rangeVar]; ok {
					rangeContent = values
				}
			}

		case RangeEndToken:
			if inRange {
				// For each value in range
				for _, rangeValue := range rangeContent {
					var rangeBuf bytes.Buffer
					// Process range block tokens
					for _, t := range tokens[rangeStart+1 : i] {
						switch t.Type {
						case TextToken:
							rangeBuf.WriteString(t.Content)
						case VarToken:
							if t.Content == rangeVar {
								// Use current iteration value
								rangeBuf.WriteString(rangeValue)
							} else if values, exists := vars[t.Content]; exists && len(values) > 0 {
								// Other vars use first value
								rangeBuf.WriteString(values[0])
							}
						}
					}
					buf.Write(rangeBuf.Bytes())
				}
				inRange = false
				rangeStart = -1
			}

		case VarToken:
			if !inRange {
				if values, exists := vars[token.Content]; exists {
					buf.WriteString(values[0])
				}
			}
		}
	}

	return buf.Bytes()
}