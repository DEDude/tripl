package encoder

import "strings"

type PrefixResolver struct {
	prefixes map[string]string
}

func NewPrefixResolver(prefixes map[string]string) *PrefixResolver {
	return &PrefixResolver{prefixes: prefixes}
}

func (pr *PrefixResolver) Expand(uri string) string {
	if !strings.Contains(uri, ":") {
		return uri
	}

	parts := strings.SplitN(uri, ":", 2)
	if len(parts) == 2 {
		if namespace, ok := pr.prefixes[parts[0]] ; ok {
			return namespace + parts[1]
		}
	}
	return uri
}

func (pr *PrefixResolver) Shorten(iri string) string {
	for prefix, uri := range pr.prefixes {
		if strings.HasPrefix(iri, uri) {
			return prefix + ":" + strings.TrimPrefix(iri, uri)
		}
	}
	return iri
}

func (pr *PrefixResolver) Get(prefix string) (string, bool) {
	uri, ok := pr.prefixes[prefix]
	return uri, ok
}

func (pr *PrefixResolver) Set(prefix, uri string) {
	pr.prefixes[prefix] = uri
}

func (pr *PrefixResolver) All() map[string]string {
	return pr.prefixes
}
