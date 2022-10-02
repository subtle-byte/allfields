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

type parsedComment struct {
	*ast.Comment
	IgnoredFields []string
}

func parseComment(pass *analysis.Pass, comment *ast.Comment) (_ parsedComment, ok bool) {
	text := strings.TrimSpace(comment.Text)
	if i := strings.Index(text, " // want"); i != -1 {
		text = text[:i]
	}
	if text == "//allfields" {
		return parsedComment{
			Comment: comment,
		}, true
	} else if strings.HasPrefix(text, "//allfields ") {
		if strings.HasPrefix(text, "//allfields ignore=") {
			return parsedComment{
				Comment:       comment,
				IgnoredFields: strings.Split(text[len("//allfields ignore="):], ","),
			}, true
		} else {
			pass.ReportRangef(comment, "invalid allfields comment")
			return parsedComment{}, false
		}
	}
	return parsedComment{}, false
}

func getAllfieldsComments(pass *analysis.Pass, file *ast.File) []parsedComment {
	allfieldsComments := []parsedComment(nil)
	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			allfieldParsedComment, ok := parseComment(pass, comment)
			if ok {
				allfieldsComments = append(allfieldsComments, allfieldParsedComment)
			}
		}
	}
	// Sort by position in ascending order
	sort.Slice(allfieldsComments, func(i, j int) bool {
		return allfieldsComments[i].Pos() < allfieldsComments[j].Pos()
	})
	return allfieldsComments
}

func findCommentForCompositeLiteral(pass *analysis.Pass, astCompositeLit *ast.CompositeLit, allfieldsComments []parsedComment) (parsedComment, bool) {
	// Index of the first allfields comment after the composite literal start
	commentI := sort.Search(len(allfieldsComments), func(i int) bool {
		return allfieldsComments[i].Pos() >= astCompositeLit.Lbrace
	})
commentsFor:
	for {
		if commentI >= len(allfieldsComments) {
			// allfields comment not found
			return parsedComment{}, false
		}
		comment := allfieldsComments[commentI]
		if comment.Pos() > astCompositeLit.Rbrace {
			// allfields comment is after the composite literal end, so it's not for this composite literal
			return parsedComment{}, false
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

			usedComments[comment.Comment] = true

			fields := map[string]bool{}
			unexportedFieldsFromAnotherPkg := map[string]bool{}
			fieldsOrdered := []string(nil)
			for i := 0; i < typesStruct.NumFields(); i++ {
				field := typesStruct.Field(i)
				if !field.Exported() && field.Pkg() != pass.Pkg {
					unexportedFieldsFromAnotherPkg[field.Name()] = true
					continue
				}
				fields[field.Name()] = false
				fieldsOrdered = append(fieldsOrdered, field.Name())
			}

			ignoredFields := map[string]bool{}
			for _, ignoredField := range comment.IgnoredFields {
				ignoredFields[ignoredField] = true
			}
			for ignoredField := range ignoredFields {
				if _, ok := fields[ignoredField]; !ok {
					if !unexportedFieldsFromAnotherPkg[ignoredField] {
						pass.ReportRangef(comment, "field %v is not present in the struct but ignored", ignoredField)
					} else {
						pass.ReportRangef(comment, "unexported field %v is not available in this package, so the field should not be ignored", ignoredField)
					}
					delete(ignoredFields, ignoredField)
				}
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
					if !ignoredFields[field] {
						notSetFields = append(notSetFields, field)
					} else {
						delete(ignoredFields, field)
					}
				}
			}
			for ignoredField := range ignoredFields {
				pass.ReportRangef(comment, "field %v is marked as ignored but is present in the struct literal", ignoredField)
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
			if !usedComments[comment.Comment] {
				pass.ReportRangef(comment, "allfields comment is not used")
			}
		}
	}
	return nil, nil
}
