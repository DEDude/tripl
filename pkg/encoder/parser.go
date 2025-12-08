package encoder

import (
	"strings"

	"github.com/DeDude/tripl/pkg/triple"
)

func parseIRI(s string) (triple.IRI, string, error) {
	if !strings.HasPrefix(s, "<") {
		return triple.IRI{}, s, &parseContext{}.error("expected IRI to start with <")
	}

	end := strings.Index(s, ">")
	if end == -1 {
		return triple.IRI{}, "", &parseContext{}.error("unclosed IRI")
	}

	value := s[1:end]
	rest := strings.TrimSpace(s[end+1:])
	return triple.IRI{Value: value}, rest, nil
}

func parseBlankNode(s string) (triple.BlankNode, string, error) {
	if !strings.HasPrefix(s, "_:") {
		return triple.BlankNode{}, s, &parseContext{}.error("expected blank node to start with _:")
	}

	parts := strings.SplitN(s[2:], " ", 2)
	value := parts[0]
	rest := ""
	if len(parts) > 1 {
		rest = parts[1]
	}

	return triple.BlankNode{Value: value}, rest, nil
}

func parseLiteral(s string) (triple.Literal, string, error) {
	if !strings.HasPrefix(s, `"`) {
		return triple.Literal{}, s, &parseContext{}.error("expected literal to start with \"")
	}

	end := 1
	for end < len(s) {
		if s[end] == '"' && (end == 1 || s[end-1] != '\\') {
			break
		}
		end++
	}

	if end >= len(s) {
		return triple.Literal{}, "", &parseContext{}.error("unclosed literal")
	}

	value := s[1:end]
	rest := strings.TrimSpace(s[end+1:])
	lit := triple.Literal{Value: value}

	if strings.HasPrefix(rest, "@") {
		parts := strings.SplitN(rest[1:], " ", 2)
		lit.Language = parts[0]

		if len(parts) > 1 {
			rest = parts[1]
		} else {
			rest = ""
		}
		return lit, rest, nil
	}

	if strings.HasPrefix(rest, "^^<") {
		end := strings.Index(rest, ">")
		if end == -1 {
			return triple.Literal{}, "", &parseContext{}.error("unclosed datatype")
		}
		lit.Datatype = rest[3:end]
		rest = strings.TrimSpace(rest[end+1:])
	}

	return lit, rest, nil
}

func parseIRIWithContext(s string, ctx *parseContext) (triple.IRI, string, error) {
	if !strings.HasPrefix(s, "<") {
		return triple.IRI{}, s, ctx.error("expected IRI to start with <")
	}

	end := strings.Index(s, ">")
	if end == -1 {
		return triple.IRI{}, "", ctx.error("unclosed IRI")
	}

	value := s[1:end]
	rest := strings.TrimSpace(s[end+1:])
	ctx.advance(end + 1)
	return triple.IRI{Value: value}, rest, nil
}

func parseBlankNodeWithContext(s string, ctx *parseContext) (triple.BlankNode, string, error) {
	if !strings.HasPrefix(s, "_:") {
		return triple.BlankNode{}, s, ctx.error("expected blank node to start with _:")
	}

	parts := strings.SplitN(s[2:], " ", 2)
	value := parts[0]
	rest := ""
	if len(parts) > 1 {
		rest = parts[1]
	}

	ctx.advance(2 + len(value))
	return triple.BlankNode{Value: value}, rest, nil
}

func parseLiteralWithContext(s string, ctx *parseContext) (triple.Literal, string, error) {
	if !strings.HasPrefix(s, `"`) {
		return triple.Literal{}, s, ctx.error("expected literal to start with \"")
	}

	end := 1
	for end < len(s) {
		if s[end] == '"' && (end == 1 || s[end-1] != '\\') {
			break
		}
		end++
	}

	if end >= len(s) {
		return triple.Literal{}, "", ctx.error("unclosed literal")
	}

	value := s[1:end]
	rest := strings.TrimSpace(s[end+1:])
	lit := triple.Literal{Value: value}
	ctx.advance(end + 1)

	if strings.HasPrefix(rest, "@") {
		parts := strings.SplitN(rest[1:], " ", 2)
		lit.Language = parts[0]
		ctx.advance(1 + len(parts[0]))

		if len(parts) > 1 {
			rest = parts[1]
		} else {
			rest = ""
		}
		return lit, rest, nil
	}

	if strings.HasPrefix(rest, "^^<") {
		end := strings.Index(rest, ">")
		if end == -1 {
			return triple.Literal{}, "", ctx.error("unclosed datatype")
		}
		lit.Datatype = rest[3:end]
		rest = strings.TrimSpace(rest[end+1:])
		ctx.advance(end + 1)
	}

	return lit, rest, nil
}
