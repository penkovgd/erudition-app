// Package neo4j is a neo4j (graph db) adapter for managing knowledge graph
package neo4j

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

// Close closes neo4j driver. If he gives an error, panics
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
		// Predicate: unique URI
		if _, err := tx.Run(ctx,
			"CREATE CONSTRAINT IF NOT EXISTS FOR (p:Predicate) REQUIRE p.uri IS UNIQUE", nil); err != nil {
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
		// Triplet: unique ID
		if _, err := tx.Run(ctx,
			"CREATE CONSTRAINT IF NOT EXISTS FOR (t:Triplet) REQUIRE t.id IS UNIQUE", nil); err != nil {
			return nil, err
		}
		// Indexes for faster lookups
		if _, err := tx.Run(ctx,
			"CREATE INDEX IF NOT EXISTS FOR (l:Literal) ON (l.value)", nil); err != nil {
			return nil, err
		}
		if _, err := tx.Run(ctx,
			"CREATE INDEX IF NOT EXISTS FOR (l:Literal) ON (l.datatype)", nil); err != nil {
			return nil, err
		}
		if _, err := tx.Run(ctx,
			"CREATE INDEX IF NOT EXISTS FOR (l:Literal) ON (l.language)", nil); err != nil {
			return nil, err
		}

		if _, err := tx.Run(ctx,
			"CREATE INDEX IF NOT EXISTS FOR (r:Resource) ON (r.label)", nil); err != nil {
			return nil, err
		}
		if _, err := tx.Run(ctx,
			"CREATE INDEX IF NOT EXISTS FOR (p:Predicate) ON (p.label)", nil); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

// generateTripletID creates a unique identifier for a triplet.
// It uses a SHA‑256 hash of the concatenation:
//
//	subject_uri + "|" + predicate_uri + "|" + object_key
//
// where object_key is either the URI (for entities) or a string representing the literal.
func generateTripletID(subjURI, predURI string, obj core.Object) string {
	var objKey string
	if obj.Kind == "uri" {
		objKey = "uri:" + obj.Value
	} else {
		// literal: value|datatype|language
		objKey = fmt.Sprintf("lit:%s|%s|%s", obj.Value, obj.Datatype, obj.Language)
	}
	raw := subjURI + "|" + predURI + "|" + objKey
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:])
}

func createResourceNode(ctx context.Context, tx neo4j.ManagedTransaction, uri, label string) error {
	params := map[string]any{"uri": uri, "label": label}
	_, err := tx.Run(ctx,
		"MERGE (r:Resource {uri: $uri}) ON CREATE SET r.label = $label ON MATCH SET r.label = $label",
		params)
	return err
}

func createPredicateNode(ctx context.Context, tx neo4j.ManagedTransaction, uri, label string) error {
	params := map[string]any{"uri": uri, "label": label}
	_, err := tx.Run(ctx,
		"MERGE (p:Predicate {uri: $uri}) ON CREATE SET p.label = $label ON MATCH SET p.label = $label",
		params)
	return err
}

func createLiteralNode(ctx context.Context, tx neo4j.ManagedTransaction, obj core.Object) error {
	_, err := tx.Run(ctx,
		`MERGE (l:Literal {value: $value, datatype: $datatype, language: $language})`,
		map[string]any{
			"value":    obj.Value,
			"datatype": obj.Datatype,
			"language": obj.Language,
		})
	return err
}

func createTopicNode(ctx context.Context, tx neo4j.ManagedTransaction, uri string) error {
	_, err := tx.Run(ctx,
		"MERGE (t:Topic {uri: $uri})",
		map[string]any{"uri": uri})
	return err
}

func createTripletNode(ctx context.Context, tx neo4j.ManagedTransaction, tripletID string) error {
	_, err := tx.Run(ctx,
		"MERGE (tr:Triplet {id: $id})",
		map[string]any{"id": tripletID})
	return err
}

func createSubjectRelationship(ctx context.Context, tx neo4j.ManagedTransaction, tripletID, subjectURI string) error {
	_, err := tx.Run(ctx,
		`MATCH (tr:Triplet {id: $tripletID})
		 MATCH (s:Resource {uri: $subjectURI})
		 MERGE (tr)-[:SUBJECT]->(s)`,
		map[string]any{
			"tripletID":  tripletID,
			"subjectURI": subjectURI,
		})
	return err
}

func createPredicateRelationship(ctx context.Context, tx neo4j.ManagedTransaction, tripletID, predicateURI string) error {
	_, err := tx.Run(ctx,
		`MATCH (tr:Triplet {id: $tripletID})
		 MATCH (p:Predicate {uri: $predicateURI})
		 MERGE (tr)-[:PREDICATE]->(p)`,
		map[string]any{
			"tripletID":    tripletID,
			"predicateURI": predicateURI,
		})
	return err
}

func createObjectRelationship(ctx context.Context, tx neo4j.ManagedTransaction, tripletID string, obj core.Object) error {
	if obj.Kind == "uri" {
		_, err := tx.Run(ctx,
			`MATCH (tr:Triplet {id: $tripletID})
			 MATCH (o:Resource {uri: $objectURI})
			 MERGE (tr)-[:OBJECT]->(o)`,
			map[string]any{
				"tripletID": tripletID,
				"objectURI": obj.Value,
			})
		return err
	}

	// literal
	_, err := tx.Run(ctx,
		`MATCH (tr:Triplet {id: $tripletID})
		 MATCH (l:Literal {value: $value, datatype: $datatype, language: $language})
		 MERGE (tr)-[:OBJECT]->(l)`,
		map[string]any{
			"tripletID": tripletID,
			"value":     obj.Value,
			"datatype":  obj.Datatype,
			"language":  obj.Language,
		})
	return err
}

func createBelongsToRelationship(ctx context.Context, tx neo4j.ManagedTransaction, tripletID, topicURI string) error {
	_, err := tx.Run(ctx,
		`MATCH (tr:Triplet {id: $tripletID})
		 MATCH (t:Topic {uri: $topicURI})
		 MERGE (tr)-[:BELONGS_TO]->(t)`,
		map[string]any{
			"tripletID": tripletID,
			"topicURI":  topicURI,
		})
	return err
}

// LoadTopics creates Topic nodes from a list of topics.
func (c *Client) LoadTopics(ctx context.Context, topics []core.Topic) error {
	session := c.NewSession(ctx, neo4j.SessionConfig{})
	defer closer.CloseOrLogContext(ctx, c.log, session)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		for _, topic := range topics {
			uri := topic.Slug
			if err := createTopicNode(ctx, tx, uri); err != nil {
				return nil, fmt.Errorf("create topic %s: %w", uri, err)
			}
			_, err := tx.Run(ctx,
				"MATCH (t:Topic {uri: $uri}) SET t.name = $name, t.description = $description, t.sparql = $sparql",
				map[string]any{
					"uri":         uri,
					"name":        topic.Name,
					"description": topic.Description,
					"sparql":      topic.SPARQL,
				})
			if err != nil {
				return nil, fmt.Errorf("set properties for topic %s: %w", uri, err)
			}
		}
		return nil, nil
	})

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

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// LoadQuads imports RDF quads into the graph.
func (c *Client) LoadQuads(ctx context.Context, quads []core.Quad) error {
	session := c.NewSession(ctx, neo4j.SessionConfig{})
	defer closer.CloseOrLogContext(ctx, c.log, session)

	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		labelPredicates := map[string]bool{
			"http://www.w3.org/2000/01/rdf-schema#label":    true,
			"http://www.w3.org/2004/02/skos/core#prefLabel": true,
		}

		resourceLabels := make(map[string]string)
		predicateLabels := make(map[string]string)

		for _, q := range quads {
			if q.Object.Kind != "literal" {
				continue
			}
			predURI := string(q.Predicate)
			if labelPredicates[predURI] {
				subjURI := string(q.Subject)
				if _, exists := resourceLabels[subjURI]; !exists {
					resourceLabels[subjURI] = q.Object.Value
				}

				if _, exists := predicateLabels[subjURI]; !exists {
					predicateLabels[subjURI] = q.Object.Value
				}
			}
		}

		for _, q := range quads {
			subjURI := string(q.Subject)
			subjLabel := resourceLabels[subjURI]
			if err := createResourceNode(ctx, tx, subjURI, subjLabel); err != nil {
				return nil, fmt.Errorf("subject node %s: %w", q.Subject, err)
			}

			predURI := string(q.Predicate)
			predLabel := predicateLabels[predURI]
			if err := createPredicateNode(ctx, tx, predURI, predLabel); err != nil {
				return nil, fmt.Errorf("predicate node %s: %w", q.Predicate, err)
			}

			if q.Object.Kind == "uri" {
				objURI := q.Object.Value
				objLabel := resourceLabels[objURI]
				if err := createResourceNode(ctx, tx, objURI, objLabel); err != nil {
					return nil, fmt.Errorf("object resource %s: %w", objURI, err)
				}
			} else {
				if err := createLiteralNode(ctx, tx, q.Object); err != nil {
					return nil, fmt.Errorf("literal node: %w", err)
				}
			}

			if q.Graph != "" {
				if err := createTopicNode(ctx, tx, q.Graph); err != nil {
					return nil, fmt.Errorf("topic node %s: %w", q.Graph, err)
				}
			}

			tripletID := generateTripletID(subjURI, predURI, q.Object)
			if err := createTripletNode(ctx, tx, tripletID); err != nil {
				return nil, fmt.Errorf("triplet node %s: %w", tripletID, err)
			}

			if err := createSubjectRelationship(ctx, tx, tripletID, subjURI); err != nil {
				return nil, fmt.Errorf("subject relationship: %w", err)
			}
			if err := createPredicateRelationship(ctx, tx, tripletID, predURI); err != nil {
				return nil, fmt.Errorf("predicate relationship: %w", err)
			}
			if err := createObjectRelationship(ctx, tx, tripletID, q.Object); err != nil {
				return nil, fmt.Errorf("object relationship: %w", err)
			}
			if q.Graph != "" {
				if err := createBelongsToRelationship(ctx, tx, tripletID, q.Graph); err != nil {
					return nil, fmt.Errorf("belongs_to relationship: %w", err)
				}
			}
		}
		return nil, nil
	})

	return err
}
