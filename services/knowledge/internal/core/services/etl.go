// Package services provides ETL logic for the knowledge service.
package services

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
)

//  TODO распараллелить etl: каждый топик может загружаться параллельно

// ETL handles Extract-Transform-Load operations for topics.
type ETL struct {
	log          *slog.Logger
	extractor    core.Extractor
	transformers []core.Transformer
	loader       core.Loader
}

// New creates a new ETL instance with the provided logger, extractor, and loader.
func New(log *slog.Logger, ext core.Extractor, loader core.Loader) (*ETL, error) {
	return &ETL{
		log:       log,
		extractor: ext,
		loader:    loader,
	}, nil
}

// LoadTopic extracts data for a topic and loads it using the configured loader.
func (e *ETL) LoadTopic(ctx context.Context, topic core.Topic) error {
	jsonld, err := e.extractor.Extract(ctx, topic.Sparql)
	if err != nil {
		return fmt.Errorf("extract jsonld: %w", err)
	}

	for _, t := range e.transformers {
		jsonld, err = t(ctx, jsonld)
		if err != nil {
			return fmt.Errorf("transform jsonld: %w", err)
		}
	}

	err = e.loader.Load(ctx, jsonld)
	if err != nil {
		return fmt.Errorf("load jsonld: %w", err)
	}
	return nil
}

// func Normalize(ctx context.Context, jsonld core.JSONLD) (core.JSONLD, error) {
// 	var jsonldObj any
// 	if err := json.Unmarshal(jsonld, &jsonldObj); err != nil {
// 		return nil, fmt.Errorf("unmarshal json-ld into Go struct: %w", err)
// 	}

// 	proc := ld.NewJsonLdProcessor()
// 	options := ld.NewJsonLdOptions("") // Can add options

// 	rdf, err := proc.ToRDF(jsonldObj, options)
// 	if err != nil {
// 		return nil, fmt.Errorf("parse json-ld to rdf: %w", err)
// 	}
// 	ds, ok := rdf.(*ld.RDFDataset)
// 	if !ok {
// 		return nil, fmt.Errorf("expected *ld.RDFDataset, got %T", rdf)
// 	}
// }
