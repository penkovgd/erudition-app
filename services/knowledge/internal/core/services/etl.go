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
	log       *slog.Logger
	extractor core.Extractor
	// transformers []core.Transformer
	loader core.Loader
}

// NewETL creates a new ETL instance
func NewETL(log *slog.Logger, ext core.Extractor, loader core.Loader) (*ETL, error) {
	return &ETL{
		log:       log,
		extractor: ext,
		loader:    loader,
	}, nil
}

// LoadTopic performs the ETL process for a given topic: it extracts quads and loads them into the knowledge graph.
func (e *ETL) LoadTopic(ctx context.Context, topic core.Topic) error {
	quads, err := e.extractor.Extract(ctx, topic)
	if err != nil {
		return fmt.Errorf("extract quads: %w", err)
	}
	err = e.loader.LoadQuads(ctx, quads)
	if err != nil {
		return fmt.Errorf("load quads to knowledge graph: %w", err)
	}
	return nil
}
