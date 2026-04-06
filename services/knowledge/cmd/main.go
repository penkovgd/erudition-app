// Package main for the knowledge service.
// knowledge is responsible for collecting data from wikidata.org and storing it in a graph database (neo4j)
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/penkovgd/erudition-app/pkg/logger"
	_ "github.com/penkovgd/erudition-app/services/knowledge/docs"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/adapters/httpserver"
	neo4jAdapter "github.com/penkovgd/erudition-app/services/knowledge/internal/adapters/neo4j"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/adapters/wdqs"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/adapters/yamlload"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/config"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/core/services"
)

// @title           Knowledge API
// @version         1.0
// @description     API микросервиса knowledge для работы с графом знаний (деревья SKOS и топики).
// @BasePath        /
func main() {
	cfg := config.MustLoad()
	log := logger.New(cfg.LogLevel)
	if err := run(cfg, log); err != nil {
		log.Error("service stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting knowledge service")
	log.Debug("debug messages are enabled")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Neo4j Client
	neo4j, err := neo4jAdapter.New(ctx, log, cfg.Neo4j.URI, cfg.Neo4j.Username, cfg.Neo4j.Password)
	if err != nil {
		return fmt.Errorf("create neo4j client: %w", err)
	}
	defer neo4j.Close(context.Background())

	// Wikidata http client
	wdqsClient, err := wdqs.New(log)
	if err != nil {
		return fmt.Errorf("create wikidata http client: %w", err)
	}

	// ETL
	etl, err := services.NewETL(log, wdqsClient, neo4j)
	if err != nil {
		return fmt.Errorf("create ETL service: %w", err)
	}

	skosProvider := yamlload.NewSKOSYAMLLoader(log)
	topicProvider := yamlload.NewTopicYAMLLoader(log)

	// Topic Syncer
	syncer, err := services.NewTopicSyncer(
		log,
		skosProvider, neo4j,
		topicProvider, neo4j,
		etl,
	)
	if err != nil {
		return fmt.Errorf("create topic syncer: %w", err)
	}

	httpServer := httpserver.NewServer(log, ":8081", neo4j, neo4j, syncer)
	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- httpServer.Start()
	}()

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		log.Info("shutdown signal received")

		// Контекст с таймаутом для плавного завершения
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := httpServer.Stop(shutdownCtx); err != nil {
			log.Error("graceful shutdown failed", "error", err)
			return err
		}

		log.Info("server stopped gracefully")
	}

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
