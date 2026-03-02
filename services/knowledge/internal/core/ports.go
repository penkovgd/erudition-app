package core

import (
	"context"
)

// Extractor extracts knowledge from external sources like wikidata
type Extractor interface {
	Extract(ctx context.Context, sparql string) (JSONLD, error)
}

// Transformer transforms serialized jsonld (i.e. normalization) and returns serialized jsonld
type Transformer func(ctx context.Context, jsonld JSONLD) (JSONLD, error)

// Loader loads knowledge into the knowledge graph
type Loader interface {
	Load(ctx context.Context, jsonld JSONLD) error
}
