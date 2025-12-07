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
		end := strings.Index(s, ">")
		if end == -1 {
			return nil, "", errors.New("unclosed IRI")
		}
		value := s[1:end]
		rest := strings.TrimSpace(s[end+1:])
		return triple.IRI{Value: value}, rest, nil
	}

	if strings.HasPrefix(s, "_:") {
		parts := strings.SplitN(s[2:], " ", 2)
		value := parts[0]
		rest := ""

		if len(parts) > 1 {
			rest = parts[1]
		}
		return triple.BlankNode{Value: value}, rest, nil
	}

	if strings.HasPrefix(s, `"`) {
		end := 1
		for end < len(s) {
			if s[end] == '"' && s[end-1] != '\\' {
				break
			}
			end++
		}
		if end >= len(s) {
			return nil, "", errors.New("unclosed literal")
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
		} else if strings.HasPrefix(rest, "^^<") {
			end := strings.Index(rest, ">")
			if end == -1 {
				return nil, "", errors.New("unclosed datatype")
			}
			lit.Datatype = rest[3:end]
			rest = strings.TrimSpace(rest[end+1:])
		}

		return lit, rest, nil
	}

	return nil, "", errors.New("invalid node format")
}
