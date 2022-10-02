package analyzer

import (
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "allfields",
	Doc:  "Checks marked (with //allfields comment) struct literals for exhaustiveness",
	Run:  run,
}

func getAllfieldsComments(pass *analysis.Pass, file *ast.File) []*ast.Comment {
	allfieldsComments := []*ast.Comment(nil)
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			text := strings.TrimSpace(comment.Text)
			if i := strings.Index(text, " // want"); i != -1 { // for testing
				text = text[:i]
			}
			if text == "//allfields" {
				allfieldsComments = append(allfieldsComments, comment)
			} else if strings.HasPrefix(text, "//allfields ") {
				if text != "//allfields " {
					pass.ReportRangef(comment, "invalid allfields comment")
				}
			}
		}
	}
	// Sort by position in ascending order
	sort.Slice(allfieldsComments, func(i, j int) bool {
		return allfieldsComments[i].Pos() < allfieldsComments[j].Pos()
	})
	return allfieldsComments
}

func findCommentForCompositeLiteral(pass *analysis.Pass, astCompositeLit *ast.CompositeLit, allfieldsComments []*ast.Comment) (*ast.Comment, bool) {
	// Index of the first allfields comment after the composite literal start
	commentI := sort.Search(len(allfieldsComments), func(i int) bool {
		return allfieldsComments[i].Pos() >= astCompositeLit.Lbrace
	})
commentsFor:
	for {
		if commentI >= len(allfieldsComments) {
			// allfields comment not found
			return nil, false
		}
		comment := allfieldsComments[commentI]
		if comment.Pos() > astCompositeLit.Rbrace {
			// allfields comment is after the composite literal end, so it's not for this composite literal
			return nil, false
		}
		for _, elem := range astCompositeLit.Elts {
			if elem.Pos() <= comment.Pos() && comment.Pos() <= elem.End() {
				// allfields comment is not placed directly in the composite literal (it is placed deeper)
				commentI++
				continue commentsFor
			}
		}
		return comment, true
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		allfieldsComments := getAllfieldsComments(pass, file)
		usedComments := map[*ast.Comment]bool{}
		ast.Inspect(file, func(node ast.Node) bool {
			astCompositeLit, ok := node.(*ast.CompositeLit)
			if !ok {
				return true
			}

			comment, ok := findCommentForCompositeLiteral(pass, astCompositeLit, allfieldsComments)
			if !ok {
				return true
			}

			typesObj, ok := pass.TypesInfo.Types[astCompositeLit]
			if !ok {
				pass.ReportRangef(astCompositeLit, "internal allfields's error: no types info for composite literal, please remove allfields comment for a while and report about the bug")
				return true
			}

			typesStruct, ok := typesObj.Type.Underlying().(*types.Struct)
			if !ok {
				return true
			}

			usedComments[comment] = true

			fields := map[string]bool{} // field name -> is used
			fieldsOrdered := []string(nil)
			for i := 0; i < typesStruct.NumFields(); i++ {
				field := typesStruct.Field(i)
				if !field.Exported() && field.Pkg() != pass.Pkg {
					continue
				}
				fields[field.Name()] = false
				fieldsOrdered = append(fieldsOrdered, field.Name())
			}

			for _, compositeElem := range astCompositeLit.Elts {
				keyValue, ok := compositeElem.(*ast.KeyValueExpr)
				if !ok {
					pass.ReportRangef(comment, "allfields comment is placed in a non-keyed composite struct literal")
					return true // we don't care about non-keyed literals
				}
				field := keyValue.Key.(*ast.Ident).Name
				fields[field] = true
			}

			notSetFields := []string(nil)
			for _, field := range fieldsOrdered {
				if !fields[field] {
					notSetFields = append(notSetFields, field)
				}
			}
			if len(notSetFields) > 0 {
				if len(notSetFields) == 1 {
					pass.ReportRangef(astCompositeLit, "field %s is not set", notSetFields[0])
				} else {
					pass.ReportRangef(astCompositeLit, "fields %s are not set", strings.Join(notSetFields, ", "))
				}
			}

			return true
		})
		for _, comment := range allfieldsComments {
			if !usedComments[comment] {
				pass.ReportRangef(comment, "allfields comment is not used")
			}
		}
	}
	return nil, nil
}
