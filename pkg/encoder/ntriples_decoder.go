package encoder

import (
	"errors"
	"github.com/DeDude/tripl/pkg/triple"
	"strings"
)

func DecodeNTriple(line string) (triple.Triple, error) {
	line = strings.TrimSpace(line)

	if line == "" || strings.HasPrefix(line, "#") {
		return triple.Triple{}, errors.New("empty or comment line")
	}

	if !strings.HasSuffix(line, ".") {
		return triple.Triple{}, errors.New("line must end with .")
	}

	line = strings.TrimSuffix(line, ".")
	line = strings.TrimSpace(line)

	subject, rest, err := parseNode(line)
	if err != nil {
		return triple.Triple{}, err
	}

	predicate, rest, err := parseNode(rest)
	if err != nil {
		return triple.Triple{}, err
	}

	object, _, err := parseNode(rest)
	if err != nil {
		return triple.Triple{}, err
	}

	return triple.Triple{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
	}, nil
}

func parseNode(s string) (triple.Node, string, error) {
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "<") {
		return parseIRI(s)
	}

	if strings.HasPrefix(s, "_:") {
		return parseBlankNode(s)
	}

	if strings.HasPrefix(s, `"`) {
		return parseLiteral(s)
	}

	return nil, "", errors.New("invalid node format")
}
