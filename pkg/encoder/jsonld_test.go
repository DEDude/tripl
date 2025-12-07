package encoder

import (
	"encoding/json"
	"github.com/DeDude/tripl/pkg/triple"
	"testing"
)

func TestEncodeJSONLD(t *testing.T) {
	tests := []struct {
		name    string
		triples []triple.Triple
		wantErr bool
	}{
		{
			name: "single triple",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "My Note"},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple triples same subject",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "Test"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/author"},
					Object:    triple.Literal{Value: "John"},
				},
			},
			wantErr: false,
		},
		{
			name: "literal with language tag",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "Hello", Language: "en"},
				},
			},
			wantErr: false,
		},
		{
			name: "literal with datatype",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/count"},
					Object:    triple.Literal{Value: "42", Datatype: "http://www.w3.org/2001/XMLSchema#integer"},
				},
			},
			wantErr: false,
		},
		{
			name: "IRI object",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/author"},
					Object:    triple.IRI{Value: "http://example.org/person1"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodeJSONLD(tt.triples)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeJSONLD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify it's valid JSON
				var parsed []map[string]interface{}
				if err := json.Unmarshal([]byte(result), &parsed); err != nil {
					t.Errorf("EncodeJSONLD() produced invalid JSON: %v", err)
				}

				// Verify we have the right number of subjects
				if len(parsed) == 0 {
					t.Error("EncodeJSONLD() produced empty result")
				}
			}
		})
	}
}

func TestDecodeJSONLD(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedTriples []triple.Triple
		wantErr         bool
	}{
		{
			name: "single triple",
			input: `[
  {
    "@id": "http://example.org/note1",
    "http://example.org/title": [
      {"@value": "My Note"}
    ]
  }
]`,
			expectedTriples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "My Note"},
				},
			},
			wantErr: false,
		},
		{
			name: "literal with language tag",
			input: `[
  {
    "@id": "http://example.org/note1",
    "http://example.org/title": [
      {"@value": "Hello", "@language": "en"}
    ]
  }
]`,
			expectedTriples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "Hello", Language: "en"},
				},
			},
			wantErr: false,
		},
		{
			name: "IRI object",
			input: `[
  {
    "@id": "http://example.org/note1",
    "http://example.org/author": [
      {"@id": "http://example.org/person1"}
    ]
  }
]`,
			expectedTriples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/author"},
					Object:    triple.IRI{Value: "http://example.org/person1"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triples, err := DecodeJSONLD(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeJSONLD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(triples) != len(tt.expectedTriples) {
					t.Errorf("DecodeJSONLD() got %d triples, want %d", len(triples), len(tt.expectedTriples))
					return
				}

				for i, expected := range tt.expectedTriples {
					if !triplesEqual(triples[i], expected) {
						t.Errorf("DecodeJSONLD() triple[%d] = %+v, want %+v", i, triples[i], expected)
					}
				}
			}
		})
	}
}

func TestJSONLDRoundTrip(t *testing.T) {
	originalTriples := []triple.Triple{
		{
			Subject:   triple.IRI{Value: "http://example.org/note1"},
			Predicate: triple.IRI{Value: "http://example.org/title"},
			Object:    triple.Literal{Value: "Test"},
		},
		{
			Subject:   triple.IRI{Value: "http://example.org/note1"},
			Predicate: triple.IRI{Value: "http://example.org/author"},
			Object:    triple.Literal{Value: "John", Language: "en"},
		},
		{
			Subject:   triple.IRI{Value: "http://example.org/note1"},
			Predicate: triple.IRI{Value: "http://example.org/related"},
			Object:    triple.IRI{Value: "http://example.org/note2"},
		},
		{
			Subject:   triple.IRI{Value: "http://example.org/note2"},
			Predicate: triple.IRI{Value: "http://example.org/title"},
			Object:    triple.Literal{Value: "Second Note"},
		},
	}

	// Encode to JSON-LD
	encoded, err := EncodeJSONLD(originalTriples)
	if err != nil {
		t.Fatalf("EncodeJSONLD() error = %v", err)
	}

	// Decode it back
	decodedTriples, err := DecodeJSONLD(encoded)
	if err != nil {
		t.Fatalf("DecodeJSONLD() error = %v", err)
	}

	// Check we got the same number of triples
	if len(decodedTriples) != len(originalTriples) {
		t.Errorf("Round trip produced %d triples, want %d", len(decodedTriples), len(originalTriples))
	}

	// Check each triple matches
	for i, original := range originalTriples {
		found := false
		for _, decoded := range decodedTriples {
			if triplesEqual(original, decoded) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Original triple %d not found in decoded triples: %+v", i, original)
		}
	}
}

func TestEncodeJSONLDCompact(t *testing.T) {
	tests := []struct {
		name    string
		triples []triple.Triple
		context map[string]string
		wantErr bool
	}{
		{
			name: "single triple with context",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "My Note"},
				},
			},
			context: map[string]string{
				"ex": "http://example.org/",
			},
			wantErr: false,
		},
		{
			name: "multiple triples same subject",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "Test"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/author"},
					Object:    triple.Literal{Value: "John"},
				},
			},
			context: map[string]string{
				"ex": "http://example.org/",
			},
			wantErr: false,
		},
		{
			name: "literal with language tag",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "Hello", Language: "en"},
				},
			},
			context: map[string]string{
				"ex": "http://example.org/",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := EncodeJSONLDCompact(tt.triples, tt.context)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncodeJSONLDCompact() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify it's valid JSON
				var parsed map[string]interface{}
				if err := json.Unmarshal([]byte(result), &parsed); err != nil {
					t.Errorf("EncodeJSONLDCompact() produced invalid JSON: %v", err)
				}

				// Verify @context exists
				if _, ok := parsed["@context"]; !ok {
					t.Error("EncodeJSONLDCompact() missing @context")
				}

				// Verify @graph exists
				if _, ok := parsed["@graph"]; !ok {
					t.Error("EncodeJSONLDCompact() missing @graph")
				}
			}
		})
	}
}
