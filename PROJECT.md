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
- CLI Phase 3: Format conversion command (stdin-based `convert` between ntriples/turtle/jsonld)
- File I/O handling for CLI (input/output flags, overwrite control)
- Batch conversion (directory or glob inputs) via `convert --batch`
- Refactoring: Extract shared parsing logic to common utilities (parser.go)
- Code refactoring and optimization
  - Created PrefixResolver type to centralize prefix handling
  - Added proper error context (line numbers, positions)
  - Optimized grouping algorithms and reduced allocations
  - Consolidated duplicate code (formatNode, nodeKey) into common.go

### In Progress
- None

### Planned
- Batch conversion support
- Relationship helper commands (create/manage RDF relationship triples)

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
