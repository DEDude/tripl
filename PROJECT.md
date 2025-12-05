# tripl

A Go CLI tool and library for encoding and decoding RDF triples in multiple serialization formats.

## Purpose

tripl is a standalone RDF triple encoder/decoder that can be used as:
- A command-line tool for converting between RDF formats
- A Go package imported by other applications (e.g., weave2 zettelkasten)

## Features

- Encode RDF triples to multiple formats:
  - Turtle (.ttl)
  - N-Triples (.nt)
  - JSON-LD (.jsonld)
- Decode from these formats back to structured data
- Usable as both CLI and Go library

## Architecture

```
tripl/
├── cmd/tripl/          # CLI entry point
├── pkg/encoder/        # Core encoding/decoding logic
└── internal/           # Internal helpers
```

## Use Cases

### Standalone CLI
```bash
tripl encode --format turtle input.json
tripl decode --format jsonld data.ttl
```

### As a Library
```go
import "github.com/yourusername/tripl/pkg/encoder"

// Use encoder package in weave2 or other apps
```

## Related Projects

- **weave2**: A zettelkasten note-taking system that uses tripl for RDF operations
- Uses SKOS vocabulary for note relationships
- Custom weave vocabulary for zettelkasten-specific properties
