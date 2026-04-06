// Package httpserver implements the HTTP server for the Knowledge service, exposing RESTful APIs to interact with the SKOS data and topics. It handles incoming HTTP requests, interacts with the core services to fetch data, and returns JSON responses. The server also includes basic logging and graceful shutdown capabilities.
package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	_ "github.com/penkovgd/erudition-app/services/knowledge/docs" // импорт для генерации Swagger документации
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
)

// Server represents the HTTP server for the Knowledge service. It holds references to the core services for SKOS and topic management, as well as a logger.
type Server struct {
	server *http.Server
	log    *slog.Logger
	skos   core.SKOSReader
	topics core.TopicReader
	syncer core.Syncer
}

// NewServer creates a new instance of the Server with the provided logger, address, and core service dependencies. It sets up the HTTP routes and middleware.
func NewServer(log *slog.Logger, address string, skos core.SKOSReader, topics core.TopicReader, syncer core.Syncer) *Server {
	s := &Server{
		log:    log,
		skos:   skos,
		topics: topics,
		syncer: syncer,
	}

	// Настраиваем роутер (Mux)
	mux := http.NewServeMux()

	// API v1
	mux.HandleFunc("GET /api/v1/skos", s.handleGetSKOS)
	mux.HandleFunc("GET /api/v1/topics", s.handleGetTopics)
	mux.HandleFunc("POST /api/v1/sync", s.handleSync)

	// Swagger UI
	// Добавляем обработчик для документации
	mux.Handle("/docs/", httpSwagger.Handler(
		// TODO подправить адрес на конфигуриемый
		httpSwagger.URL("http://localhost"+address+"/docs/doc.json"),
	))

	var handler http.Handler = mux
	handler = s.corsMiddleware(handler)
	handler = s.loggingMiddleware(handler)

	s.server = &http.Server{
		Addr:         address,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	return s
}

// Start запускает HTTP сервер и начинает слушать входящие запросы. Если сервер останавливается с ошибкой, которая не является ErrServerClosed, возвращает эту ошибку.
func (s *Server) Start() error {
	s.log.Info("Starting HTTP server", "address", s.server.Addr)
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// Stop останавливает HTTP сервер gracefully, позволяя завершить обработку текущих запросов. Если возникает ошибка при остановке, возвращает эту ошибку.
func (s *Server) Stop(ctx context.Context) error {
	s.log.Info("Stopping HTTP server gracefully")
	return s.server.Shutdown(ctx)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode JSON response", http.StatusInternalServerError)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// loggingMiddleware логирует каждый входящий запрос
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		s.log.Info("HTTP Request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", time.Since(start).String(),
		)
	})
}

// corsMiddleware добавляет заголовки CORS и обрабатывает preflight-запросы (OPTIONS)
func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Временно разрешаем запросы с любых доменов
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// Разрешаем стандартные методы, включая OPTIONS для preflight
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		// Разрешаем стандартные заголовки, которые обычно шлет клиент (Content-Type обязателен для JSON)
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		// Если это preflight-запрос от браузера, просто возвращаем 200 OK
		// и прерываем цепочку (не передаем в роутер)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Передаем управление следующему обработчику в цепочке
		next.ServeHTTP(w, r)
	})
}
