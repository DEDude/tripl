package triple

type Node interface {
	isNode()
}

type IRI struct {
	Value string
}

func (IRI) isNode() {}
type Literal struct {
	Value    string
	Language string
	Datatype string
}

func (Literal) isNode() {}
type BlankNode struct {
	Value string
}

func (BlankNode) isNode() {}
type Triple struct {
	Subject   Node
	Predicate Node
	Object    Node
}
