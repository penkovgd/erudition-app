package yamlload

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
	"gopkg.in/yaml.v3"
)

// SKOSYAMLLoader is a file-based loader for SKOS classification data
type SKOSYAMLLoader struct {
	log     *slog.Logger
	dataDir string
}

// NewSKOSYAMLLoader creates a new SKOSYAMLLoader
func NewSKOSYAMLLoader(log *slog.Logger) *SKOSYAMLLoader {
	return &SKOSYAMLLoader{
		log:     log,
		dataDir: findDataDir(),
	}
}

type skosYAML struct {
	Schemes  map[string]schemeYAML  `yaml:"schemes"`
	Concepts map[string]conceptYAML `yaml:"concepts"`
}

type schemeYAML struct {
	PrefLabel string `yaml:"prefLabel"`
}

type conceptYAML struct {
	PrefLabel string   `yaml:"prefLabel"`
	InScheme  string   `yaml:"inScheme"`
	Broader   []string `yaml:"broader"`
	Related   []string `yaml:"related"`
}

// GetSKOS loads SKOS data from the skos.yaml file and returns it as SKOSData
func (l *SKOSYAMLLoader) GetSKOS(_ context.Context) (core.SKOSData, error) {
	skosPath := filepath.Join(l.dataDir, "skos.yaml")
	cleanSKOSPath := filepath.Clean(skosPath)
	cleanDataDir := filepath.Clean(l.dataDir)

	if !strings.HasPrefix(cleanSKOSPath, cleanDataDir+string(os.PathSeparator)) {
		return core.SKOSData{}, fmt.Errorf("invalid skos path: %q", skosPath)
	}

	data, err := os.ReadFile(cleanSKOSPath)
	if err != nil {
		return core.SKOSData{}, fmt.Errorf("read skos file: %w", err)
	}

	var parsed skosYAML
	if err := yaml.Unmarshal(data, &parsed); err != nil {
		return core.SKOSData{}, fmt.Errorf("unmarshal skos.yaml: %w", err)
	}

	var result core.SKOSData

	for id, s := range parsed.Schemes {
		result.Schemes = append(result.Schemes, core.ConceptScheme{
			ID:        id,
			PrefLabel: s.PrefLabel,
		})
	}

	for id, c := range parsed.Concepts {
		result.Concepts = append(result.Concepts, core.Concept{
			ID:        id,
			PrefLabel: c.PrefLabel,
			InScheme:  c.InScheme,
			Broader:   c.Broader,
			Related:   c.Related,
		})
	}

	return result, nil
}
