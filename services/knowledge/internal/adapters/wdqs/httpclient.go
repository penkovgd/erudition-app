// Package wdqs responsible for querying Wikidata Query Service (WDQS). Provides http client to fetch RDF datasets
package wdqs

import (
	"context"
	"encoding/json"
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
	"github.com/piprate/json-gold/ld"
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

// Extract takes sparql query and makes request to wdqs. Returns core rdf triples
func (c *Client) Extract(ctx context.Context, topic core.Topic) ([]core.Quad, error) {
	sparql := topic.SPARQL
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

	quads, err := c.JsonldToQuads(jsonld)
	if err != nil {
		return nil, fmt.Errorf("convert json-ld to quads: %w", err)
	}

	for i := range quads {
		quads[i].Graph = topic.Slug
	}

	return quads, nil
}

// JsonldToQuads parses jsonld via json-gold module and maps it to core model
func (c *Client) JsonldToQuads(jsonld []byte) ([]core.Quad, error) {
	var jsonldObj any
	if err := json.Unmarshal(jsonld, &jsonldObj); err != nil {
		return nil, fmt.Errorf("unmarshal json-ld into Go struct: %w", err)
	}

	proc := ld.NewJsonLdProcessor()
	options := ld.NewJsonLdOptions("")

	rdf, err := proc.ToRDF(jsonldObj, options)
	if err != nil {
		return nil, fmt.Errorf("convert json-ld to RDF: %w", err)
	}
	rdfDataset, ok := rdf.(*ld.RDFDataset)
	if !ok {
		return nil, fmt.Errorf("expected *ld.RDFDataset, got %T", rdf)
	}

	quads := rdfDataset.GetQuads("@default")

	if len(quads) == 0 {
		return nil, fmt.Errorf("named graphs are not supported")
	}

	triples := make([]core.Quad, 0, len(quads))
	for _, q := range quads {
		triple := core.Quad{}

		switch subj := q.Subject.(type) {
		case ld.IRI:
			triple.Subject = core.URI(subj.GetValue())
		case ld.BlankNode:
			return nil, fmt.Errorf("blank nodes not supported")
		default:
			return nil, fmt.Errorf("expected ld.IRI or ld.BlankNode for subject, got %T", q.Subject)
		}

		pred, ok := q.Predicate.(ld.IRI)
		if !ok {
			return nil, fmt.Errorf("expected ld.IRI for predicate, got %T", q.Predicate)
		}
		triple.Predicate = core.URI(pred.GetValue())

		switch obj := q.Object.(type) {
		case ld.Literal:
			triple.Object = core.Object{
				Kind:     "literal",
				Value:    obj.Value,
				Datatype: obj.Datatype,
				Language: obj.Language,
			}
		case ld.IRI:
			triple.Object = core.Object{
				Kind:  "uri",
				Value: obj.GetValue(),
			}
		default:
			return nil, fmt.Errorf("expected ld.Literal or ld.IRI for object, got %T", q.Object)
		}

		triples = append(triples, triple)
	}

	return triples, nil
}
