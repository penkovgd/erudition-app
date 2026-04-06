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
	LoadTopic(ctx context.Context, topic Topic) error
}

// TopicRepository provides read and write access to the topics
type TopicRepository interface {
	GetAll(ctx context.Context) ([]Topic, error)
	// Add(ctx context.Context, topic Topic) error
}

// TopicProvider reads Topics from e.g. file
type TopicProvider interface {
	GetAll(ctx context.Context) ([]Topic, error)
}

// SKOSProvider reads SKOS reads SKOS data from e.g. file
type SKOSProvider interface {
	GetSKOS(ctx context.Context) (SKOSData, error)
}

// SKOSLoader saves SKOS classification into knowledge graph
type SKOSLoader interface {
	LoadSKOS(ctx context.Context, data SKOSData) error
}

type Syncer interface {
	Sync(ctx context.Context) error
}

// SKOSReader читает деревья фасетов из графа
type SKOSReader interface {
	GetTrees(ctx context.Context) ([]SKOSFacet, error)
}

// TopicReader читает топики с учетом фильтрации
type TopicReader interface {
	// filters - это мапа вида map[schemeID][]conceptID
	// Например: map["scheme:geography"]: ["concept:russia", "concept:italy"]
	GetFiltered(ctx context.Context, filters map[string][]string) ([]Topic, error)
}
