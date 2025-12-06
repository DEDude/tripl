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
	var statementBuilder strings.Builder
	
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		
		if strings.HasPrefix(trimmed, "@prefix") {
			prefix, uri, err := parsePrefix(trimmed)
			if err != nil {
				return nil, nil, err
			}
			prefixes[prefix] = uri
			continue
		}
		
		statementBuilder.WriteString(" ")
		statementBuilder.WriteString(line)
		
		if strings.HasSuffix(trimmed, ".") {
			statement := statementBuilder.String()
			statementBuilder.Reset()
			
			ts, err := parseTurtleStatement(statement, prefixes)
			if err != nil {
				return nil, nil, err
			}
			triples = append(triples, ts...)
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

func parseTurtleStatement(statement string, prefixes map[string]string) ([]triple.Triple, error) {
	statement = strings.TrimSpace(statement)
	statement = strings.TrimSuffix(statement, ".")
	statement = strings.TrimSpace(statement)
	
	var triples []triple.Triple
	
	subject, rest, err := parseTurtleNode(statement, prefixes)
	if err != nil {
		return nil, err
	}
	
	predicateObjectPairs := splitBySemicolon(rest)
	
	for _, pair := range predicateObjectPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}
		
		predicate, objectsStr, err := parseTurtleNode(pair, prefixes)
		if err != nil {
			return nil, err
		}
		
		if pred, ok := predicate.(triple.IRI); ok && pred.Value == "a" {
			predicate = triple.IRI{Value: rdfType}
		}
		
		objects := splitByComma(objectsStr)
		
		for _, objStr := range objects {
			objStr = strings.TrimSpace(objStr)
			if objStr == "" {
				continue
			}
			
			object, _, err := parseTurtleNode(objStr, prefixes)
			if err != nil {
				return nil, err
			}
			
			triples = append(triples, triple.Triple{
				Subject:   subject,
				Predicate: predicate,
				Object:    object,
			})
		}
	}
	
	return triples, nil
}

func splitBySemicolon(s string) []string {
	return splitByDelimiter(s, ';')
}

func splitByComma(s string) []string {
	return splitByDelimiter(s, ',')
}

func splitByDelimiter(s string, delim rune) []string {
	var parts []string
	var current strings.Builder
	inQuotes := false
	inAngleBrackets := false
	
	for i, ch := range s {
		switch ch {
		case '"':
			if i == 0 || s[i-1] != '\\' {
				inQuotes = !inQuotes
			}
			current.WriteRune(ch)
		case '<':
			if !inQuotes {
				inAngleBrackets = true
			}
			current.WriteRune(ch)
		case '>':
			if !inQuotes {
				inAngleBrackets = false
			}
			current.WriteRune(ch)
		case delim:
			if !inQuotes && !inAngleBrackets {
				parts = append(parts, current.String())
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}
	
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	
	return parts
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
	
	if prefixedName == "a" {
		return triple.IRI{Value: "a"}, rest, nil
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


