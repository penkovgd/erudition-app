// Package main for the knowledge service.
// knowledge is responsible for collecting data from wikidata.org and storing it in a graph database (neo4j)
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/penkovgd/erudition-app/pkg/logger"
	neo4jAdapter "github.com/penkovgd/erudition-app/services/knowledge/internal/adapters/neo4j"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/adapters/wdqs"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/adapters/yamlload"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/config"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/core/services"
)

func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.LogLevel)
	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	ctx := context.Background()

	// Neo4j Client
	neo4j, err := neo4jAdapter.New(ctx, log, cfg.Neo4j.URI, cfg.Neo4j.Username, cfg.Neo4j.Password)
	if err != nil {
		return fmt.Errorf("create neo4j client: %w", err)
	}
	defer neo4j.Close(ctx)

	// Wikidata http client
	wdqs, err := wdqs.New(log)
	if err != nil {
		return fmt.Errorf("create wikidata http client: %w", err)
	}

	// ETL
	etl, err := services.NewETL(log, wdqs, neo4j)
	if err != nil {
		return fmt.Errorf("create ETL service: %w", err)
	}

	// Topic Provider
	topicProvider := yamlload.NewTopicYAMLLoader(log)

	// Topic Syncer
	syncer, err := services.NewTopicSyncer(log, topicProvider, neo4j, etl)
	if err != nil {
		return fmt.Errorf("create topic syncer: %w", err)
	}

	// Sync topics
	if err := syncer.Sync(ctx); err != nil {
		return fmt.Errorf("sync topics: %w", err)
	}

	log.Info("topics synced successfully")

	return nil

	// gRPC server
	// listener, err := net.Listen("tcp", cfg.GRPCAddress())
	// if err != nil {
	// 	return fmt.Errorf("listen: %w", err)
	// }

	// s := grpc.NewServer()
	// reflection.Register(s)

	// ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	// defer stop()

	// go func() {
	// 	<-ctx.Done()
	// 	log.Debug("trying to shutdown gracefully...")
	// 	timer := time.AfterFunc(5*time.Second, func() {
	// 		log.Warn("server couldn't stop gracefully in time. doing force stop")
	// 		s.Stop()
	// 	})
	// 	defer timer.Stop()
	// 	s.GracefulStop()
	// 	log.Debug("server stopped gracefully")
	// }()

	// if err := s.Serve(listener); err != nil {
	// 	return fmt.Errorf("serve: %w", err)
	// }
}
