package encoder

import (
	"fmt"
	"strings"
	"github.com/DeDude/tripl/pkg/triple"
)

func EncodeTurtle(triples []triple.Triple, prefixes map[string]string) string {
	var result strings.Builder

	for prefix, uri := range prefixes {
		result.WriteString(fmt.Sprintf("@prefix %s: <%s> .\n", prefix, uri))
	}

	if len(prefixes) > 0 {
		result.WriteString("\n")
	}

	for _, t := range triples {
		subject := formatTurtleNode(t.Subject, prefixes)
		predicate := formatTurtleNode(t.Predicate, prefixes)
		object := formatTurtleNode(t.Object, prefixes)
		result.WriteString(fmt.Sprintf("%s %s %s .\n", subject, predicate, object))
	}

	return result.String()
}

func formatTurtleNode(n triple.Node, prefixes map[string]string) string {
	switch node := n.(type) {
	case triple.IRI:
		shortened := shortenIRI(node.Value, prefixes)
		if shortened != node.Value {
			return shortened
		}
		return fmt.Sprintf("<%s>", node.Value)
	case triple.Literal:
		result := fmt.Sprintf(`"%s"`, node.Value)
		if node.Language != "" {
			result += "@" + node.Language
		}
		if node.Datatype != "" {
			datatypeShort := shortenIRI(node.Datatype, prefixes)
			if datatypeShort != node.Datatype {
				result += "^^" + datatypeShort
			} else {
				result += "^^<" + node.Datatype + ">"
			}
		}
		return result
	case triple.BlankNode:
		return "_:" + node.Value
	default:
		return ""
	}
}

func shortenIRI(iri string, prefixes map[string]string) string {
	for prefix, uri := range prefixes {
		if strings.HasPrefix(iri, uri) {
			localPart := strings.TrimPrefix(iri, uri)
			return prefix + ":" + localPart
		}
	}
	return iri
}
