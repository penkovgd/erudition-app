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

// Topic represents named subgraph to be extracted and loaded into the knowledge graph. Contains metadata and a SPARQL query to fetch the data.
type Topic struct {
	Name        string
	Slug        string
	Description string
	SPARQL      string
}
