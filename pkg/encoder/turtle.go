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

func EncodeTurtleCompact(triples []triple.Triple, prefixes map[string]string) string {
	var result strings.Builder

	for prefix, uri := range prefixes {
		result.WriteString(fmt.Sprintf("@prefix %s: <%s> .\n", prefix, uri))
	}

	if len(prefixes) > 0{
		result.WriteString("\n")
	}

	grouped := groupTriples(triples)

	for i, subjectGroup := range grouped {
		if i > 0 {
			result.WriteString("\n")
		}

		subject := formatTurtleNode(subjectGroup.subject, prefixes)
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
				result.WriteString(formatTurtleNode(predicate, prefixes))
			}

			for k, obj := range predGroup.objects {
				if k == 0 {
					result.WriteString(" ")
				} else {
					result.WriteString(", ")
				}

				result.WriteString(formatTurtleNode(obj, prefixes))
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
	subjectMap := make(map[string]*subjectGroup)
	var orderedSubjects []string
	
	for _, t := range triples {
		subjectKey := nodeKey(t.Subject)
		
		if _, exists := subjectMap[subjectKey]; !exists {
			subjectMap[subjectKey] = &subjectGroup{
				subject:    t.Subject,
				predicates: []predicateGroup{},
			}
			orderedSubjects = append(orderedSubjects, subjectKey)
		}
		
		sg := subjectMap[subjectKey]
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
				objects:   []triple.Node{},
			})
			pg = &sg.predicates[len(sg.predicates)-1]
		}
		
		pg.objects = append(pg.objects, t.Object)
	}
	
	var result []subjectGroup
	for _, key := range orderedSubjects {
		result = append(result, *subjectMap[key])
	}
	
	return result
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

