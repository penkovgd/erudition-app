// Package core contains the core data structures and types for the knowledge service
package core

// Quad represents a single RDF statement in the form of subject-predicate-object-graph
type Quad struct {
	Subject   URI
	Predicate URI
	Object    Object
	Graph     string
}

// URI represents node in knowledge graph i.e. http://www.wikidata.org/entity/Q12418
type URI string

// Object represents object in rdf quad. Can be uri or literal
type Object struct {
	Kind     string // "uri" or "literal"
	Value    string
	Datatype string
	Language string
}

// JSONLD is a type for json-ld rdf data
type JSONLD []byte

// Topic represents the topic for the quiz.
type Topic struct {
	Name        string   `json:"name"`
	Slug        string   `json:"slug"`
	Description string   `json:"description"`
	SPARQL      string   `json:"-"`
	Concepts    []string `json:"concepts"`
}

// ConceptScheme represents the facet (dimension) of the classification.
type ConceptScheme struct {
	ID        string
	PrefLabel string
}

// Concept represents a specific item inside a facet.
type Concept struct {
	ID        string
	PrefLabel string
	InScheme  string   // ID of parent ConceptScheme
	Broader   []string // ID of more general concepts (parents)
	Related   []string // ID of related concepts
}

// SKOSData container for all the concepts and concept schemes in the knowledge graph, used for quiz generation
type SKOSData struct {
	Schemes  []ConceptScheme
	Concepts []Concept
}

// SKOSTreeNode представляет узел в дереве фасета для фронтенда
type SKOSTreeNode struct {
	ID       string          `json:"id"`
	Label    string          `json:"label"`
	Children []*SKOSTreeNode `json:"children,omitempty"`
	Related  []string        `json:"related,omitempty"`
}

// SKOSFacet представляет корневой фасет со списком деревьев
type SKOSFacet struct {
	ID       string          `json:"id"`
	Label    string          `json:"label"`
	Concepts []*SKOSTreeNode `json:"concepts"`
}
