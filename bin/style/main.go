package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
)

type warning struct {
	message string
	token.Position
}

type visitor struct {
	fileSet *token.FileSet

	lastConstSpec    string
	lastFuncDecl     string
	lastReceiverFunc string
	lastReceiver     string
	lastTypeSpec     string
	lastVarSpec      string

	warnings []warning
}

func (v *visitor) Visit(node ast.Node) ast.Visitor {
	switch typedNode := node.(type) {
	case *ast.File:
		return v
	case *ast.GenDecl:
		if typedNode.Tok == token.CONST {
			v.checkConst(typedNode)
		} else if typedNode.Tok == token.VAR {
			v.checkVar(typedNode)
		}
		return v
	case *ast.FuncDecl:
		v.checkFunc(typedNode)
	case *ast.TypeSpec:
		v.checkType(typedNode)
	}

	return nil
}

func (v *visitor) addWarning(pos token.Pos, message string, subs ...interface{}) {
	coloredSubs := make([]interface{}, len(subs))
	for i, sub := range subs {
		coloredSubs[i] = color.CyanString(sub.(string))
	}

	v.warnings = append(v.warnings, warning{
		message:  fmt.Sprintf(message, coloredSubs...),
		Position: v.fileSet.Position(pos),
	})
}

func (v *visitor) checkConst(node *ast.GenDecl) {
	constName := node.Specs[0].(*ast.ValueSpec).Names[0].Name

	if v.lastFuncDecl != "" {
		v.addWarning(node.Pos(), "constant %s defined after a function declaration", constName)
	}
	if v.lastTypeSpec != "" {
		v.addWarning(node.Pos(), "constant %s defined after a type declaration", constName)
	}
	if v.lastVarSpec != "" {
		v.addWarning(node.Pos(), "constant %s defined after a variable declaration", constName)
	}

	if strings.Compare(constName, v.lastConstSpec) == -1 {
		v.addWarning(node.Pos(), "constant %s defined after constant %s", constName, v.lastConstSpec)
	}

	v.lastConstSpec = constName
}

func (v *visitor) checkFunc(node *ast.FuncDecl) {
	if node.Recv != nil {
		v.checkFuncWithReceiver(node)
	} else {
		funcName := node.Name.Name

		if strings.Compare(funcName, v.lastFuncDecl) == -1 {
			v.addWarning(node.Pos(), "function %s defined after function %s", funcName, v.lastFuncDecl)
		}

		v.lastFuncDecl = funcName
	}
}

func (v *visitor) checkFuncWithReceiver(node *ast.FuncDecl) {
	funcName := node.Name.Name

	var receiver string
	switch typedType := node.Recv.List[0].Type.(type) {
	case *ast.Ident:
		receiver = typedType.Name
	case *ast.StarExpr:
		receiver = typedType.X.(*ast.Ident).Name
	}
	if v.lastFuncDecl != "" {
		v.addWarning(node.Pos(), "method %s.%s defined after function %s", receiver, funcName, v.lastFuncDecl)
	}
	if v.lastTypeSpec != "" && receiver != v.lastTypeSpec {
		v.addWarning(node.Pos(), "method %s.%s should be defined immediately after type %s", receiver, funcName, receiver)
	}
	if receiver == v.lastReceiver {
		if strings.Compare(funcName, v.lastReceiverFunc) == -1 {
			v.addWarning(node.Pos(), "method %s.%s defined after method %s.%s", receiver, funcName, receiver, v.lastReceiverFunc)
		}
	}

	v.lastReceiver = receiver
	v.lastReceiverFunc = funcName
}

func (v *visitor) checkType(node *ast.TypeSpec) {
	typeName := node.Name.Name
	if v.lastFuncDecl != "" {
		v.addWarning(node.Pos(), "type declaration %s defined after a function declaration", typeName)
	}
	v.lastTypeSpec = typeName
}

func (v *visitor) checkVar(node *ast.GenDecl) {
	varName := node.Specs[0].(*ast.ValueSpec).Names[0].Name

	if v.lastFuncDecl != "" {
		v.addWarning(node.Pos(), "variable %s defined after a function declaration", varName)
	}
	if v.lastTypeSpec != "" {
		v.addWarning(node.Pos(), "variable %s defined after a type declaration", varName)
	}

	if strings.Compare(varName, v.lastVarSpec) == -1 {
		v.addWarning(node.Pos(), "variable %s defined after variable %s", varName, v.lastVarSpec)
	}

	v.lastVarSpec = varName
}

func main() {
	var allWarnings []warning

	fileSet := token.NewFileSet()

	err := filepath.Walk(os.Args[1], func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		base := filepath.Base(path)
		if base == "vendor" || base == ".git" || strings.HasSuffix(base, "fakes") {
			return filepath.SkipDir
		}

		packages, err := parser.ParseDir(fileSet, path, shouldParseFile, 0)
		if err != nil {
			return err
		}

		var packageNames []string
		for packageName, _ := range packages {
			packageNames = append(packageNames, packageName)
		}
		sort.Strings(packageNames)

		for _, packageName := range packageNames {
			var fileNames []string
			for fileName, _ := range packages[packageName].Files {
				fileNames = append(fileNames, fileName)
			}
			sort.Strings(fileNames)

			for _, fileName := range fileNames {
				v := visitor{
					fileSet: fileSet,
				}
				ast.Walk(&v, packages[packageName].Files[fileName])
				allWarnings = append(allWarnings, v.warnings...)
			}
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	for _, warning := range allWarnings {
		fmt.Printf("%s +%d %s\n", color.CyanString(warning.Position.Filename), warning.Position.Line, warning.message)
	}

	if len(allWarnings) > 0 {
		os.Exit(1)
	}
}

func shouldParseFile(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}
