# tripl

A Go CLI tool and library for encoding and decoding RDF triples in multiple serialization formats.

## Purpose

tripl is a standalone RDF triple encoder/decoder that can be used as:
- A command-line tool for converting between RDF formats
- A Go package imported by other applications (e.g., weave2 zettelkasten)

## Current Status

### Completed
- Core RDF data model (IRI, Literal, BlankNode, Triple)
- N-Triples encoder/decoder with full test coverage
- Turtle encoder (basic and compact with semicolons/commas/`a`)
- Turtle decoder with advanced syntax support
- JSON-LD encoder (expanded and compact with @context)
- JSON-LD decoder
- Round-trip validation for all formats
- CLI Phase 1: Create command with prefix support
- CLI Phase 2: Compact output flag
- Refactoring: Extract shared parsing logic to common utilities (parser.go)

### In Progress
- Code refactoring and optimization
  - Create prefix resolver type to centralize prefix handling
  - Add proper error context (line numbers, positions)
  - Optimize grouping algorithms and reduce allocations
  - Consolidate duplicate code (formatNode, expandURI, nodeKey)

### Planned
- CLI Phase 3: Add format conversion commands
- File I/O handling
- Batch conversion support

## Features

- Encode RDF triples to multiple formats:
  - N-Triples (.nt) ✓
  - Turtle (.ttl) ✓ (basic and compact encoders, advanced decoder)
  - JSON-LD (.jsonld) ✓ (expanded and compact with @context)
- Decode from these formats back to structured data
- Usable as both CLI and Go library

## Architecture

```
tripl/
├── cmd/tripl/          # CLI entry point (planned)
├── pkg/
│   ├── triple/         # Core RDF data structures
│   └── encoder/        # Format-specific encoders/decoders
└── internal/           # Internal helpers
```

## Use Cases

### Standalone CLI
```bash
# Create a triple with full URIs
tripl create --subject http://example.org/note1 \
             --predicate http://example.org/title \
             --object "My Note"

# Create with prefixes
tripl create --prefix ex=http://example.org/ \
             --subject ex:note1 \
             --predicate ex:title \
             --object "My Note" \
             --format turtle

# Output as JSON-LD
tripl create --subject http://example.org/note1 \
             --predicate http://example.org/title \
             --object "My Note" \
             --format jsonld

# Use compact format with @context
tripl create --prefix ex=http://example.org/ \
             --subject ex:note1 \
             --predicate ex:title \
             --object "My Note" \
             --format jsonld \
             --compact
```

### As a Library
```go
import "github.com/DeDude/tripl/pkg/encoder"
import "github.com/DeDude/tripl/pkg/triple"

// Create a triple
t := triple.Triple{
    Subject:   triple.IRI{Value: "http://example.org/note1"},
    Predicate: triple.IRI{Value: "http://example.org/title"},
    Object:    triple.Literal{Value: "My Note"},
}

// Encode to N-Triples
ntriples := encoder.EncodeNTriple(t)

// Decode from N-Triples
decoded, err := encoder.DecodeNTriple(ntriples)
```

## Related Projects

- **weave2**: A zettelkasten note-taking system that uses tripl for RDF operations
- Uses SKOS vocabulary for note relationships
- Custom weave vocabulary for zettelkasten-specific properties
