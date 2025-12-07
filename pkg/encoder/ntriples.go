package encoder

import (
	"fmt"
	"github.com/DeDude/tripl/pkg/triple"
)

func EncodeNTriple(t triple.Triple) string {
	subject := formatNode(t.Subject)
	predicate := formatNode(t.Predicate)
	object := formatNode(t.Object)
	return fmt.Sprintf("%s %s %s .", subject, predicate, object)
}

func formatNode(n triple.Node) string {
	switch node := n.(type) {
	case triple.IRI:
		return fmt.Sprintf("<%s>", node.Value)
	case triple.Literal:
		result := fmt.Sprintf(`"%s"`, node.Value)
		if node.Language != "" {
			result += "@" + node.Language
		}
		if node.Datatype != "" {
			result += "^^<" + node.Datatype + ">"
		}
		return result
	case triple.BlankNode:
		return "_:" + node.Value
	default:
		return ""
	}
}
