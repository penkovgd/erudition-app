package core

import (
	"context"
)

// import "context"

// type Storage interface {
// 	Save(context.Context)
// }

//type Wikidata interface {
//	GetKnowledgeItems(ctx context.Context, sparql string) ([]models.KnowledgeItem, error)
//}

// Extractor extracts knowledge from external sources like wikidata
type Extractor interface {
	Extract(ctx context.Context, sparql string) (JSONLD, error)
}

// type Transformer interface {
// 	Transform(ctx context.Context, jsonld JSONLD) (JSONLD, error)
// }s

// Loader loads knowledge into the knowledge graph
type Loader interface {
	Load(ctx context.Context, jsonld JSONLD) error
}
