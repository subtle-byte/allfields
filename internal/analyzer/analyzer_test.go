package analyzer

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestWithFixtures(t *testing.T) {
	fixturesDir, err := filepath.Abs("./testdata")
	if err != nil {
		t.Fatal(err.Error())
	}
	// analysistest.Run uses "GOPATH=<fixturesDir> GO111MODULE=off", so we use GOPATH-like dir structure in the fixturesDir.
	analysistest.Run(t, fixturesDir, Analyzer, "github.com/subtle-byte/allfields_fixtures")
}
