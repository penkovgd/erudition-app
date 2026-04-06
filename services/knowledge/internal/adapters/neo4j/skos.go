package neo4j

import (
	"context"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/penkovgd/erudition-app/pkg/closer"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
)

// LoadSKOS cleans old SKOS data and loads new one. It creates ConceptScheme and Concept nodes, links concepts to schemes and builds hierarchy with broader relation. All operations are done in a single transaction.
func (c *Client) LoadSKOS(ctx context.Context, data core.SKOSData) error {
	session := c.NewSession(ctx, neo4j.SessionConfig{})
	defer closer.CloseOrLogContext(ctx, c.log, session)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// 1. Создаем схемы (Фасеты)
		for _, scheme := range data.Schemes {
			query := `MERGE (s:ConceptScheme {uri: $id}) SET s.prefLabel = $label`
			if _, err := tx.Run(ctx, query, map[string]any{"id": "scheme:" + scheme.ID, "label": scheme.PrefLabel}); err != nil {
				return nil, err
			}
		}

		// 2. Создаем концепты и привязываем к схемам
		for _, concept := range data.Concepts {
			query := `
                MERGE (c:Concept {uri: $id}) 
                SET c.prefLabel = $label
                WITH c
                MATCH (s:ConceptScheme {uri: $schemeId})
                MERGE (c)-[:IN_SCHEME]->(s)
            `
			params := map[string]any{
				"id":       "concept:" + concept.ID,
				"label":    concept.PrefLabel,
				"schemeId": "scheme:" + concept.InScheme,
			}
			if _, err := tx.Run(ctx, query, params); err != nil {
				return nil, err
			}
		}

		// 3. Строим иерархию (broader)
		// Делаем вторым проходом, чтобы гарантировать, что все концепты уже созданы
		for _, concept := range data.Concepts {
			for _, broaderID := range concept.Broader {
				query := `
                    MATCH (c:Concept {uri: $id})
                    MATCH (b:Concept {uri: $broaderId})
                    MERGE (c)-[:BROADER]->(b)
                `
				if _, err := tx.Run(ctx, query, map[string]any{
					"id":        "concept:" + concept.ID,
					"broaderId": "concept:" + broaderID,
				}); err != nil {
					return nil, err
				}
			}
		}
		// 4. Строим связи RELATED
		for _, concept := range data.Concepts {
			for _, relatedID := range concept.Related {
				query := `
					MATCH (c:Concept {uri: $id})
					MATCH (r:Concept {uri: $relId})
					MERGE (c)-[:RELATED]-(r)
				`
				if _, err := tx.Run(ctx, query, map[string]any{
					"id":    "concept:" + concept.ID,
					"relId": "concept:" + relatedID,
				}); err != nil {
					return nil, err
				}
			}
		}
		return nil, nil
	})
	return err
}

// LoadTopic loads a topic with its subjects. It creates/updates a Topic node and links it to Concept nodes via SUBJECT relation. All operations are done in a single transaction. If the topic already exists, old SUBJECT relations are deleted and replaced with new ones.
func (c *Client) LoadTopic(ctx context.Context, topic core.Topic) error {
	session := c.NewSession(ctx, neo4j.SessionConfig{})
	defer closer.CloseOrLogContext(ctx, c.log, session)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		uri := topic.Slug
		// Создаем сам узел (как было)
		_, err := tx.Run(ctx,
			"MERGE (t:Topic {uri: $uri}) SET t.name = $name, t.description = $description, t.sparql = $sparql",
			map[string]any{"uri": uri, "name": topic.Name, "description": topic.Description, "sparql": topic.SPARQL})
		if err != nil {
			return nil, err
		}

		// Очищаем старые связи с классификацией (если топик обновился)
		if _, err := tx.Run(ctx, "MATCH (t:Topic {uri: $uri})-[r:CLASSIFIED_AS]->() DELETE r", map[string]any{"uri": uri}); err != nil {
			return nil, err
		}

		// Добавляем новые связи
		for _, conceptID := range topic.Concepts {
			query := `
                MATCH (t:Topic {uri: $uri})
                MATCH (c:Concept {uri: $conceptId})
                MERGE (t)-[:CLASSIFIED_AS]->(c)
            `
			if _, err := tx.Run(ctx, query, map[string]any{
				"uri":       uri,
				"conceptId": "concept:" + conceptID, // добавляем префикс, как при создании концепта
			}); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}
