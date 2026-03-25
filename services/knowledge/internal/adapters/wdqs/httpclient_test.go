package wdqs_test

import (
	"log/slog"
	"testing"

	"github.com/penkovgd/erudition-app/services/knowledge/internal/adapters/wdqs"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// func prepare(t *testing.T) *wdqs.Client {
// 	t.Helper()

// 	client, err := wdqs.New(slog.Default())
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	return client
// }

// func TestClient_Extract(t *testing.T) {
// 	tests := []struct {
// 		name    string
// 		wantErr bool
// 	}{
// 		{
// 			name:    "mona",
// 			wantErr: false,
// 		},
// 		{
// 			name:    "empty",
// 			wantErr: true,
// 		},
// 		{
// 			name:    "invalid-syntax",
// 			wantErr: true,
// 		},
// 	}

// 	client := prepare(t)

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			t.Parallel()

// 			inputData, goldenData := testutils.ReadTestFiles(t, "testdata", tt.name)

// 			gotData, gotErr := client.Extract(context.Background(), string(inputData))

// 			if gotErr != nil {
// 				if !tt.wantErr {
// 					// got error but don't want it
// 					t.Fatal(gotErr)
// 				}
// 				// got error and want it - check gotErr and golden
// 				if !strings.Contains(gotErr.Error(), strings.TrimSpace(string(goldenData))) {
// 					t.Errorf("Extract() error mismatch\n got: %s\nwant: %s", gotErr.Error(), goldenData)
// 				}
// 				return
// 			}
// 			if tt.wantErr {
// 				// no error but want it
// 				t.Fatal("Extract() succeeded unexpectedly")
// 			}

// 			var got, want any
// 			if err := json.Unmarshal(gotData, &got); err != nil {
// 				t.Fatal(err)
// 			}
// 			if err := json.Unmarshal(goldenData, &want); err != nil {
// 				t.Fatal(err)
// 			}

// 			if !reflect.DeepEqual(got, want) {
// 				t.Errorf("Extract() = %q, want %q", got, want)
// 			}
// 		})
// 	}

// }

func TestJsonldToQuads(t *testing.T) {
	tests := []struct {
		name    string
		jsonld  []byte
		want    []core.Quad
		wantErr bool
	}{
		{
			name: "simple",
			jsonld: []byte(`{
				"@id": "http://example.com/subject",
				"http://example.com/predicate": "object"
			}`),
			want: []core.Quad{
				{
					Subject:   core.URI("http://example.com/subject"),
					Predicate: core.URI("http://example.com/predicate"),
					Object: core.Object{
						Kind:     "literal",
						Value:    "object",
						Datatype: "http://www.w3.org/2001/XMLSchema#string",
					},
				},
			},
		},
		{
			name: "object is URI",
			jsonld: []byte(`{
				"@context": {
					"ex": "http://example.org/"
				},
				"@id": "ex:mona_lisa",
				"ex:author": {
					"@id": "ex:leonardo"
				}
			}`),
			want: []core.Quad{
				{
					Subject:   core.URI("http://example.org/mona_lisa"),
					Predicate: core.URI("http://example.org/author"),
					Object: core.Object{
						Kind:  "uri",
						Value: "http://example.org/leonardo",
					},
				},
			},
		},
		{
			name: "object is lang string",
			jsonld: []byte(`{
				"@context": {
					"ex": "http://example.org/"
				},
				"@id": "ex:mona_lisa",
				"ex:name": {
					"@value": "Mona Lisa",
					"@language": "en"
				}
			}`),
			want: []core.Quad{
				{
					Subject:   core.URI("http://example.org/mona_lisa"),
					Predicate: core.URI("http://example.org/name"),
					Object: core.Object{
						Kind:     "literal",
						Value:    "Mona Lisa",
						Datatype: "http://www.w3.org/1999/02/22-rdf-syntax-ns#langString",
						Language: "en",
					},
				},
			},
		},
		{
			name: "object is date literal",
			jsonld: []byte(`{
				"@context": {
					"ex": "http://example.org/",
					"xsd": "http://www.w3.org/2001/XMLSchema#"
				},
				"@id": "ex:mona_lisa",
				"ex:creationDate": {
					"@value": "1503-01-01",
					"@type": "xsd:date"
				}
			}`),
			want: []core.Quad{
				{
					Subject:   core.URI("http://example.org/mona_lisa"),
					Predicate: core.URI("http://example.org/creationDate"),
					Object: core.Object{
						Kind:     "literal",
						Value:    "1503-01-01",
						Datatype: "http://www.w3.org/2001/XMLSchema#date",
					},
				},
			},
		},
		{
			name: "object is integer literal",
			jsonld: []byte(`{
				"@context": {
					"ex": "http://example.org/",
					"xsd": "http://www.w3.org/2001/XMLSchema#"
				},
				"@id": "ex:mona_lisa",
				"ex:height": {
					"@value": "77",
					"@type": "xsd:integer"
				}
			}`),
			want: []core.Quad{
				{
					Subject:   core.URI("http://example.org/mona_lisa"),
					Predicate: core.URI("http://example.org/height"),
					Object: core.Object{
						Kind:     "literal",
						Value:    "77",
						Datatype: "http://www.w3.org/2001/XMLSchema#integer",
					},
				},
			},
		},
		{
			name: "blank node is not supported",
			jsonld: []byte(`{
				"@context": {
					"ex": "http://example.org/"
				},
				"ex:name": "Jane Doe"
			}`),
			want:    []core.Quad{},
			wantErr: true,
		},
		{
			name: "named graph (quads)",
			jsonld: []byte(`{
				"@context": {
					"ex": "http://example.org/"
				},
				"@graph": {
					"@id": "ex:graph1",
					"@graph": [
					{
						"@id": "ex:mona_lisa",
						"ex:author": {
						"@id": "ex:leonardo"
						}
					}
					]
				}
			}`),
			want:    []core.Quad{},
			wantErr: true,
		},
	}

	c, err := wdqs.New(slog.Default())
	require.NoError(t, err)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotErr := c.JsonldToQuads(tt.jsonld)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("JsonldToQuads() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("JsonldToQuads() succeeded unexpectedly")
			}
			assert.ElementsMatch(t, got, tt.want)
		})
	}
}
