package httpserver

import (
	"net/http"
	"strings"

	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
)

// handleGetSKOS godoc
// @Summary      Получить дерево классификации (SKOS)
// @Description  Возвращает фасеты и концепты, которые используются для фильтрации в UI.
// @Tags         skos
// @Produce      json
// @Success      200  {object}  SKOSResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/skos [get]
func (s *Server) handleGetSKOS(w http.ResponseWriter, r *http.Request) {
	trees, err := s.skos.GetTrees(r.Context())
	if err != nil {
		s.log.Error("failed to get SKOS trees", "error", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"facets": trees,
	})
}

// handleGetTopics godoc
// @Summary      Получение списка топиков
// @Description  Возвращает топики. Поддерживает фильтрацию: пересечение фасетов (AND) и объединение внутри фасета (OR через запятую).
// @Tags         topics
// @Produce      json
// @Param        scheme:geography query string false "Фильтр по географии (например: concept:russia,concept:europe)"
// @Param        scheme:time      query string false "Фильтр по времени (например: concept:20th_century)"
// @Param        scheme:domain    query string false "Фильтр по области знаний"
// @Success      200  {object}  TopicsResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/topics [get]
func (s *Server) handleGetTopics(w http.ResponseWriter, r *http.Request) {
	// Ожидаем формат: ?scheme:geography=concept:russia,concept:italy

	queryFilters := r.URL.Query()
	filters := make(map[string][]string)

	for key, values := range queryFilters {
		var concepts []string

		for _, val := range values {
			// Разбиваем строку по запятой
			parts := strings.SplitSeq(val, ",")
			for part := range parts {
				// Убираем пробелы по краям на случай, если клиент прислал "concept:russia, concept:italy"
				cleanPart := strings.TrimSpace(part)
				if cleanPart != "" {
					concepts = append(concepts, cleanPart)
				}
			}
		}

		if len(concepts) > 0 {
			filters[key] = concepts
		}
	}

	topics, err := s.topics.GetFiltered(r.Context(), filters)
	if err != nil {
		s.log.Error("failed to get filtered topics", "error", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Возвращаем пустой массив вместо null, если нет результатов
	if topics == nil {
		topics = make([]core.Topic, 0)
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"topics": topics,
	})
}

// handleSync godoc
// @Summary      Запуск ручной синхронизации
// @Description  Синхронизирует данные из YAML файлов с графовой БД Neo4j.
// @Tags         sync
// @Produce      json
// @Success      200  {object}  SyncResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /api/v1/sync [post]
func (s *Server) handleSync(w http.ResponseWriter, r *http.Request) {
	s.log.Info("Manual synchronization triggered via API")

	// Вызываем бизнес-логику синхронизации
	err := s.syncer.Sync(r.Context())
	if err != nil {
		s.log.Error("Synchronization failed", "error", err)
		writeError(w, http.StatusInternalServerError, "Failed to synchronize data: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "success",
		"message": "Synchronization completed successfully",
	})
}
