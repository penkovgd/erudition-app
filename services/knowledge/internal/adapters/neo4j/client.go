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

	log.Debug("connected to neo4j")
	return &Client{Driver: driver, log: log}, nil
}

// Close closes neo4j driver. If he gives an error, panics
func (c *Client) Close(ctx context.Context) {
	closer.CloseOrPanicContext(ctx, c.Driver)
}

// Load loads given json-ld rdf into neo4j via n10s
func (c *Client) Load(ctx context.Context, jsonld core.JSONLD) error {
	result, err := neo4j.ExecuteQuery(ctx, c.Driver,
		`CALL n10s.rdf.import.inline($jsonld, "JSON-LD")`,
		map[string]any{"jsonld": string(jsonld)},
		neo4j.EagerResultTransformer,
		neo4j.ExecuteQueryWithDatabase("neo4j"),
	)
	if err != nil {
		return fmt.Errorf("execute neo4j query: %w", err)
	}

	if len(result.Records) == 0 {
		return fmt.Errorf("neo4j query returned no records")
	}

	terminationStatus, ok := result.Records[0].Get("terminationStatus")
	if !ok {
		return fmt.Errorf("record missing 'terminationStatus' field")
	}
	extraInfo, _ := result.Records[0].Get("extraInfo")

	c.log.Debug(
		"query result",
		"termination status", terminationStatus,
		"extra info", extraInfo,
	)

	if terminationStatus != "OK" {
		return fmt.Errorf("neo4j query failed: %v, extra info: %v", terminationStatus, extraInfo)
	}

	counters := result.Summary.Counters()
	c.log.Debug(
		"query summary",
		"query did write?", counters.ContainsUpdates(),
		"nodes created", counters.NodesCreated(),
		"labels added", counters.LabelsAdded(),
		"properties set", counters.PropertiesSet(),
		"relationships created", counters.RelationshipsCreated(),
	)
	return nil
}
