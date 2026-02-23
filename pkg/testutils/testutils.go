// Package testutils provides utility functions for testing, such as reading golden files.
package testutils

import (
	"os"
	"path/filepath"
	"testing"
)

// ReadTestFiles reads the input and golden files for a given test case.
// dir is the directory where the test files are located (i.e. testdata).
// name - test case name. Required file names: <name>.input and <name>.golden.
func ReadTestFiles(t *testing.T, dir, name string) (input, golden []byte) {
	t.Helper()

	inputFile := filepath.Join(dir, name+".input")
	goldenFile := filepath.Join(dir, name+".golden")

	input, err := os.ReadFile(filepath.Clean(inputFile))
	if err != nil {
		t.Fatal(err)
	}
	golden, err = os.ReadFile(filepath.Clean(goldenFile))
	if err != nil {
		t.Fatal(err)
	}
	return input, golden
}
