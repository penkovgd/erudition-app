// Package core contains the core data structures and types for the knowledge service
package core

// JSONLD is a type for json-ld rdf data
type JSONLD []byte

// Topic is a collection of knowledge and questions that are united by a single theme (example: Italian Renaissance painting)
type Topic struct {
	// Name   string
	// Desc string
	Sparql string
}
