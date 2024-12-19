package template

import (
	"bytes"
	"strings"
)

type TokenType int

const (
	TextToken TokenType = iota
	RangeStartToken
	RangeEndToken
	VarToken
	ComponentToken
	StyleToken
	ScriptToken
)

type Token struct {
	Type    TokenType
	Content string // Variable name for Var/Range, raw content for Text
}

type Tokenizer struct {
	template []byte
	pos      int
	tokens   []Token
}

func NewTokenizer(template []byte) *Tokenizer {
	return &Tokenizer{
		template: template,
		tokens:   make([]Token, 0),
	}
}

func (t *Tokenizer) Tokenize() []Token {
	for t.pos < len(t.template) {
		if t.template[t.pos] == '{' && t.pos+1 < len(t.template) && t.template[t.pos+1] == '{' {
			// Handle accumulated text before directive
			if t.pos > 0 && len(t.template) > 0 {
				t.tokens = append(t.tokens, Token{
					Type:    TextToken,
					Content: string(t.template[0:t.pos]),
				})
				t.template = t.template[t.pos:]
				t.pos = 0
			}

			// Find directive end
			end := bytes.Index(t.template[2:], []byte("}}"))
			if end == -1 {
				// Malformed template - treat rest as text
				t.tokens = append(t.tokens, Token{
					Type:    TextToken,
					Content: string(t.template),
				})
				break
			}

			directive := strings.TrimSpace(string(t.template[2 : end+2]))

			switch {
			case directive == "component":
				t.tokens = append(t.tokens, Token{
					Type: ComponentToken,
				})
			case directive == "range end":
				t.tokens = append(t.tokens, Token{
					Type: RangeEndToken,
				})
			case strings.HasPrefix(directive, "range ."):
				t.tokens = append(t.tokens, Token{
					Type:    RangeStartToken,
					Content: strings.TrimPrefix(directive, "range ."),
				})
			case directive == "styles":
				t.tokens = append(t.tokens, Token{
					Type: StyleToken,
				})
			case directive == "script":
				t.tokens = append(t.tokens, Token{
					Type: ScriptToken,
				})
			case strings.HasPrefix(directive, "."):
				t.tokens = append(t.tokens, Token{
					Type:    VarToken,
					Content: strings.TrimPrefix(directive, "."),
				})
			}

			t.template = t.template[end+4:]
			t.pos = 0
			continue
		}
		t.pos++
	}

	// Handle remaining text
	if len(t.template) > 0 {
		t.tokens = append(t.tokens, Token{
			Type:    TextToken,
			Content: string(t.template),
		})
	}

	return t.tokens
}