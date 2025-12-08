package encoder

import (
	"fmt"
	"github.com/DeDude/tripl/pkg/triple"
)

type parseContext struct {
	line   int
	column int
	input  string
}

func (pc *parseContext) error(msg string) error {
	return fmt.Errorf("%s at line %d, column %d", msg, pc.line, pc.column)
}

func (pc *parseContext) advance(n int) {
	pc.column += n
}

func (pc *parseContext) newLine() {
	pc.line++
	pc.column = 1
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

func nodeKey(n triple.Node) string {
	switch node := n.(type) {
	case triple.IRI:
		return "iri:" + node.Value
	case triple.Literal:
		return fmt.Sprintf("lit:%s:%s:%s", node.Value, node.Language, node.Datatype)
	case triple.BlankNode:
		return "blank:" + node.Value
	default:
		return ""
	}
}
