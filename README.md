# tripl

A Go CLI and library for encoding and decoding RDF triples across N-Triples, Turtle, and JSON-LD. Use it to convert data on the command line or embed it as a package.

## Features
- Encode/decode RDF triples: N-Triples (`.nt`), Turtle (`.ttl`), JSON-LD (`.jsonld`)
- CLI `create` and `convert` commands with prefix support and compact output options
- Batch conversion (`--batch`) for directories or glob patterns while reusing single-file logic
- Library API for building triples and converting between formats

## Install
Requires Go 1.22+.

```bash
go install github.com/DeDude/tripl/cmd/tripl@latest
# or build locally
go build -o tripl ./cmd/tripl
```

## CLI Usage
### Create a triple
```bash
tripl create \
  --subject http://example.org/note1 \
  --predicate http://example.org/title \
  --object "My Note" \
  --format turtle
```

With prefixes and compact output:
```bash
tripl create \
  --prefix ex=http://example.org/ \
  --subject ex:note1 \
  --predicate ex:title \
  --object "My Note" \
  --format jsonld \
  --compact
```

### Convert between formats
Single input (stdin or file):
```bash
cat input.ttl | tripl convert --from turtle --to jsonld --compact --prefix ex=http://example.org/
# or
tripl convert --from ntriples --to turtle --input input.nt --output output.ttl
```

### Batch conversion (directory or glob)
Process all matching files; outputs go alongside sources unless `--output` points to a directory:
```bash
tripl convert --batch --from ttl --to jsonld \
  --input "data/**/*.ttl" \
  --output out \
  --compact
```
- `--input` accepts a directory or glob.
- Files are filtered by the `--from` format’s extension.
- `--force` allows overwriting existing outputs.

Run `tripl help` for full flag descriptions.

## Library Usage
```go
import (
    "github.com/DeDude/tripl/pkg/encoder"
    "github.com/DeDude/tripl/pkg/triple"
)

func example() error {
    t := triple.Triple{
        Subject:   triple.IRI{Value: "http://example.org/note1"},
        Predicate: triple.IRI{Value: "http://example.org/title"},
        Object:    triple.Literal{Value: "My Note"},
    }

    // Encode to N-Triples
    nt := encoder.EncodeNTriple(t)

    // Decode from N-Triples
    decoded, err := encoder.DecodeNTriple(nt)
    if err != nil {
        return err
    }
    _ = decoded

    // Encode to compact Turtle with prefixes
    ttl := encoder.EncodeTurtleCompact([]triple.Triple{t}, map[string]string{
        "ex": "http://example.org/",
    })
    _ = ttl
    return nil
}
```

## Development
- Format: `gofmt -w .`
- Test: `go test ./...`

## Status
Core formats, CLI (including batch conversion), and library APIs are complete. Open to extensions for additional commands or formats.***
