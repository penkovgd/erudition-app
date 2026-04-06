package services

import (
	"context"
	"log/slog"
	"runtime"

	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
	"golang.org/x/sync/errgroup"
)

// TopicSyncer координирует синхронизацию топиков и SKOS классификации. Он загружает SKOS данные, сравнивает топики из провайдера и репозитория, и запускает ETL для обновленных топиков. Синхронизация выполняется параллельно с ограничением по количеству воркеров.
type TopicSyncer struct {
	log          *slog.Logger
	skosProvider core.SKOSProvider
	skosLoader   core.SKOSLoader
	topicProv    core.TopicProvider
	topicRepo    core.TopicRepository
	etl          *ETL
	workers      int
}

// NewTopicSyncer creates a new TopicSyncer with the given dependencies. It initializes the number of workers for parallel processing based on the number of CPU cores.
func NewTopicSyncer(
	log *slog.Logger,
	sp core.SKOSProvider, sl core.SKOSLoader,
	tp core.TopicProvider, tr core.TopicRepository,
	etl *ETL,
) (*TopicSyncer, error) {
	return &TopicSyncer{
		log: log, skosProvider: sp, skosLoader: sl,
		topicProv: tp, topicRepo: tr, etl: etl,
		workers: max(runtime.NumCPU(), 1),
	}, nil
}

// Sync координирует обе ветки загрузки
func (s *TopicSyncer) Sync(ctx context.Context) error {
	// ВЕТКА А: 1. Загрузка справочников (SKOS)
	s.log.Info("sync: starting SKOS synchronization")
	skosData, err := s.skosProvider.GetSKOS(ctx)
	if err != nil {
		return err
	}
	if err := s.skosLoader.LoadSKOS(ctx, skosData); err != nil {
		return err
	}
	s.log.Info("sync: SKOS loaded successfully")

	// ВЕТКА Б: 2. Загрузка топиков и их связей с концептами
	providerTopics, err := s.topicProv.GetAll(ctx)
	if err != nil {
		return err
	}

	repoTopics, err := s.topicRepo.GetAll(ctx)
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
		s.log.Info("sync: no topic updates needed")
		return nil
	}

	s.log.Info("sync: loading changed topics and executing ETL", "count", len(topicsToLoad))

	// Параллельный запуск ETL (как было у тебя)
	g, ctx := errgroup.WithContext(ctx)
	sem := make(chan struct{}, s.workers)

	for _, topic := range topicsToLoad {
		topic := topic
		sem <- struct{}{}

		g.Go(func() error {
			defer func() { <-sem }()

			s.log.Debug("sync: executing ETL for topic", "slug", topic.Slug)

			if err := s.etl.LoadTopic(ctx, topic); err != nil {
				s.log.Error("sync: ETL failed", "slug", topic.Slug, "error", err)
				return err
			}
			return nil
		})
	}

	return g.Wait()
}

func equalTopic(a, b core.Topic) bool {
	if a.Name != b.Name || a.Slug != b.Slug || a.Description != b.Description || a.SPARQL != b.SPARQL {
		return false
	}
	if len(a.Concepts) != len(b.Concepts) {
		return false
	}
	for i := range a.Concepts {
		if a.Concepts[i] != b.Concepts[i] {
			return false
		}
	}
	return true
}
