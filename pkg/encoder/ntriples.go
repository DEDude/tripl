package encoder

import (
	"fmt"
	"github.com/DeDude/tripl/pkg/triple"
)

func EncodeNTriple(t triple.Triple) string {
	subject := formatNode(t.Subject)
	predicate := formatNode(t.Predicate)
	object := formatNode(t.Object)
	return fmt.Sprintf("%s %s %s .", subject, predicate, object)
}
