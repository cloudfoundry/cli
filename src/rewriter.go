package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
)

func main() {
	var filepath string
	flag.StringVar(&filepath, "refactor", "", "")
	flag.Parse()

	println(filepath)

	fset := token.NewFileSet()
	newFile, err := parser.ParseFile(fset, filepath, nil, 0)
	if err != nil {
		panic(err)
	}

	refactorings := map[string]string{
		"Application":            "app",
		"AppSummary":             "appSummary",
		"Domain":                 "domain",
		"EventFields":            "event",
		"Route":                  "route",
		"RouteSummary":           "routeSummary",
		"Stack":                  "stack",
		"AppInstanceFields":      "appInstance",
		"ServicePlan":            "plan",
		"ServiceOffering":        "offering",
		"ServiceInstance":        "serviceInstance",
		"ServiceBindingFields":   "binding",
		"Quota":                  "quota",
		"ServiceAuthTokenFields": "authToken",
		"ServiceBroker":          "broker",
		"User":             "user",
		"Buildpack":              "buildpack",
		"Organization":           "org",
		"Space":                  "space",
	}

	for structName, varName := range refactorings {
		count := 1
		ast.Inspect(newFile, func(n ast.Node) bool {
			switch n := n.(type) {
			case *ast.CompositeLit:
				switch s := n.Type.(type) {
				case *ast.SelectorExpr:
					func() {
						defer func() {
							if err := recover(); err != nil {

							}
						}()

						if s.Sel.Name == structName {
							if len(n.Elts) == 0 {
								return
							}

							var name string
							if count == 1 {
								name = fmt.Sprintf("%s", varName)
							} else {
								name = fmt.Sprintf("%s%d", varName, count)
							}

							count++

							rewriteStructLiteralAsIdentifierAtTopOfBlock(newFile, n, name)
						}
					}()
				}
			}

			return true
		})
	}

	src, err := gofmtFile(newFile, fset)
	if err != nil {
		println(err.Error())
	}

	//	println(string(src))
	ioutil.WriteFile(filepath, src, 0666)
}

func rewriteStructLiteralAsIdentifierAtTopOfBlock(newFile *ast.File, literalToReplace *ast.CompositeLit, name string) {
	var (
		foundStmtNode           ast.Stmt
		blocksSeen              []*ast.BlockStmt
		stmtsSeen               []ast.Stmt
		deleteOriginalStatement bool
		newAssignStmtToken      = token.DEFINE
	)

	ast.Inspect(newFile, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.BlockStmt:
			if foundStmtNode == nil {
				blocksSeen = append(blocksSeen, node)
			}

		case ast.Stmt:
			stmtsSeen = append(stmtsSeen, node)
		}
		return true
	})

	var block *ast.BlockStmt
	var insertionIndex int

	for i := len(blocksSeen) - 1; i >= 0; i-- {
		b := blocksSeen[i]
		for j, stmt := range b.List {
			ast.Inspect(stmt, func(parentNode ast.Node) bool {
				switch parentNode := parentNode.(type) {
				case *ast.CompositeLit:
					if parentNode == literalToReplace {
						block = b
						insertionIndex = j
						return false
					}
				}
				return true
			})
		}
	}

	for i := len(stmtsSeen) - 1; i >= 0; i-- {
		s := stmtsSeen[i]
		ast.Inspect(s, func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.CompositeLit:
				if node == literalToReplace {
					foundStmtNode = s
					return false
				}
			}
			return true
		})
	}

	switch foundStmtNode := foundStmtNode.(type) {
	case *ast.AssignStmt:
		expr := foundStmtNode.Rhs[0]
		if expr == literalToReplace {
			deleteOriginalStatement = true

			switch ident := foundStmtNode.Lhs[0].(type) {
			case *ast.Ident:
				name = ident.Name
			}

			if foundStmtNode.Tok == token.ASSIGN {
				newAssignStmtToken = token.ASSIGN
			}
		}
	}

	lhsExpr := []ast.Expr{ast.NewIdent(name)}
	rhsExpr := []ast.Expr{&ast.CompositeLit{Type: literalToReplace.Type}}

	block.List = insert(block.List, insertionIndex, ast.AssignStmt{
		Lhs: lhsExpr,
		Rhs: rhsExpr,
		Tok: newAssignStmtToken,
	})
	insertionIndex++

	for _, elt := range literalToReplace.Elts {
		keyVal := elt.(*ast.KeyValueExpr)
		fieldName := keyVal.Key.(*ast.Ident)

		selector := &ast.SelectorExpr{
			X:   ast.NewIdent(name),
			Sel: ast.NewIdent(fieldName.Name),
		}
		innerLhs := []ast.Expr{selector}
		innerRhs := []ast.Expr{keyVal.Value}

		block.List = insert(block.List, insertionIndex, ast.AssignStmt{
			Lhs: innerLhs,
			Rhs: innerRhs,
			Tok: token.ASSIGN,
		})

		insertionIndex++
	}

	if deleteOriginalStatement {
		copy(block.List[insertionIndex:], block.List[insertionIndex+1:])
		block.List[len(block.List)-1] = &ast.EmptyStmt{}
		block.List = block.List[:len(block.List)-1]
	} else {
		replaceStructLiteralWithIdentifier(foundStmtNode, literalToReplace, name)
	}
}

func replaceStructLiteralWithIdentifier(statement ast.Node, n *ast.CompositeLit, name string) {
	ast.Inspect(statement, func(parentNode ast.Node) bool {
		switch parentNode := parentNode.(type) {
		case *ast.KeyValueExpr:
			if parentNode.Value == n {
				parentNode.Value = ast.NewIdent(name)
			}
		case *ast.CallExpr:
			for i, expr := range parentNode.Args {
				if expr == n {
					parentNode.Args[i] = ast.NewIdent(name)
				}
			}
		case *ast.AssignStmt:
			if parentNode.Rhs[0] == n {
				parentNode.Rhs[0] = ast.NewIdent(name)
			}
		case *ast.ReturnStmt:
			for i, expr := range parentNode.Results {
				if expr == n {
					parentNode.Results[i] = ast.NewIdent(name)
				}
			}
		case *ast.CompositeLit:
			for i, expr := range parentNode.Elts {
				if expr == n {
					parentNode.Elts[i] = ast.NewIdent(name)
				}
			}
		}

		return true
	})
}

func insert(list []ast.Stmt, i int, stmt ast.AssignStmt) []ast.Stmt {
	list = append(list, &stmt)
	copy(list[i+1:], list[i:])
	list[i] = &stmt
	return list
}

func gofmtFile(f *ast.File, fset *token.FileSet) ([]byte, error) {
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
