// Package core contains the core data structures and types for the knowledge service
package core

// JSONLD is a type for json-ld rdf data
type JSONLD []byte

// Topic represents a topic in the knowledge graph
type Topic struct {
	// name   string
	Sparql string
}
