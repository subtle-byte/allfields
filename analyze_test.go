package allfields

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyze(t *testing.T) {
	type A struct {
		Name string
		Age  int
	}
	_ = A{
		Name: "John",
		//allfields:lint
	}
	_ = A{
		Name: "John",
		Age:  10,
		//allfields
	}
	_ = A{
		//allfields
	}
	_ = A{}
	errs := []string(nil)
	Analyze(AnalyzeConfig{
		PackagesPattern: ".",
		ReportErr: func(message string) {
			errs = append(errs, message)
		},
		Tests: true,
	})
	for _, err := range errs {
		t.Log(err)
	}
	require.Len(t, errs, 2)
	assert.Contains(t, errs[0], "analyze_test.go:15:6: field Age is not set")
	assert.Contains(t, errs[1], "analyze_test.go:24:6: fields Name, Age are not set")
}
