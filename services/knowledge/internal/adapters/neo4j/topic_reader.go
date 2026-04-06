package neo4j

import (
	"context"
	"fmt"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/penkovgd/erudition-app/pkg/closer"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
)

// GetFiltered retrieves topics from Neo4j with optional filtering by associated concepts.
func (c *Client) GetFiltered(ctx context.Context, filters map[string][]string) ([]core.Topic, error) {
	session := c.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
	defer closer.CloseOrLogContext(ctx, c.log, session)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		query := "MATCH (t:Topic)"
		params := make(map[string]any)
		var conditions []string

		i := 0
		for _, concepts := range filters {
			if len(concepts) == 0 {
				continue
			}
			paramName := fmt.Sprintf("concepts_%d", i)
			params[paramName] = concepts

			cond := fmt.Sprintf(`EXISTS { MATCH (t)-[:CLASSIFIED_AS]->()-[:BROADER*0..]->(c:Concept) WHERE c.uri IN $%s }`, paramName)
			conditions = append(conditions, cond)
			i++
		}

		if len(conditions) > 0 {
			query += " WHERE " + strings.Join(conditions, " AND ")
		}

		query += `
			WITH t
			OPTIONAL MATCH (t)-[:CLASSIFIED_AS]->(c:Concept)
			RETURN 
				t.uri AS uri, 
				t.name AS name, 
				t.description AS description, 
				t.sparql AS sparql,
				collect(c.uri) AS concepts
		`

		c.log.Debug("Executing Cypher", "query", query, "params", params)

		res, err := tx.Run(ctx, query, params)
		if err != nil {
			return nil, err
		}

		var topics []core.Topic
		for res.Next(ctx) {
			rec := res.Record()
			uri, _ := rec.Get("uri")
			name, _ := rec.Get("name")
			desc, _ := rec.Get("description")
			sparql, _ := rec.Get("sparql")

			// Парсим массив концептов
			conceptsRaw, _ := rec.Get("concepts")
			var concepts []string

			if conceptsList, ok := conceptsRaw.([]any); ok {
				for _, c := range conceptsList {
					if c != nil {
						// Очищаем префикс "concept:", если хотим отдавать на фронт чистые ID
						// или оставляем как есть. В примере оставляем как есть.
						concepts = append(concepts, toString(c))
					}
				}
			}

			// Если список оказался пустым (Neo4j collect может вернуть [null]),
			// делаем пустой массив вместо nil, чтобы в JSON было []
			if len(concepts) == 0 {
				concepts = make([]string, 0)
			}

			topics = append(topics, core.Topic{
				Slug:        toString(uri),
				Name:        toString(name),
				Description: toString(desc),
				SPARQL:      toString(sparql),
				Concepts:    concepts,
			})
		}
		return topics, nil
	})

	if err != nil {
		return nil, fmt.Errorf("execute get filtered topics query: %w", err)
	}

	return result.([]core.Topic), nil
}
