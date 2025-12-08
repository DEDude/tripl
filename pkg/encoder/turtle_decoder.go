package encoder

import (
	"bufio"
	"fmt"
	"github.com/DeDude/tripl/pkg/triple"
	"strings"
)

func DecodeTurtle(input string) ([]triple.Triple, map[string]string, error) {
	var triples []triple.Triple
	prefixes := make(map[string]string)
	resolver := NewPrefixResolver(prefixes)

	scanner := bufio.NewScanner(strings.NewReader(input))
	var statementBuilder strings.Builder
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		if strings.HasPrefix(trimmed, "@prefix") {
			prefix, uri, err := parsePrefixWithLine(trimmed, lineNum)
			if err != nil {
				return nil, nil, err
			}
			resolver.Set(prefix, uri)
			continue
		}

		statementBuilder.WriteString(" ")
		statementBuilder.WriteString(line)

		if strings.HasSuffix(trimmed, ".") {
			statement := statementBuilder.String()
			statementBuilder.Reset()

			ts, err := parseTurtleStatementWithLine(statement, resolver, lineNum)
			if err != nil {
				return nil, nil, err
			}
			triples = append(triples, ts...)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return triples, resolver.All(), nil
}

func parsePrefixWithLine(line string, lineNum int) (string, string, error) {
	ctx := &parseContext{line: lineNum, column: 1}
	line = strings.TrimPrefix(line, "@prefix")
	line = strings.TrimSpace(line)
	line = strings.TrimSuffix(line, ".")
	line = strings.TrimSpace(line)

	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return "", "", ctx.error("invalid prefix declaration")
	}

	prefix := strings.TrimSpace(parts[0])
	uriPart := strings.TrimSpace(parts[1])

	if !strings.HasPrefix(uriPart, "<") || !strings.HasSuffix(uriPart, ">") {
		return "", "", ctx.error("prefix URI must be in angle brackets")
	}

	uri := uriPart[1 : len(uriPart)-1]

	return prefix, uri, nil
}

func parseTurtleStatementWithLine(statement string, resolver *PrefixResolver, lineNum int) ([]triple.Triple, error) {
	ctx := &parseContext{line: lineNum, column: 1}
	statement = strings.TrimSpace(statement)
	statement = strings.TrimSuffix(statement, ".")
	statement = strings.TrimSpace(statement)

	var triples []triple.Triple

	subject, rest, err := parseTurtleNode(statement, resolver)
	if err != nil {
		return nil, ctx.error(fmt.Sprintf("parsing subject: %v", err))
	}

	predicateObjectPairs := splitBySemicolon(rest)

	for _, pair := range predicateObjectPairs {
		pair = strings.TrimSpace(pair)
		if pair == "" {
			continue
		}

		predicate, objectsStr, err := parseTurtleNode(pair, resolver)
		if err != nil {
			return nil, ctx.error(fmt.Sprintf("parsing predicate: %v", err))
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

			object, _, err := parseTurtleNode(objStr, resolver)
			if err != nil {
				return nil, ctx.error(fmt.Sprintf("parsing object: %v", err))
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

func parseTurtleNode(s string, resolver *PrefixResolver) (triple.Node, string, error) {
	s = strings.TrimSpace(s)

	if strings.HasPrefix(s, "<") {
		return parseIRI(s)
	}

	if strings.HasPrefix(s, "_:") {
		return parseBlankNode(s)
	}

	if strings.HasPrefix(s, `"`) {
		lit, rest, err := parseLiteral(s)
		if err != nil {
			return nil, "", err
		}

		if lit.Datatype == "" && strings.HasPrefix(rest, "^^") {
			rest = rest[2:]
			if !strings.HasPrefix(rest, "<") {
				parts := strings.SplitN(rest, " ", 2)
				datatypePrefix := parts[0]
				lit.Datatype = resolver.Expand(datatypePrefix)
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

	expanded := resolver.Expand(prefixedName)
	return triple.IRI{Value: expanded}, rest, nil
}
