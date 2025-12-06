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
- Turtle decoder with advanced syntax support (semicolons, commas, `a` shortcut)
- Basic Turtle encoder

### In Progress
- Compact Turtle encoder (using semicolons and commas for grouping)

### Planned
- JSON-LD encoder/decoder
- CLI implementation
- File I/O handling
- Batch conversion support

## Features

- Encode RDF triples to multiple formats:
  - N-Triples (.nt) ✓
  - Turtle (.ttl) ✓ (basic encoder, advanced decoder with semicolons/commas/`a`)
  - JSON-LD (.jsonld) - planned
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
tripl encode --format turtle input.nt
tripl decode --format jsonld data.ttl
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
