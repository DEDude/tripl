package encoder

import (
	"github.com/DeDude/tripl/pkg/triple"
	"strings"
)

func DecodeNTriple(line string) (triple.Triple, error) {
	ctx := &parseContext{line: 1, column: 1, input: line}
	line = strings.TrimSpace(line)

	if line == "" || strings.HasPrefix(line, "#") {
		return triple.Triple{}, ctx.error("empty or comment line")
	}

	if !strings.HasSuffix(line, ".") {
		return triple.Triple{}, ctx.error("line must end with .")
	}

	line = strings.TrimSuffix(line, ".")
	line = strings.TrimSpace(line)

	subject, rest, err := parseNodeWithContext(line, ctx)
	if err != nil {
		return triple.Triple{}, err
	}

	predicate, rest, err := parseNodeWithContext(rest, ctx)
	if err != nil {
		return triple.Triple{}, err
	}

	object, _, err := parseNodeWithContext(rest, ctx)
	if err != nil {
		return triple.Triple{}, err
	}

	return triple.Triple{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
	}, nil
}

func parseNodeWithContext(s string, ctx *parseContext) (triple.Node, string, error) {
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "<") {
		return parseIRIWithContext(s, ctx)
	}

	if strings.HasPrefix(s, "_:") {
		return parseBlankNodeWithContext(s, ctx)
	}

	if strings.HasPrefix(s, `"`) {
		return parseLiteralWithContext(s, ctx)
	}

	return nil, "", ctx.error("invalid node format")
}
