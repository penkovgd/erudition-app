// Package neo4j is a neo4j (graph db) adapter for managing knowledge graph
package neo4j

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/penkovgd/erudition-app/pkg/closer"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
)

// Client uses the neo4j driver to send queries
type Client struct {
	neo4j.Driver
	log *slog.Logger
}

// New creates a neo4j client with a given credentials and verifies connection
func New(ctx context.Context, log *slog.Logger, uri, username, password string) (*Client, error) {
	driver, err := neo4j.NewDriver(
		uri,
		neo4j.BasicAuth(username, password, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("create neo4j driver: %w", err)
	}

	if err = driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("verify connection to neo4j: %w", err)
	}

	if err := ensureConstraints(ctx, driver); err != nil {
		return nil, fmt.Errorf("ensure constraints/indexes: %w", err)
	}

	log.Debug("connected to neo4j")
	return &Client{Driver: driver, log: log}, nil
}

// Close closes neo4j driver. If he gives an error, logs
func (c *Client) Close(ctx context.Context) {
	closer.CloseOrLogContext(ctx, c.log, c.Driver)
}

// ensureConstraints creates indexes and constraints for the new schema.
func ensureConstraints(ctx context.Context, driver neo4j.Driver) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer closer.CloseOrPanicContext(ctx, session)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Resource: unique URI
		if _, err := tx.Run(ctx,
			"CREATE CONSTRAINT IF NOT EXISTS FOR (r:Resource) REQUIRE r.uri IS UNIQUE", nil); err != nil {
			return nil, err
		}
		// Topic: unique URI
		if _, err := tx.Run(ctx,
			"CREATE CONSTRAINT IF NOT EXISTS FOR (t:Topic) REQUIRE t.uri IS UNIQUE", nil); err != nil {
			return nil, err
		}
		// Literal: unique combination of value, datatype, language
		if _, err := tx.Run(ctx,
			"CREATE CONSTRAINT IF NOT EXISTS FOR (l:Literal) REQUIRE (l.value, l.datatype, l.language) IS UNIQUE", nil); err != nil {
			return nil, err
		}

		// Indexes for faster lookups
		if _, err := tx.Run(ctx, "CREATE INDEX IF NOT EXISTS FOR (r:Resource) ON (r.label)", nil); err != nil {
			return nil, err
		}
		if _, err := tx.Run(ctx, "CREATE INDEX IF NOT EXISTS FOR (r:Resource) ON (r.topics)", nil); err != nil {
			return nil, err
		}
		if _, err := tx.Run(ctx, "CREATE INDEX IF NOT EXISTS FOR (l:Literal) ON (l.value)", nil); err != nil {
			return nil, err
		}
		if _, err := tx.Run(ctx, "CREATE INDEX IF NOT EXISTS FOR (l:Literal) ON (l.topics)", nil); err != nil {
			return nil, err
		}
		// Index for Relationship properties
		if _, err := tx.Run(ctx, "CREATE INDEX IF NOT EXISTS FOR ()-[r:REL]-() ON (r.uri)", nil); err != nil {
			return nil, err
		}

		return nil, nil
	})
	return err
}

// Добавлен параметр description
func createResourceNode(ctx context.Context, tx neo4j.ManagedTransaction, uri, label, description, topic string) error {
	query := `
        MERGE (r:Resource {uri: $uri})
        ON CREATE SET 
            r.label = $label,
            r.description = $description,
            r.topics = CASE WHEN $topic <> "" THEN [$topic] ELSE [] END
        ON MATCH SET 
            r.label = CASE WHEN $label <> "" THEN $label ELSE r.label END,
            r.description = CASE WHEN $description <> "" THEN $description ELSE r.description END,
            r.topics = CASE WHEN $topic <> "" AND NOT $topic IN r.topics THEN r.topics + $topic ELSE r.topics END
    `
	_, err := tx.Run(ctx, query, map[string]any{
		"uri":         uri,
		"label":       label,
		"description": description,
		"topic":       topic,
	})
	return err
}

func createLiteralNode(ctx context.Context, tx neo4j.ManagedTransaction, obj core.Object, topic string) error {
	query := `
        MERGE (l:Literal {value: $value, datatype: $datatype, language: $language})
        ON CREATE SET 
            l.topics = CASE WHEN $topic <> "" THEN [$topic] ELSE [] END
        ON MATCH SET 
            l.topics = CASE WHEN $topic <> "" AND NOT $topic IN l.topics THEN l.topics + $topic ELSE l.topics END
    `
	_, err := tx.Run(ctx, query, map[string]any{
		"value":    obj.Value,
		"datatype": obj.Datatype,
		"language": obj.Language,
		"topic":    topic,
	})
	return err
}

// Добавлен параметр predDesc
func createRelationship(ctx context.Context, tx neo4j.ManagedTransaction, subjURI, predURI, predLabel, predDesc string, obj core.Object, topic string) error {
	params := map[string]any{
		"subjURI":   subjURI,
		"predURI":   predURI,
		"predLabel": predLabel,
		"predDesc":  predDesc,
		"topic":     topic,
	}

	var query string
	if obj.Kind == "uri" {
		query = `
            MATCH (s:Resource {uri: $subjURI})
            MATCH (o:Resource {uri: $objURI})
            MERGE (s)-[rel:REL {uri: $predURI}]->(o)
            ON CREATE SET 
                rel.label = $predLabel,
                rel.description = $predDesc,
                rel.topics = CASE WHEN $topic <> "" THEN [$topic] ELSE [] END
            ON MATCH SET 
                rel.label = CASE WHEN $predLabel <> "" THEN $predLabel ELSE rel.label END,
                rel.description = CASE WHEN $predDesc <> "" THEN $predDesc ELSE rel.description END,
                rel.topics = CASE WHEN $topic <> "" AND NOT $topic IN rel.topics THEN rel.topics + $topic ELSE rel.topics END
        `
		params["objURI"] = obj.Value
	} else {
		query = `
            MATCH (s:Resource {uri: $subjURI})
            MATCH (o:Literal {value: $value, datatype: $datatype, language: $language})
            MERGE (s)-[rel:REL {uri: $predURI}]->(o)
            ON CREATE SET 
                rel.label = $predLabel,
                rel.description = $predDesc,
                rel.topics = CASE WHEN $topic <> "" THEN [$topic] ELSE [] END
            ON MATCH SET 
                rel.label = CASE WHEN $predLabel <> "" THEN $predLabel ELSE rel.label END,
                rel.description = CASE WHEN $predDesc <> "" THEN $predDesc ELSE rel.description END,
                rel.topics = CASE WHEN $topic <> "" AND NOT $topic IN rel.topics THEN rel.topics + $topic ELSE rel.topics END
        `
		params["value"] = obj.Value
		params["datatype"] = obj.Datatype
		params["language"] = obj.Language
	}

	_, err := tx.Run(ctx, query, params)
	return err
}

// GetAll retrieves all topics from the database.
func (c *Client) GetAll(ctx context.Context) ([]core.Topic, error) {
	session := c.NewSession(ctx, neo4j.SessionConfig{})
	defer closer.CloseOrLogContext(ctx, c.log, session)

	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(ctx, `
            MATCH (t:Topic) 
            RETURN t.uri AS uri, t.name AS name, t.description AS description, t.sparql AS sparql
        `, nil)
		if err != nil {
			return nil, fmt.Errorf("execute query: %w", err)
		}

		var topics []core.Topic
		for result.Next(ctx) {
			record := result.Record()

			uri, _ := record.Get("uri")
			name, _ := record.Get("name")
			description, _ := record.Get("description")
			sparql, _ := record.Get("sparql")

			topic := core.Topic{
				Slug:        toString(uri),
				Name:        toString(name),
				Description: toString(description),
				SPARQL:      toString(sparql),
			}
			topics = append(topics, topic)
		}

		if err := result.Err(); err != nil {
			return nil, fmt.Errorf("iteration error: %w", err)
		}

		return topics, nil
	})

	if err != nil {
		return nil, err
	}

	topics, ok := result.([]core.Topic)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
	return topics, nil
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// LoadQuads imports RDF quads into the graph with direct relationships.
func (c *Client) LoadQuads(ctx context.Context, quads []core.Quad) error {
	session := c.NewSession(ctx, neo4j.SessionConfig{})
	defer closer.CloseOrLogContext(ctx, c.log, session)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// Добавлен словарь предикатов, которые являются метками
		labelPredicates := map[string]bool{
			"http://www.w3.org/2000/01/rdf-schema#label":    true,
			"http://www.w3.org/2004/02/skos/core#prefLabel": true,
		}
		// Добавлен словарь предикатов, которые являются описаниями
		descriptionPredicates := map[string]bool{
			"http://schema.org/description": true,
		}

		// 1. Собираем все label и description для Resource и Predicate
		resourceLabels := make(map[string]string)
		resourceDescriptions := make(map[string]string)

		for _, q := range quads {
			if q.Object.Kind != "literal" {
				continue
			}
			predURI := string(q.Predicate)
			subjURI := string(q.Subject)

			if labelPredicates[predURI] {
				if _, exists := resourceLabels[subjURI]; !exists {
					resourceLabels[subjURI] = q.Object.Value
				}
			} else if descriptionPredicates[predURI] {
				if _, exists := resourceDescriptions[subjURI]; !exists {
					resourceDescriptions[subjURI] = q.Object.Value
				}
			}
		}

		// 2. Основной цикл загрузки узлов и связей
		for _, q := range quads {
			subjURI := string(q.Subject)
			subjLabel := resourceLabels[subjURI]
			subjDesc := resourceDescriptions[subjURI]
			topicURI := q.Graph
			predURI := string(q.Predicate)

			// Создаем субъект (теперь и с label, и с description)
			if err := createResourceNode(ctx, tx, subjURI, subjLabel, subjDesc, topicURI); err != nil {
				return nil, fmt.Errorf("subject node %s: %w", subjURI, err)
			}

			// Если квад задает label или description узлу, то свойство мы уже сохранили, хвост не нужен
			if labelPredicates[predURI] || descriptionPredicates[predURI] {
				continue
			}

			// Создаем объект
			if q.Object.Kind == "uri" {
				objURI := q.Object.Value
				objLabel := resourceLabels[objURI]
				objDesc := resourceDescriptions[objURI]
				if err := createResourceNode(ctx, tx, objURI, objLabel, objDesc, topicURI); err != nil {
					return nil, fmt.Errorf("object resource %s: %w", objURI, err)
				}
			} else {
				// Обычный строковый литерал
				if err := createLiteralNode(ctx, tx, q.Object, topicURI); err != nil {
					return nil, fmt.Errorf("literal node: %w", err)
				}
			}

			// Создаем связь (Предикат) с его возможным label и description
			predLabel := resourceLabels[predURI]
			predDesc := resourceDescriptions[predURI]

			if err := createRelationship(ctx, tx, subjURI, predURI, predLabel, predDesc, q.Object, topicURI); err != nil {
				return nil, fmt.Errorf("relationship creation error: %w", err)
			}
		}
		return nil, nil
	})

	return err
}
