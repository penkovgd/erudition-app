package wdqs_test

import (
	"context"
	"encoding/json"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	"github.com/penkovgd/erudition-app/pkg/testutils"
	"github.com/penkovgd/erudition-app/services/knowledge/internal/adapters/wdqs"
)

func prepare(t *testing.T) *wdqs.Client {
	t.Helper()

	client, err := wdqs.New(slog.Default())
	if err != nil {
		t.Fatal(err)
	}
	return client
}

func TestClient_Extract(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "mona",
			wantErr: false,
		},
		{
			name:    "empty",
			wantErr: true,
		},
		{
			name:    "invalid-syntax",
			wantErr: true,
		},
	}

	client := prepare(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inputData, goldenData := testutils.ReadTestFiles(t, "testdata", tt.name)

			gotData, gotErr := client.Extract(context.Background(), string(inputData))

			if gotErr != nil {
				if !tt.wantErr {
					// got error but don't want it
					t.Fatal(gotErr)
				}
				// got error and want it - check gotErr and golden
				if !strings.Contains(gotErr.Error(), strings.TrimSpace(string(goldenData))) {
					t.Errorf("Extract() error mismatch\n got: %s\nwant: %s", gotErr.Error(), goldenData)
				}
				return
			}
			if tt.wantErr {
				// no error but want it
				t.Fatal("Extract() succeeded unexpectedly")
			}

			var got, want any
			if err := json.Unmarshal(gotData, &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal(goldenData, &want); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("Extract() = %q, want %q", got, want)
			}
		})
	}

}
