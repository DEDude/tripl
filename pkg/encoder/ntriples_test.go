package encoder

import (
	"github.com/DeDude/tripl/pkg/triple"
	"testing"
)

func TestEncodeNTriple(t *testing.T) {
	tests := []struct {
		name     string
		triple   triple.Triple
		expected string
	}{
		{
			name: "simple triple with IRIs and literal",
			triple: triple.Triple{
				Subject:   triple.IRI{Value: "http://example.org/note1"},
				Predicate: triple.IRI{Value: "http://example.org/title"},
				Object:    triple.Literal{Value: "My Note"},
			},
			expected: `<http://example.org/note1> <http://example.org/title> "My Note" .`,
		},
		{
			name: "literal with language tag",
			triple: triple.Triple{
				Subject:   triple.IRI{Value: "http://example.org/note1"},
				Predicate: triple.IRI{Value: "http://example.org/title"},
				Object:    triple.Literal{Value: "Hello", Language: "en"},
			},
			expected: `<http://example.org/note1> <http://example.org/title> "Hello"@en .`,
		},
		{
			name: "literal with datatype",
			triple: triple.Triple{
				Subject:   triple.IRI{Value: "http://example.org/note1"},
				Predicate: triple.IRI{Value: "http://example.org/count"},
				Object:    triple.Literal{Value: "42", Datatype: "http://www.w3.org/2001/XMLSchema#integer"},
			},
			expected: `<http://example.org/note1> <http://example.org/count> "42"^^<http://www.w3.org/2001/XMLSchema#integer> .`,
		},
		{
			name: "blank node",
			triple: triple.Triple{
				Subject:   triple.BlankNode{Value: "b1"},
				Predicate: triple.IRI{Value: "http://example.org/title"},
				Object:    triple.Literal{Value: "Test"},
			},
			expected: `_:b1 <http://example.org/title> "Test" .`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeNTriple(tt.triple)
			if result != tt.expected {
				t.Errorf("EncodeNTriple() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestDecodeNTriple(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected triple.Triple
		wantErr  bool
	}{
		{
			name:  "simple triple with IRIs and literal",
			input: `<http://example.org/note1> <http://example.org/title> "My Note" .`,
			expected: triple.Triple{
				Subject:   triple.IRI{Value: "http://example.org/note1"},
				Predicate: triple.IRI{Value: "http://example.org/title"},
				Object:    triple.Literal{Value: "My Note"},
			},
			wantErr: false,
		},
		{
			name:  "literal with language tag",
			input: `<http://example.org/note1> <http://example.org/title> "Hello"@en .`,
			expected: triple.Triple{
				Subject:   triple.IRI{Value: "http://example.org/note1"},
				Predicate: triple.IRI{Value: "http://example.org/title"},
				Object:    triple.Literal{Value: "Hello", Language: "en"},
			},
			wantErr: false,
		},
		{
			name:  "literal with datatype",
			input: `<http://example.org/note1> <http://example.org/count> "42"^^<http://www.w3.org/2001/XMLSchema#integer> .`,
			expected: triple.Triple{
				Subject:   triple.IRI{Value: "http://example.org/note1"},
				Predicate: triple.IRI{Value: "http://example.org/count"},
				Object:    triple.Literal{Value: "42", Datatype: "http://www.w3.org/2001/XMLSchema#integer"},
			},
			wantErr: false,
		},
		{
			name:  "blank node",
			input: `_:b1 <http://example.org/title> "Test" .`,
			expected: triple.Triple{
				Subject:   triple.BlankNode{Value: "b1"},
				Predicate: triple.IRI{Value: "http://example.org/title"},
				Object:    triple.Literal{Value: "Test"},
			},
			wantErr: false,
		},
		{
			name:     "missing period",
			input:    `<http://example.org/note1> <http://example.org/title> "Test"`,
			expected: triple.Triple{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DecodeNTriple(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecodeNTriple() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !triplesEqual(result, tt.expected) {
				t.Errorf("DecodeNTriple() = %+v, want %+v", result, tt.expected)
			}
		})
	}
}

func triplesEqual(a, b triple.Triple) bool {
	return nodesEqual(a.Subject, b.Subject) &&
		nodesEqual(a.Predicate, b.Predicate) &&
		nodesEqual(a.Object, b.Object)
}

func nodesEqual(a, b triple.Node) bool {
	switch aNode := a.(type) {
	case triple.IRI:
		bNode, ok := b.(triple.IRI)
		return ok && aNode.Value == bNode.Value
	case triple.Literal:
		bNode, ok := b.(triple.Literal)
		return ok && aNode.Value == bNode.Value &&
			aNode.Language == bNode.Language &&
			aNode.Datatype == bNode.Datatype
	case triple.BlankNode:
		bNode, ok := b.(triple.BlankNode)
		return ok && aNode.Value == bNode.Value
	}
	return false
}
