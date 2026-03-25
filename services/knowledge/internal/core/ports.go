package core

import (
	"context"
)

// Extractor extracts rdf triples from external sources like wikidata
type Extractor interface {
	Extract(ctx context.Context, topic Topic) ([]Quad, error)
}

// Transformer transforms serialized jsonld (i.e. normalization) and returns serialized jsonld
// type Transformer func(ctx context.Context, jsonld JSONLD) (JSONLD, error)

// Loader loads rdf triples into the knowledge graph
type Loader interface {
	LoadQuads(ctx context.Context, quads []Quad) error
}

// TopicProvider provides read access to the topics that should be extracted and loaded into the knowledge graph
type TopicProvider interface {
	GetAll(ctx context.Context) ([]Topic, error)
	// Get(ctx context.Context, slug string) (Topic, error)
}

// TopicRepository provides read and write access to the topics
type TopicRepository interface {
	GetAll(ctx context.Context) ([]Topic, error)
	// Add(ctx context.Context, topic Topic) error
}
