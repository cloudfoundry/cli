package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
)

// NB: this assumes that translation strings are globally unique
//     as of the day we wrote this, they are not unique
func main() {
	walkTranslationFilesAndPromptUser()
}

func walkTranslationFilesAndPromptUser() {
	stringsFromCode := readSourceCode()

	dir := "cf/i18n/resources"
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err.Error())
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(info.Name()) != ".json" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			panic(err.Error())
		}
		data, err := ioutil.ReadAll(file)
		if err != nil {
			panic(err.Error())
		}
		maps := []map[string]string{}
		err = json.Unmarshal(data, &maps)
		if err != nil {
			panic(err.Error())
		}

		indicesToRemove := []int{}
		for index, tmap := range maps {
			str := tmap["id"]

			foundStr := false
			for _, codeStr := range stringsFromCode {
				if codeStr == str {
					foundStr = true
					break
				}
			}

			if !foundStr {
				fmt.Printf("Did not find this string in the source code:\n")
				fmt.Printf("'%s'\n", str)
				println()

				answer := ""
				fmt.Printf("Would you like to delete it from %s? [y|n]", path)
				fmt.Fscanln(os.Stdin, &answer)

				if answer == "y" {
					indicesToRemove = append(indicesToRemove, index)
				}
			}
		}

		if len(indicesToRemove) > 0 {
			println("Removing", len(indicesToRemove), "translations from", path)

			newMaps := []map[string]string{}
			for i, mapp := range maps {

				foundIndex := false
				for _, index := range indicesToRemove {
					if index == i {
						foundIndex = true
						break
					}
				}

				if !foundIndex {
					newMaps = append(newMaps, mapp)
				}
			}

			bytes, err := json.Marshal(newMaps) // consider json.MarshalIndent
			if err != nil {
				panic(err.Error())
			}

			newFile, err := os.Create(path)
			if err != nil {
				panic(err.Error())
			}

			_, err = newFile.Write(bytes)
			if err != nil {
				panic(err.Error())
			}
		}

		return nil
	})
}

func readSourceCode() []string {
	strings := []string{}

	dir, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}

	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			panic(err.Error())
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(info.Name()) != ".go" {
			return nil
		}

		fileSet := token.NewFileSet()
		astFile, err := parser.ParseFile(fileSet, path, nil, 0)
		if err != nil {
			panic(err.Error())
		}

		for _, declaration := range astFile.Decls {
			ast.Inspect(declaration, func(node ast.Node) bool {
				callExpr, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				funcNode, ok := callExpr.Fun.(*ast.Ident)
				if !ok {
					return true
				}

				if funcNode.Name != "T" {
					return true
				}

				firstArg := callExpr.Args[0]

				argAsString, ok := firstArg.(*ast.BasicLit)
				if !ok {
					return true
				}

				// remove quotes around string literal
				end := len(argAsString.Value) - 1
				strings = append(strings, argAsString.Value[1:end])
				return true
			})
		}

		return nil
	})

	return strings
}
