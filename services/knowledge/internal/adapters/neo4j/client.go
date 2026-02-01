// Package neo4j is a neo4j adapter for storing knowledge graph
package neo4j

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/penkovgd/erudition-app/pkg/closer"
)

// Client uses the neo4j driver to send queries
type Client struct {
	driver neo4j.Driver
	log    *slog.Logger
}

// New cleates a neo4j client with a given credentials and verifies connection
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
	return &Client{driver: driver, log: log}, nil
}

// Close closes neo4j driver. If he gives an error, panics
func (c *Client) Close(ctx context.Context) {
	closer.CloseOrPanicContext(ctx, c.driver)
}
