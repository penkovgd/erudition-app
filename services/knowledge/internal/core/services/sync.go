package services

import (
	"context"
	"errors"
	"runtime"

	"log/slog"

	"golang.org/x/sync/errgroup"

	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
)

// TopicLoader defines the interface for loading a topic into the knowledge graph.
type TopicLoader interface {
	LoadTopic(ctx context.Context, topic core.Topic) error
}

// TopicSyncer synchronizes topics from a provider to the knowledge graph via ETL, comparing against repository state.
type TopicSyncer struct {
	log      *slog.Logger
	provider core.TopicProvider
	repo     core.TopicRepository
	loader   TopicLoader
	workers  int
}

// NewTopicSyncer creates a new TopicSyncer.
func NewTopicSyncer(log *slog.Logger, provider core.TopicProvider, repo core.TopicRepository, loader TopicLoader) (*TopicSyncer, error) {
	if log == nil {
		return nil, errors.New("log is required")
	}
	if provider == nil {
		return nil, errors.New("provider is required")
	}
	if repo == nil {
		return nil, errors.New("repo is required")
	}
	if loader == nil {
		return nil, errors.New("loader is required")
	}

	workers := max(runtime.NumCPU(), 1)

	return &TopicSyncer{
		log:      log,
		provider: provider,
		repo:     repo,
		loader:   loader,
		workers:  workers,
	}, nil
}

// Sync loads modified topics in parallel (by slug). If topic is absent in repository or differs in any field, it is reloaded.
func (s *TopicSyncer) Sync(ctx context.Context) error {
	providerTopics, err := s.provider.GetAll(ctx)
	if err != nil {
		return err
	}

	repoTopics, err := s.repo.GetAll(ctx)
	if err != nil {
		return err
	}

	repoBySlug := make(map[string]core.Topic, len(repoTopics))
	for _, t := range repoTopics {
		repoBySlug[t.Slug] = t
	}

	var topicsToLoad []core.Topic
	for _, p := range providerTopics {
		r, ok := repoBySlug[p.Slug]
		if !ok || !equalTopic(p, r) {
			topicsToLoad = append(topicsToLoad, p)
		}
	}

	if len(topicsToLoad) == 0 {
		s.log.Debug("topic sync: no updates needed")
		return nil
	}

	s.log.Info("topic sync: loading changed topics", "count", len(topicsToLoad))

	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, s.workers)

	for _, topic := range topicsToLoad {
		topic := topic
		sem <- struct{}{}

		g.Go(func() error {
			defer func() { <-sem }()

			s.log.Debug("topic sync: loading topic", "slug", topic.Slug)
			if err := s.loader.LoadTopic(ctx, topic); err != nil {
				s.log.Error("topic sync: load failed", "slug", topic.Slug, "error", err)
				return err
			}
			s.log.Info("topic sync: loaded", "slug", topic.Slug)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func equalTopic(a, b core.Topic) bool {
	return a.Name == b.Name && a.Slug == b.Slug && a.Description == b.Description && a.SPARQL == b.SPARQL
}
