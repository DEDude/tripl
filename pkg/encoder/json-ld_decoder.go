package encoder

import (
	"encoding/json"
	"github.com/DeDude/tripl/pkg/triple"
)

func DecodeJSONLD(input string) ([]triple.Triple, error) {
	var data []map[string]interface{}
	
	if err := json.Unmarshal([]byte(input), &data); err != nil {
		return nil, err
	}
	
	var triples []triple.Triple
	
	for _, obj := range data {
		subjectID, ok := obj["@id"].(string)
		if !ok {
			continue
		}
		
		subject := triple.IRI{Value: subjectID}
		
		for key, value := range obj {
			if key == "@id" {
				continue
			}
			
			predicate := triple.IRI{Value: key}
			
			values, ok := value.([]interface{})
			if !ok {
				continue
			}
			
			for _, v := range values {
				objMap, ok := v.(map[string]interface{})
				if !ok {
					continue
				}
				
				object := jsonLDToNode(objMap)
				if object != nil {
					triples = append(triples, triple.Triple{
						Subject:   subject,
						Predicate: predicate,
						Object:    object,
					})
				}
			}
		}
	}
	
	return triples, nil
}

func jsonLDToNode(obj map[string]interface{}) triple.Node {
	if id, ok := obj["@id"].(string); ok {
		if len(id) > 2 && id[:2] == "_:" {
			return triple.BlankNode{Value: id[2:]}
		}
		return triple.IRI{Value: id}
	}
	
	if value, ok := obj["@value"].(string); ok {
		lit := triple.Literal{Value: value}
		
		if lang, ok := obj["@language"].(string); ok {
			lit.Language = lang
		}
		
		if dtype, ok := obj["@type"].(string); ok {
			lit.Datatype = dtype
		}
		
		return lit
	}
	
	return nil
}
