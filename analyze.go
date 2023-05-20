package allfields

import (
	"fmt"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/packages"
)

type AnalyzeConfig struct {
	PackagesPattern string // e.g. "./..."
	ReportErr       func(errMessage string)
	Tests           bool // if true, tests will be analyzed too
}

func Analyze(config AnalyzeConfig) {
	if config.ReportErr == nil {
		panic("config.ReportErr is nil, must be set")
	}
	if config.PackagesPattern == "" {
		config.ReportErr("AnalyzeConfig.PackagesPattern is empty, must be set")
		return
	}
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles |
			packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Tests: config.Tests,
	}, config.PackagesPattern)
	if err != nil {
		config.ReportErr("load packages: " + err.Error())
		return
	}
	reportedPositions := make(map[token.Pos]bool) // needed because test packages are separate packages
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			for _, err := range pkg.Errors {
				config.ReportErr(err.Error())
			}
			continue
		}
		_, err := Analyzer.Run(&analysis.Pass{
			Files:     pkg.Syntax,
			TypesInfo: pkg.TypesInfo,
			Pkg:       pkg.Types,
			Report: func(d analysis.Diagnostic) {
				if reportedPositions[d.Pos] {
					return
				}
				reportedPositions[d.Pos] = true
				pos := pkg.Fset.Position(d.Pos)
				config.ReportErr(pos.String() + ": " + d.Message)
			},
		})
		if err != nil {
			config.ReportErr(fmt.Sprintf("run analyzer for package %q: %v", pkg.PkgPath, err.Error()))
		}
	}
}
