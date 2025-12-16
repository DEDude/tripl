package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	batch := convertFlags.Bool("batch", false, "Convert all files in input directory matching the source format extension")
	inputPath := convertFlags.String("input", "", "File path to read input from (default: stdin)")
	outputPath := convertFlags.String("output", "", "File path to write output (default: stdout)")
	force := convertFlags.Bool("force", false, "Allow overwriting existing output file")

	convertFlags.Parse(os.Args[2:])

	if *fromFormat == "" || *toFormat == "" {
		fmt.Fprintln(os.Stderr, "Error: --from and --to are required")
		convertFlags.Usage()
		os.Exit(1)
	}

	userPrefixes := parsePrefixes(*prefixFlag)

	if *batch {
		if *inputPath == "" {
			fmt.Fprintln(os.Stderr, "Error: --input directory is required in batch mode")
			os.Exit(1)
		}
		if err := convertBatch(strings.ToLower(*fromFormat), strings.ToLower(*toFormat), *compact, userPrefixes, *inputPath, *outputPath, *force); err != nil {
			fmt.Fprintf(os.Stderr, "Error converting batch: %v\n", err)
			os.Exit(1)
		}
		return
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

func convertBatch(fromFormat, toFormat string, compact bool, userPrefixes map[string]string, inputDir, outputDir string, force bool) error {
	fromExt, err := formatExtension(fromFormat)
	if err != nil {
		return err
	}
	toExt, err := formatExtension(toFormat)
	if err != nil {
		return err
	}

	if outputDir != "" {
		if err := ensureDir(outputDir); err != nil {
			return err
		}
	}

	files, err := collectBatchInputs(inputDir)
	if err != nil {
		return err
	}

	processed := 0

	for _, inPath := range files {
		if !strings.EqualFold(filepath.Ext(inPath), fromExt) {
			continue
		}

		inputBytes, err := os.ReadFile(inPath)
		if err != nil {
			return fmt.Errorf("reading %s: %w", inPath, err)
		}

		input := strings.TrimSpace(string(inputBytes))
		if input == "" {
			return fmt.Errorf("%s is empty", inPath)
		}

		triples, detectedPrefixes, err := decodeTriples(fromFormat, input)
		if err != nil {
			return fmt.Errorf("decoding %s: %w", inPath, err)
		}

		if detectedPrefixes == nil {
			detectedPrefixes = map[string]string{}
		}
		for k, v := range userPrefixes {
			detectedPrefixes[k] = v
		}

		output, err := encodeTriples(triples, toFormat, compact, detectedPrefixes)
		if err != nil {
			return fmt.Errorf("encoding %s: %w", inPath, err)
		}

		base := strings.TrimSuffix(filepath.Base(inPath), filepath.Ext(inPath))
		targetDir := outputDir
		if targetDir == "" {
			targetDir = filepath.Dir(inPath)
		}
		if err := ensureDir(targetDir); err != nil {
			return err
		}

		outPath := filepath.Join(targetDir, base+toExt)

		if !force {
			if _, err := os.Stat(outPath); err == nil {
				return fmt.Errorf("output file %s exists (use --force to overwrite)", outPath)
			} else if err != nil && !os.IsNotExist(err) {
				return err
			}
		}

		if err := os.WriteFile(outPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outPath, err)
		}

		processed++
	}

	if processed == 0 {
		return fmt.Errorf("no files with extension %s found in %s", fromExt, inputDir)
	}

	return nil
}

func formatExtension(format string) (string, error) {
	switch format {
	case "ntriples", "nt":
		return ".nt", nil
	case "turtle", "ttl":
		return ".ttl", nil
	case "jsonld":
		return ".jsonld", nil
	}
	return "", fmt.Errorf("unsupported format: %s", format)
}

func ensureDir(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("output path %s exists and is not a directory", path)
	}
	return nil
}

func collectBatchInputs(inputPattern string) ([]string, error) {
	if containsGlob(inputPattern) {
		matches, err := filepath.Glob(inputPattern)
		if err != nil {
			return nil, err
		}

		var files []string
		for _, m := range matches {
			info, err := os.Stat(m)
			if err != nil {
				return nil, err
			}
			if info.IsDir() {
				continue
			}
			files = append(files, m)
		}
		if len(files) == 0 {
			return nil, fmt.Errorf("no files match glob %s", inputPattern)
		}
		return files, nil
	}

	info, err := os.Stat(inputPattern)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("batch mode requires --input to be a directory or glob pattern")
	}

	entries, err := os.ReadDir(inputPattern)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		files = append(files, filepath.Join(inputPattern, entry.Name()))
	}
	return files, nil
}

func containsGlob(path string) bool {
	return strings.ContainsAny(path, "*?[")
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
	fmt.Println("  --batch                Convert all files in an input directory (requires --input dir)")
	fmt.Println("  --input string         File path to read input (default: stdin) or directory in batch mode")
	fmt.Println("  --output string        File path to write output (default: stdout) or directory in batch mode")
	fmt.Println("  --force                Allow overwriting existing output file")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tripl create --subject http://example.org/note1 --predicate http://example.org/title --object \"My Note\"")
	fmt.Println("  tripl create --prefix ex=http://example.org/ --subject ex:note1 --predicate ex:title --object \"My Note\" --format turtle")
	fmt.Println("  tripl create --prefix ex=http://example.org/ --subject ex:note1 --predicate ex:title --object \"My Note\" --format turtle --compact")
	fmt.Println("  cat input.ttl | tripl convert --from turtle --to jsonld")
	fmt.Println("  cat input.nt  | tripl convert --from ntriples --to turtle --compact --prefix ex=http://example.org/")
}
