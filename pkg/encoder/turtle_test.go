package encoder

import (
	"testing"
	"github.com/DeDude/tripl/pkg/triple"
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
		name            string
		input           string
		expectedTriples []triple.Triple
		expectedPrefixes map[string]string
		wantErr         bool
	}{
		{
			name: "simple triple without prefixes",
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

