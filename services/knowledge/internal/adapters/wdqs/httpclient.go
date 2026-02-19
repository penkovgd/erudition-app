// Package wdqs responsible for querying Wikidata Query Service (WDQS). Provides http client to fetch RDF datasets
package wdqs

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/penkovgd/erudition-app/pkg/closer"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
)

// Client is HTTP-client to query wikidata query service
type Client struct {
	baseURL   string
	userAgent string
	accept    string
	client    *http.Client
	log       *slog.Logger
}

// New creates a new http client
func New(log *slog.Logger) (*Client, error) {
	// TODO мб добавить более гибкую конфигурацию - хедеров, базового url
	return &Client{
		baseURL:   "https://query.wikidata.org/sparql",
		userAgent: "erudition-app-bot/0.0 (https://github.com/penkovgd/erudition-app) go-http-client/1.1",
		accept:    "application/ld+json",
		client:    &http.Client{Timeout: 60 * time.Second},
		log:       log,
	}, nil
}

// Extract takes sparql query and makes request to wdqs. Returns serialized json-ld
func (c *Client) Extract(ctx context.Context, sparql string) (core.JSONLD, error) {
	if strings.TrimSpace(sparql) == "" {
		return nil, errors.New("empty sparql query")
	}

	url := c.baseURL + "?query=" + url.QueryEscape(sparql)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}

	req.Header.Add("user-agent", c.userAgent)
	req.Header.Add("accept", c.accept)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("perform http request: %w", err)
	}
	defer closer.CloseOrLog(c.log, resp.Body)

	// TODO добавить повторные запросы если не получилось или если too many requests (retry after)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-OK http status: %s", resp.Status)
	}

	jsonld, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	return jsonld, nil
}
