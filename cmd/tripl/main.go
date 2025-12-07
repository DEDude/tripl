package main

import (
	"flag"
	"fmt"
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

	createFlags.Parse(os.Args[2:])

	if *subject == "" || *predicate == "" || *object == "" {
		fmt.Fprintln(os.Stderr, "Error: subject, predicate, and object are required")
		createFlags.Usage()
		os.Exit(1)
	}

	prefixes := parsePrefixes(*prefixFlag)
	subjectIRI := expandURI(*subject, prefixes)
	predicateIRI := expandURI(*predicate, prefixes)

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
			lit.Datatype = expandURI(*datatype, prefixes)
		}
		t.Object = lit
	case "iri":
		t.Object = triple.IRI{Value: expandURI(*object, prefixes)}
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

	fmt.Print(output)
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

func expandURI(uri string, prefixes map[string]string) string {
	if !strings.Contains(uri, ":") {
		return uri
	}

	parts := strings.SplitN(uri, ":", 2)
	if len(parts) == 2 {
		prefix := parts[0]
		localPart := parts[1]

		if namespace, ok := prefixes[prefix]; ok {
			return namespace + localPart
		}
	}

	return uri
}

func printUsage() {
	fmt.Println("tripl - RDF triple encoder/decoder")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  tripl create [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  create    Create a triple and output in specified format")
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
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  tripl create --subject http://example.org/note1 --predicate http://example.org/title --object \"My Note\"")
	fmt.Println("  tripl create --prefix ex=http://example.org/ --subject ex:note1 --predicate ex:title --object \"My Note\" --format turtle")
	fmt.Println("  tripl create --prefix ex=http://example.org/ --subject ex:note1 --predicate ex:title --object \"My Note\" --format turtle --compact")
}

