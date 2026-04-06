// Package yamlload provides an adapter to load quiz data from YAML files.
package yamlload

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
	"gopkg.in/yaml.v3"
)

// TopicYAMLLoader is a file-based loader for quiz topics
type TopicYAMLLoader struct {
	log     *slog.Logger
	dataDir string
}

// NewTopicYAMLLoader creates a new TopicYAMLLoader
func NewTopicYAMLLoader(log *slog.Logger) *TopicYAMLLoader {
	return &TopicYAMLLoader{
		log:     log,
		dataDir: findDataDir(),
	}
}

func findDataDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "data"
	}

	dataDir := filepath.Join(filepath.Dir(filename), "..", "..", "..", "data")
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return "data"
	}
	return absDataDir
}

type topicYAML struct {
	Name        string   `yaml:"name"`
	Slug        string   `yaml:"slug"`
	Description string   `yaml:"description"`
	SPARQL      string   `yaml:"sparql"`
	Concepts    []string `yaml:"concepts"`
}

// GetAll loads all topics from the data/topics directory
func (r *TopicYAMLLoader) GetAll(_ context.Context) ([]core.Topic, error) {
	log.Printf("Loading topics from %s", r.dataDir)

	topicsDir := filepath.Join(r.dataDir, "topics")

	entries, err := os.ReadDir(topicsDir)
	if err != nil {
		return nil, fmt.Errorf("read topics dir: %w", err)
	}

	topics := make([]core.Topic, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		topicData, err := os.ReadFile(filepath.Clean(filepath.Join(topicsDir, entry.Name())))
		if err != nil {
			return nil, fmt.Errorf("open topic file %s: %w", entry.Name(), err)
		}
		var topic topicYAML
		if err := yaml.Unmarshal(topicData, &topic); err != nil {
			return nil, fmt.Errorf("unmarshal %s: %w", entry.Name(), err)
		}

		topics = append(topics, core.Topic{
			Name:        topic.Name,
			Slug:        topic.Slug,
			Description: topic.Description,
			SPARQL:      topic.SPARQL,
			Concepts:    topic.Concepts,
		})
	}

	return topics, nil
}
