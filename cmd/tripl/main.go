package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/DeDude/tripl/pkg/encoder"
	"github.com/DeDude/tripl/pkg/triple"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "create":
		createCommand()
	case "convert":
		convertCommand()
	case "help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()

		os.Exit(1)
	}
}

func createCommand() {
	createFlags := flag.NewFlagSet("create", flag.ExitOnError)

	format := createFlags.String("format", "turtle", "Output format: ntriples, turtle, jsonld")
	subject := createFlags.String("subject", "", "Subject IRI")
	predicate := createFlags.String("predicate", "", "Predicate IRI")
	object := createFlags.String("object", "", "Object value")
	objectType := createFlags.String("object-type", "literal", "Object type: literal. iri, blank")
	language := createFlags.String("language", "", "Language tag for literal (optional)")
	datatype := createFlags.String("datatype", "", "Datatype IRI for literal (optional)")
	prefixFlag := createFlags.String("prefix", "", "Prefix definitions (format: prefix=uri, prefix2=uri2)")
	compact := createFlags.Bool("compact", false, "Use compact output format (for turtle and jsonld)")
	outputPath := createFlags.String("output", "", "File path to write output (default: stdout)")
	force := createFlags.Bool("force", false, "Allow overwriting existing output file")

	createFlags.Parse(os.Args[2:])

	if *subject == "" || *predicate == "" || *object == "" {
		fmt.Fprintln(os.Stderr, "Error: subject, predicate, and object are required")
		createFlags.Usage()
		os.Exit(1)
	}

	prefixes := parsePrefixes(*prefixFlag)
	resolver := encoder.NewPrefixResolver(prefixes)
	subjectIRI := resolver.Expand(*subject)
	predicateIRI := resolver.Expand(*predicate)

	t := triple.Triple{
		Subject:   triple.IRI{Value: subjectIRI},
		Predicate: triple.IRI{Value: predicateIRI},
	}

	switch *objectType {
	case "literal":
		lit := triple.Literal{Value: *object}
		if *language != "" {
			lit.Language = *language
		}
		if *datatype != "" {
			lit.Datatype = resolver.Expand(*datatype)
		}
		t.Object = lit
	case "iri":
		t.Object = triple.IRI{Value: resolver.Expand(*object)}
	case "blank":
		t.Object = triple.BlankNode{Value: *object}
	default:
		fmt.Fprintf(os.Stderr, "Unknown object type: %s\n", *objectType)
		os.Exit(1)
	}

	triples := []triple.Triple{t}

	var output string
	var err error

	switch *format {
	case "ntriples":
		output = encoder.EncodeNTriple(t)
	case "turtle":
		if *compact {
			output = encoder.EncodeTurtleCompact(triples, prefixes)
		} else {
			output = encoder.EncodeTurtle(triples, prefixes)
		}
	case "jsonld":
		if *compact {
			output, err = encoder.EncodeJSONLDCompact(triples, prefixes)
		} else {
			output, err = encoder.EncodeJSONLD(triples)
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding JSON-LD: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Unknown format: %s\n", *format)
		os.Exit(1)
	}

	if err := writeOutput(output, *outputPath, *force); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
}

func convertCommand() {
	convertFlags := flag.NewFlagSet("convert", flag.ExitOnError)

	fromFormat := convertFlags.String("from", "", "Input format: ntriples, turtle, jsonld")
	toFormat := convertFlags.String("to", "", "Output format: ntriples, turtle, jsonld")
	compact := convertFlags.Bool("compact", false, "Use compact output format (turtle/jsonld)")
	prefixFlag := convertFlags.String("prefix", "", "Prefix definitions for output (format: prefix=uri, prefix2=uri2)")
	inputPath := convertFlags.String("input", "", "File path to read input from (default: stdin)")
	outputPath := convertFlags.String("output", "", "File path to write output (default: stdout)")
	force := convertFlags.Bool("force", false, "Allow overwriting existing output file")

	convertFlags.Parse(os.Args[2:])

	if *fromFormat == "" || *toFormat == "" {
		fmt.Fprintln(os.Stderr, "Error: --from and --to are required")
		convertFlags.Usage()
		os.Exit(1)
	}

	inputBytes, err := readInput(*inputPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	input := strings.TrimSpace(string(inputBytes))
	if input == "" {
		fmt.Fprintln(os.Stderr, "Error: no input provided")
		os.Exit(1)
	}

	userPrefixes := parsePrefixes(*prefixFlag)

	triples, detectedPrefixes, err := decodeTriples(strings.ToLower(*fromFormat), input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding input: %v\n", err)
		os.Exit(1)
	}

	if detectedPrefixes == nil {
		detectedPrefixes = map[string]string{}
	}
	for k, v := range userPrefixes {
		detectedPrefixes[k] = v
	}

	output, err := encodeTriples(triples, strings.ToLower(*toFormat), *compact, detectedPrefixes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding output: %v\n", err)
		os.Exit(1)
	}

	if err := writeOutput(output, *outputPath, *force); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}
}

func decodeTriples(format, data string) ([]triple.Triple, map[string]string, error) {
	switch format {
	case "ntriples", "nt":
		var triples []triple.Triple
		lines := strings.Split(data, "\n")
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "#") {
				continue
			}
			t, err := encoder.DecodeNTriple(trimmed)
			if err != nil {
				return nil, nil, fmt.Errorf("line %q: %w", trimmed, err)
			}
			triples = append(triples, t)
		}
		return triples, map[string]string{}, nil
	case "turtle", "ttl":
		triples, prefixes, err := encoder.DecodeTurtle(data)
		return triples, prefixes, err
	case "jsonld":
		triples, err := encoder.DecodeJSONLD(data)
		return triples, map[string]string{}, err
	default:
		return nil, nil, fmt.Errorf("unsupported input format: %s", format)
	}
}

func encodeTriples(triples []triple.Triple, format string, compact bool, prefixes map[string]string) (string, error) {
	switch format {
	case "ntriples", "nt":
		var b strings.Builder
		for i, t := range triples {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(encoder.EncodeNTriple(t))
		}
		return b.String(), nil
	case "turtle", "ttl":
		if compact {
			return encoder.EncodeTurtleCompact(triples, prefixes), nil
		}
		return encoder.EncodeTurtle(triples, prefixes), nil
	case "jsonld":
		if compact {
			return encoder.EncodeJSONLDCompact(triples, prefixes)
		}
		return encoder.EncodeJSONLD(triples)
	default:
		return "", fmt.Errorf("unsupported output format: %s", format)
	}
}

func parsePrefixes(prefixStr string) map[string]string {
	prefixes := make(map[string]string)

	if prefixStr == "" {
		return prefixes
	}

	pairs := strings.Split(prefixStr, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) == 2 {
			prefixes[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}

	return prefixes
}

func readInput(path string) ([]byte, error) {
	if path == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(path)
}

func writeOutput(data string, path string, force bool) error {
	if path == "" {
		fmt.Print(data)
		return nil
	}

	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("output file %s exists (use --force to overwrite)", path)
		} else if !os.IsNotExist(err) {
			return err
		}
	}

	return os.WriteFile(path, []byte(data), 0644)
}

func printUsage() {
	fmt.Println("tripl - RDF triple encoder/decoder")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tripl create [flags]")
	fmt.Println("  tripl convert [flags] < input")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create    Create a triple and output in specified format")
	fmt.Println("  convert   Convert triples between formats (reads from stdin)")
	fmt.Println("  help      Show this help message")
	fmt.Println()
	fmt.Println("Create flags:")
	fmt.Println("  --format string        Output format: ntriples, turtle, jsonld (default: turtle)")
	fmt.Println("  --subject string       Subject IRI (required)")
	fmt.Println("  --predicate string     Predicate IRI (required)")
	fmt.Println("  --object string        Object value (required)")
	fmt.Println("  --object-type string   Object type: literal, iri, blank (default: literal)")
	fmt.Println("  --language string      Language tag for literal")
	fmt.Println("  --datatype string      Datatype IRI for literal")
	fmt.Println("  --prefix string        Prefix definitions (format: ex=http://example.org/)")
	fmt.Println("  --compact              Use compact output format (turtle: semicolons/commas, jsonld: @context)")
	fmt.Println("  --output string        File path to write output (default: stdout)")
	fmt.Println("  --force                Allow overwriting existing output file")
	fmt.Println()
	fmt.Println("Convert flags:")
	fmt.Println("  --from string          Input format: ntriples, turtle, jsonld (required)")
	fmt.Println("  --to string            Output format: ntriples, turtle, jsonld (required)")
	fmt.Println("  --prefix string        Prefix definitions for output (format: ex=http://example.org/)")
	fmt.Println("  --compact              Use compact output format (turtle/jsonld)")
	fmt.Println("  --input string         File path to read input (default: stdin)")
	fmt.Println("  --output string        File path to write output (default: stdout)")
	fmt.Println("  --force                Allow overwriting existing output file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tripl create --subject http://example.org/note1 --predicate http://example.org/title --object \"My Note\"")
	fmt.Println("  tripl create --prefix ex=http://example.org/ --subject ex:note1 --predicate ex:title --object \"My Note\" --format turtle")
	fmt.Println("  tripl create --prefix ex=http://example.org/ --subject ex:note1 --predicate ex:title --object \"My Note\" --format turtle --compact")
	fmt.Println("  cat input.ttl | tripl convert --from turtle --to jsonld")
	fmt.Println("  cat input.nt  | tripl convert --from ntriples --to turtle --compact --prefix ex=http://example.org/")
}
