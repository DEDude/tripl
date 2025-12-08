package encoder

import (
	"fmt"
	"github.com/DeDude/tripl/pkg/triple"
	"strings"
)

func EncodeTurtle(triples []triple.Triple, prefixes map[string]string) string {
	resolver := NewPrefixResolver(prefixes)
	var result strings.Builder

	for prefix, uri := range prefixes {
		result.WriteString(fmt.Sprintf("@prefix %s: <%s> .\n", prefix, uri))
	}

	if len(prefixes) > 0 {
		result.WriteString("\n")
	}

	for _, t := range triples {
		subject := formatTurtleNode(t.Subject, resolver)
		predicate := formatTurtleNode(t.Predicate, resolver)
		object := formatTurtleNode(t.Object, resolver)
		result.WriteString(fmt.Sprintf("%s %s %s .\n", subject, predicate, object))
	}

	return result.String()
}

func formatTurtleNode(n triple.Node, resolver *PrefixResolver) string {
	switch node := n.(type) {
	case triple.IRI:
		shortened := resolver.Shorten(node.Value)
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
			datatypeShort := resolver.Shorten(node.Datatype)
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

func EncodeTurtleCompact(triples []triple.Triple, prefixes map[string]string) string {
	resolver := NewPrefixResolver(prefixes)
	var result strings.Builder

	for prefix, uri := range prefixes {
		result.WriteString(fmt.Sprintf("@prefix %s: <%s> .\n", prefix, uri))
	}

	if len(prefixes) > 0 {
		result.WriteString("\n")
	}

	grouped := groupTriples(triples)

	for i, subjectGroup := range grouped {
		if i > 0 {
			result.WriteString("\n")
		}

		subject := formatTurtleNode(subjectGroup.subject, resolver)
		result.WriteString(subject)

		for j, predGroup := range subjectGroup.predicates {
			if j == 0 {
				result.WriteString(" ")
			} else {
				result.WriteString(" ;\n         ")
			}

			predicate := predGroup.predicate

			if iri, ok := predicate.(triple.IRI); ok && iri.Value == rdfType {
				result.WriteString("a")
			} else {
				result.WriteString(formatTurtleNode(predicate, resolver))
			}

			for k, obj := range predGroup.objects {
				if k == 0 {
					result.WriteString(" ")
				} else {
					result.WriteString(", ")
				}

				result.WriteString(formatTurtleNode(obj, resolver))
			}
		}

		result.WriteString(" .\n")
	}

	return result.String()
}

const rdfType = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"

type subjectGroup struct {
	subject    triple.Node
	predicates []predicateGroup
}

type predicateGroup struct {
	predicate triple.Node
	objects   []triple.Node
}

func groupTriples(triples []triple.Triple) []subjectGroup {
	subjectMap := make(map[string]*subjectGroup, len(triples)/2)
	orderedSubjects := make([]string, 0, len(triples)/2)

	for _, t := range triples {
		subjectKey := nodeKey(t.Subject)

		sg, exists := subjectMap[subjectKey]
		if !exists {
			sg = &subjectGroup{
				subject:    t.Subject,
				predicates: make([]predicateGroup, 0, 2),
			}
			subjectMap[subjectKey] = sg
			orderedSubjects = append(orderedSubjects, subjectKey)
		}

		predicateKey := nodeKey(t.Predicate)

		var pg *predicateGroup
		for i := range sg.predicates {
			if nodeKey(sg.predicates[i].predicate) == predicateKey {
				pg = &sg.predicates[i]
				break
			}
		}

		if pg == nil {
			sg.predicates = append(sg.predicates, predicateGroup{
				predicate: t.Predicate,
				objects:   make([]triple.Node, 0, 1),
			})
			pg = &sg.predicates[len(sg.predicates)-1]
		}

		pg.objects = append(pg.objects, t.Object)
	}

	result := make([]subjectGroup, len(orderedSubjects))
	for i, key := range orderedSubjects {
		result[i] = *subjectMap[key]
	}

	return result
}
