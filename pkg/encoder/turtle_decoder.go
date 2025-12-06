package encoder

import (
	"bufio"
	"errors"
	"strings"
	"github.com/DeDude/tripl/pkg/triple"
)

func DecodeTurtle(input string) ([]triple.Triple, map[string]string, error) {
	var triples []triple.Triple
	prefixes := make(map[string]string)

	scanner := bufio.NewScanner(strings.NewReader(input))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "@prefix") {
			prefix, uri, err := parsePrefix(line)
			if err != nil {
				return nil, nil, err
			}
			prefixes[prefix] = uri
			continue
		}

		if strings.HasSuffix(line, ".") {
			t, err := parseTurtleTriple(line, prefixes)
			if err != nil {
				return nil, nil, err
			}
			triples = append(triples, t)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return triples, prefixes, nil
}

func parsePrefix(line string) (string, string, error) {
	line = strings.TrimPrefix(line, "@prefix")
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, ".")
	line = strings.TrimSpace(line)

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", errors.New("invalid prefix declaration")
	}

	prefix  := strings.TrimSpace(parts[0])
	uriPart := strings.TrimSpace(parts[1])

	if !strings.HasPrefix(uriPart, "<") || !strings.HasSuffix(uriPart, ">") {
		return "", "", errors.New("prefix URI must be in angle brackets") 
	}

	uri := uriPart[1: len(uriPart)-1]

	return prefix, uri, nil
}

func parseTurtleTriple(line string, prefixes map[string]string) (triple.Triple, error) {
	line = strings.TrimSuffix(line, ".")
	line = strings.TrimSpace(line)

	subject, rest, err := parseTurtleNode(line, prefixes)
	if err != nil {
		return triple.Triple{}, err
	}

	predicate, rest, err := parseTurtleNode(rest, prefixes)
	if err != nil {
		return triple.Triple{}, err
	}

	object, _, err := parseTurtleNode(rest, prefixes)
	if err != nil {
		return triple.Triple{}, err
	}

	return triple.Triple{
		Subject:   subject,
		Predicate: predicate,
		Object:    object,
	}, nil
}

func parseTurtleNode(s string, prefixes map[string]string) (triple.Node, string, error) {
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
		rest  := strings.TrimSpace(s[end+1:])
		lit   := triple.Literal{Value: value}

		if strings.HasPrefix(rest, "@") {
			parts := strings.SplitN(rest[1:], " ", 2)
			lit.Language = parts[0]
			if len(parts) > 1 {
				rest = parts[1]
			} else {
				rest = ""
			}
		} else if strings.HasPrefix(rest, "^^") {
			rest = rest[2:]
			if strings.HasPrefix(rest, "<") {
				end := strings.Index(rest, ">")
				if end == -1 {
					return nil, "", errors.New("unclosed datatype")
				}
				lit.Datatype = rest[1:end]
				rest = strings.TrimSpace(rest[end+1:])
			} else {
				parts := strings.SplitN(rest, " ", 2)
				datatypePrefix := parts[0]
				lit.Datatype = expandPrefix(datatypePrefix, prefixes)
				if len(parts) > 1 {
					rest = parts[1]
				} else {
					rest = ""
				}
			}
		}

		return lit, rest, nil
	}

	parts := strings.SplitN(s, " ", 2)
	prefixedName := parts[0]
	rest := ""
	if len(parts) > 1 {
		rest = parts[1]
	}

	expanded := expandPrefix(prefixedName, prefixes)
	return triple.IRI{Value: expanded}, rest, nil
}

func expandPrefix(prefixedName string, prefixes map[string]string) string {
	parts := strings.SplitN(prefixedName, ":", 2)
	if len(parts) != 2 {
		return prefixedName
	}

	prefix := parts[0]
	localPart := parts[1]

	if uri, ok := prefixes[prefix]; ok {
		return uri + localPart
	}

	return prefixedName
}


