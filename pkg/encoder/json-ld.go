package encoder

import (
		"strings"
		"encoding/json"
		"github.com/DeDude/tripl/pkg/triple"
)

func EncodeJSONLD(triples []triple.Triple) (string, error) {
	grouped := groupBySubject(triples)

	var result []map[string]interface{}

	for _, group := range grouped {
		obj := make(map[string]interface{})

		subjectIRI, ok := group.subject.(triple.IRI)

		if !ok {continue}

		obj["@id"] = subjectIRI.Value

		for predicate, objects := range group.properties {
			var values []map[string]interface{}

			for _, objNode := range objects {
				value := nodeToJSONLD(objNode)

				values = append(values, value)
			}

			obj[predicate] = values
		}

		result = append(result, obj)
	}

	bytes, err := json.MarshalIndent(result, "", " ")
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

type subjectProperties struct {
	subject triple.Node
	properties map[string][]triple.Node
}

func groupBySubject(triples []triple.Triple) []subjectProperties {
	subjectMap := make(map[string]*subjectProperties)

	var orderedKeys []string

	for _, t := range triples {
		subjectKey := nodeKey(t.Subject)

		if _, exists := subjectMap[subjectKey]; !exists {
			subjectMap[subjectKey] = &subjectProperties{
				subject: t.Subject,
				properties: make(map[string][]triple.Node),
			}
			orderedKeys = append(orderedKeys, subjectKey)
		}

		sp := subjectMap[subjectKey]

		predicateIRI, ok := t.Predicate.(triple.IRI)

		if !ok {continue}

		sp.properties[predicateIRI.Value] = append(sp.properties[predicateIRI.Value], t.Object)
	}

	var result []subjectProperties

	for _, key := range orderedKeys {
		result = append(result, *subjectMap[key])
	}

	return result
}

func nodeToJSONLD(n triple.Node) map[string]interface{} {
	result := make(map[string]interface{})

	switch node := n.(type) {
		case triple.IRI:
			result["@id"] = node.Value
		case triple.Literal:
			result["@value"] = node.Value
			if node.Language != "" {
				result["@language"] = node.Language
			}
			if node.Datatype != "" {
				result["@type"] = node.Datatype
			}
		case triple.BlankNode:
			result["@id"] = "_:" + node.Value
	}

	return result
}

func EncodeJSONLDCompact(triples []triple.Triple, context map[string]string) (string, error) {
	grouped := groupBySubject(triples)

	var graph []map[string]interface{}

	for _, group := range grouped {
		obj := make(map[string]interface{})

		subjectIRI, ok := group.subject.(triple.IRI)

		if !ok {continue}

		obj["@id"] = shortenURI(subjectIRI.Value, context)

		for predicate, objects := range group.properties {
			shortPred := shortenURI(predicate, context)

			if len(objects) == 1 {
				if lit, ok := objects[0].(triple.Literal); ok && lit.Language == "" && lit.Datatype == "" {
					obj[shortPred] = lit.Value
					continue
				}
			}

			var values []interface{}

			for _, objNode := range objects {
				value := nodeToJSONLDCompact(objNode, context)
				values = append(values, value)
			}

			if len(values) == 1 {
				obj[shortPred] = values[0]
			} else {
				obj[shortPred] = values
			}
		}

		graph = append(graph, obj)
	}

	result := map[string]interface{}{
		"@context": context,
		"@graph": graph,
	}

	bytes, err := json.MarshalIndent(result, "", " ")

	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func shortenURI(uri string, context map[string]string) string {
	for prefix, namespace := range context {
		if strings.HasPrefix(uri, namespace) {
			localPart := strings.TrimPrefix(uri, namespace)
			return prefix + ":" + localPart
		}
	}
	return uri
}

func nodeToJSONLDCompact(n triple.Node, context map[string]string) interface{} {
	switch node := n.(type) {
		case triple.IRI:
			shortened := shortenURI(node.Value, context)
			return map[string]interface{}{"@id": shortened}
		case triple.Literal:
			if node.Language == "" && node.Datatype == "" {
				return node.Value
			}
			result := map[string]interface{}{"@value": node.Value}
			if node.Language != "" {
				result["@language"] = node.Language
			}
			if node.Datatype != "" {
				result["@type"] = node.Datatype
			}
			return result
	case triple.BlankNode:
		return map[string]interface{}{"@id": "_:" + node.Value}
	}
	return nil
}
