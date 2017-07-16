package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/fatih/color"
)

type warning struct {
	format string
	vars   []interface{}
	token.Position
}

type warningPrinter struct {
	warnings []warning
}

func (w warningPrinter) print(writer io.Writer) {
	w.sortWarnings()

	for _, warning := range w.warnings {
		coloredVars := make([]interface{}, len(warning.vars))
		for i, v := range warning.vars {
			coloredVars[i] = color.CyanString(v.(string))
		}

		message := fmt.Sprintf(warning.format, coloredVars...)

		fmt.Printf(
			"%s %s %s\n",
			color.MagentaString(warning.Position.Filename),
			color.MagentaString(fmt.Sprintf("+%d", warning.Position.Line)),
			message)
	}
}

func (w warningPrinter) sortWarnings() {
	sort.Slice(w.warnings, func(i int, j int) bool {
		if w.warnings[i].Position.Filename < w.warnings[j].Position.Filename {
			return true
		}
		if w.warnings[i].Position.Filename > w.warnings[j].Position.Filename {
			return false
		}

		if w.warnings[i].Position.Line < w.warnings[j].Position.Line {
			return true
		}
		if w.warnings[i].Position.Line > w.warnings[j].Position.Line {
			return false
		}

		iMessage := fmt.Sprintf(w.warnings[i].format, w.warnings[i].vars...)
		jMessage := fmt.Sprintf(w.warnings[j].format, w.warnings[j].vars...)

		return iMessage < jMessage
	})
}

type visitor struct {
	fileSet *token.FileSet

	lastConstSpec    string
	lastFuncDecl     string
	lastReceiverFunc string
	lastReceiver     string
	lastVarSpec      string
	typeSpecs        []string

	warnings []warning

	previousPass *visitor
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

func (v *visitor) addWarning(pos token.Pos, format string, vars ...interface{}) {
	v.warnings = append(v.warnings, warning{
		format:   format,
		vars:     vars,
		Position: v.fileSet.Position(pos),
	})
}

func (v *visitor) checkConst(node *ast.GenDecl) {
	constName := node.Specs[0].(*ast.ValueSpec).Names[0].Name

	if v.lastFuncDecl != "" {
		v.addWarning(node.Pos(), "constant %s defined after a function declaration", constName)
	}
	if len(v.typeSpecs) != 0 {
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
		if funcName == "Execute" || strings.HasPrefix(funcName, "New") {
			return
		}

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
	if len(v.typeSpecs) > 0 {
		lastTypeSpec := v.typeSpecs[len(v.typeSpecs)-1]
		if v.typeDefinedInFile(receiver) && receiver != lastTypeSpec {
			v.addWarning(node.Pos(), "method %s.%s should be defined immediately after type %s", receiver, funcName, receiver)
		}
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
	v.typeSpecs = append(v.typeSpecs, typeName)
}

func (v *visitor) checkVar(node *ast.GenDecl) {
	varName := node.Specs[0].(*ast.ValueSpec).Names[0].Name

	if v.lastFuncDecl != "" {
		v.addWarning(node.Pos(), "variable %s defined after a function declaration", varName)
	}
	if len(v.typeSpecs) != 0 {
		v.addWarning(node.Pos(), "variable %s defined after a type declaration", varName)
	}

	if strings.Compare(varName, v.lastVarSpec) == -1 {
		v.addWarning(node.Pos(), "variable %s defined after variable %s", varName, v.lastVarSpec)
	}

	v.lastVarSpec = varName
}

func (v *visitor) typeDefinedInFile(typeName string) bool {
	if v.previousPass == nil {
		return true
	}

	for _, definedTypeName := range v.previousPass.typeSpecs {
		if definedTypeName == typeName {
			return true
		}
	}

	return false
}

func check(fileSet *token.FileSet, path string) ([]warning, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if stat.IsDir() {
		return checkDir(fileSet, path)
	} else {
		return checkFile(fileSet, path)
	}
}

func checkDir(fileSet *token.FileSet, path string) ([]warning, error) {
	var warnings []warning

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}

		if shouldSkipDir(path) {
			return filepath.SkipDir
		}

		packages, err := parser.ParseDir(fileSet, path, shouldParseFile, 0)
		if err != nil {
			return err
		}

		for _, packag := range packages {
			for _, file := range packag.Files {
				warnings = append(warnings, walkFile(fileSet, file)...)
			}
		}

		return nil
	})

	return warnings, err
}

func checkFile(fileSet *token.FileSet, path string) ([]warning, error) {
	file, err := parser.ParseFile(fileSet, path, nil, 0)
	if err != nil {
		return nil, err
	}

	return walkFile(fileSet, file), nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [--] [FILE or DIRECTORY]...\n", os.Args[0])
		os.Exit(1)
	}

	var allWarnings []warning

	args := os.Args[1:]
	if args[0] == "--" {
		args = args[1:]
	}

	fileSet := token.NewFileSet()

	for _, arg := range args {
		warnings, err := check(fileSet, arg)
		if err != nil {
			panic(err)
		}
		allWarnings = append(allWarnings, warnings...)
	}

	warningPrinter := warningPrinter{
		warnings: allWarnings,
	}
	warningPrinter.print(os.Stdout)

	if len(allWarnings) > 0 {
		os.Exit(1)
	}
}

func shouldParseFile(info os.FileInfo) bool {
	return !strings.HasSuffix(info.Name(), "_test.go")
}

func shouldSkipDir(path string) bool {
	base := filepath.Base(path)
	return base == "vendor" || base == ".git" || strings.HasSuffix(base, "fakes")
}

func walkFile(fileSet *token.FileSet, file *ast.File) []warning {
	firstPass := visitor{
		fileSet: fileSet,
	}
	ast.Walk(&firstPass, file)

	v := visitor{
		fileSet:      fileSet,
		previousPass: &firstPass,
	}
	ast.Walk(&v, file)

	return v.warnings
}
