package neo4j

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/neo4j/neo4j-go-driver/v6/neo4j"
	"github.com/penkovgd/erudition-app/pkg/closer"
)

type Client struct {
	driver neo4j.Driver
	log    *slog.Logger
}

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

func (c *Client) Close(ctx context.Context) {
	closer.CloseOrPanicContext(ctx, c.driver)
}
