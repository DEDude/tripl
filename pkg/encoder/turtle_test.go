package encoder

import (
	"github.com/DeDude/tripl/pkg/triple"
	"testing"
)

func TestEncodeTurtle(t *testing.T) {
	tests := []struct {
		name     string
		triples  []triple.Triple
		prefixes map[string]string
		expected string
	}{
		{
			name: "simple triple without prefixes",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "My Note"},
				},
			},
			prefixes: map[string]string{},
			expected: `<http://example.org/note1> <http://example.org/title> "My Note" .
`,
		},
		{
			name: "triple with prefixes",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "My Note"},
				},
			},
			prefixes: map[string]string{
				"ex": "http://example.org/",
			},
			expected: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "My Note" .
`,
		},
		{
			name: "multiple triples with prefixes",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "First Note"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note2"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "Second Note"},
				},
			},
			prefixes: map[string]string{
				"ex": "http://example.org/",
			},
			expected: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "First Note" .
ex:note2 ex:title "Second Note" .
`,
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
			prefixes: map[string]string{
				"ex": "http://example.org/",
			},
			expected: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "Hello"@en .
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeTurtle(tt.triples, tt.prefixes)
			if result != tt.expected {
				t.Errorf("EncodeTurtle() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDecodeTurtle(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedTriples  []triple.Triple
		expectedPrefixes map[string]string
		wantErr          bool
	}{
		{
			name:  "simple triple without prefixes",
			input: `<http://example.org/note1> <http://example.org/title> "My Note" .`,
			expectedTriples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "My Note"},
				},
			},
			expectedPrefixes: map[string]string{},
			wantErr:          false,
		},
		{
			name: "triple with prefixes",
			input: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "My Note" .`,
			expectedTriples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "My Note"},
				},
			},
			expectedPrefixes: map[string]string{
				"ex": "http://example.org/",
			},
			wantErr: false,
		},
		{
			name: "multiple triples with prefixes",
			input: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "First Note" .
ex:note2 ex:title "Second Note" .`,
			expectedTriples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "First Note"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note2"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "Second Note"},
				},
			},
			expectedPrefixes: map[string]string{
				"ex": "http://example.org/",
			},
			wantErr: false,
		},
		{
			name: "semicolon syntax - shared subject",
			input: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "Test" ;
         ex:author "John" .`,
			expectedTriples: []triple.Triple{
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
			expectedPrefixes: map[string]string{
				"ex": "http://example.org/",
			},
			wantErr: false,
		},
		{
			name: "comma syntax - shared subject and predicate",
			input: `@prefix ex: <http://example.org/> .

ex:note1 ex:tag "work", "important" .`,
			expectedTriples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/tag"},
					Object:    triple.Literal{Value: "work"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/tag"},
					Object:    triple.Literal{Value: "important"},
				},
			},
			expectedPrefixes: map[string]string{
				"ex": "http://example.org/",
			},
			wantErr: false,
		},
		{
			name: "a shortcut for rdf:type",
			input: `@prefix ex: <http://example.org/> .

ex:note1 a ex:Note .`,
			expectedTriples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
					Object:    triple.IRI{Value: "http://example.org/Note"},
				},
			},
			expectedPrefixes: map[string]string{
				"ex": "http://example.org/",
			},
			wantErr: false,
		},
		{
			name: "combined semicolons and commas",
			input: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "Test" ;
         ex:tag "work", "important" ;
         ex:author "John" .`,
			expectedTriples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "Test"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/tag"},
					Object:    triple.Literal{Value: "work"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/tag"},
					Object:    triple.Literal{Value: "important"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/author"},
					Object:    triple.Literal{Value: "John"},
				},
			},
			expectedPrefixes: map[string]string{
				"ex": "http://example.org/",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triples, prefixes, err := DecodeTurtle(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeTurtle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if len(triples) != len(tt.expectedTriples) {
					t.Errorf("DecodeTurtle() got %d triples, want %d", len(triples), len(tt.expectedTriples))
					return
				}
				for i, tr := range triples {
					if !triplesEqual(tr, tt.expectedTriples[i]) {
						t.Errorf("DecodeTurtle() triple[%d] = %+v, want %+v", i, tr, tt.expectedTriples[i])
					}
				}
				if len(prefixes) != len(tt.expectedPrefixes) {
					t.Errorf("DecodeTurtle() got %d prefixes, want %d", len(prefixes), len(tt.expectedPrefixes))
				}
				for k, v := range tt.expectedPrefixes {
					if prefixes[k] != v {
						t.Errorf("DecodeTurtle() prefix[%s] = %s, want %s", k, prefixes[k], v)
					}
				}
			}
		})
	}
}

func TestEncodeTurtleCompact(t *testing.T) {
	tests := []struct {
		name     string
		triples  []triple.Triple
		prefixes map[string]string
		expected string
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
			prefixes: map[string]string{
				"ex": "http://example.org/",
			},
			expected: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "My Note" .
`,
		},
		{
			name: "shared subject with semicolons",
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
			prefixes: map[string]string{
				"ex": "http://example.org/",
			},
			expected: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "Test" ;
         ex:author "John" .
`,
		},
		{
			name: "shared subject and predicate with commas",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/tag"},
					Object:    triple.Literal{Value: "work"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/tag"},
					Object:    triple.Literal{Value: "important"},
				},
			},
			prefixes: map[string]string{
				"ex": "http://example.org/",
			},
			expected: `@prefix ex: <http://example.org/> .

ex:note1 ex:tag "work", "important" .
`,
		},
		{
			name: "rdf:type uses 'a' shortcut",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
					Object:    triple.IRI{Value: "http://example.org/Note"},
				},
			},
			prefixes: map[string]string{
				"ex": "http://example.org/",
			},
			expected: `@prefix ex: <http://example.org/> .

ex:note1 a ex:Note .
`,
		},
		{
			name: "combined semicolons and commas",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "Test"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/tag"},
					Object:    triple.Literal{Value: "work"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/tag"},
					Object:    triple.Literal{Value: "important"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/author"},
					Object:    triple.Literal{Value: "John"},
				},
			},
			prefixes: map[string]string{
				"ex": "http://example.org/",
			},
			expected: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "Test" ;
         ex:tag "work", "important" ;
         ex:author "John" .
`,
		},
		{
			name: "multiple subjects",
			triples: []triple.Triple{
				{
					Subject:   triple.IRI{Value: "http://example.org/note1"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "First"},
				},
				{
					Subject:   triple.IRI{Value: "http://example.org/note2"},
					Predicate: triple.IRI{Value: "http://example.org/title"},
					Object:    triple.Literal{Value: "Second"},
				},
			},
			prefixes: map[string]string{
				"ex": "http://example.org/",
			},
			expected: `@prefix ex: <http://example.org/> .

ex:note1 ex:title "First" .

ex:note2 ex:title "Second" .
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeTurtleCompact(tt.triples, tt.prefixes)
			if result != tt.expected {
				t.Errorf("EncodeTurtleCompact() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestTurtleCompactRoundTrip(t *testing.T) {
	originalTriples := []triple.Triple{
		{
			Subject:   triple.IRI{Value: "http://example.org/note1"},
			Predicate: triple.IRI{Value: "http://example.org/title"},
			Object:    triple.Literal{Value: "Test"},
		},
		{
			Subject:   triple.IRI{Value: "http://example.org/note1"},
			Predicate: triple.IRI{Value: "http://example.org/tag"},
			Object:    triple.Literal{Value: "work"},
		},
		{
			Subject:   triple.IRI{Value: "http://example.org/note1"},
			Predicate: triple.IRI{Value: "http://example.org/tag"},
			Object:    triple.Literal{Value: "important"},
		},
		{
			Subject:   triple.IRI{Value: "http://example.org/note1"},
			Predicate: triple.IRI{Value: "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"},
			Object:    triple.IRI{Value: "http://example.org/Note"},
		},
		{
			Subject:   triple.IRI{Value: "http://example.org/note2"},
			Predicate: triple.IRI{Value: "http://example.org/title"},
			Object:    triple.Literal{Value: "Second Note"},
		},
	}

	prefixes := map[string]string{
		"ex": "http://example.org/",
	}

	// Encode to compact Turtle
	encoded := EncodeTurtleCompact(originalTriples, prefixes)

	// Decode it back
	decodedTriples, decodedPrefixes, err := DecodeTurtle(encoded)
	if err != nil {
		t.Fatalf("DecodeTurtle() error = %v", err)
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

	// Check prefixes
	if decodedPrefixes["ex"] != prefixes["ex"] {
		t.Errorf("Prefix 'ex' = %s, want %s", decodedPrefixes["ex"], prefixes["ex"])
	}
}
