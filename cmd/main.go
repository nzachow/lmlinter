package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// TestCase represents a test case with its name, string representation, and tested function code
type TestCase struct {
	Name           string
	Str            string
	TestedFunction string
}

// Function to find all test cases in a Go test file
func findTestCases(filePath string) (map[string][]TestCase, error) {
	// Read the file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	// Create a new file set
	fset := token.NewFileSet()
	// Parse the file content
	node, err := parser.ParseFile(fset, "", content, parser.AllErrors)
	if err != nil {
		return nil, err
	}
	// Map to store test case names and their sub-tests
	testCases := make(map[string][]TestCase)
	// Inspect the AST
	ast.Inspect(node, func(n ast.Node) bool {
		// Find function declarations
		if fn, ok := n.(*ast.FuncDecl); ok {
			// Check if the function name starts with "Test"
			if strings.HasPrefix(fn.Name.Name, "Test") {
				// Initialize the slice for sub-tests
				testCases[fn.Name.Name] = []TestCase{}
				// Find sub-tests in table-driven tests
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					// Look for composite literals
					if cl, ok := n.(*ast.CompositeLit); ok {
						// Check if the type is a slice of structs
						if _, ok := cl.Type.(*ast.ArrayType); ok {
							for _, elt := range cl.Elts {
								if kv, ok := elt.(*ast.CompositeLit); ok {
									var testCase TestCase
									for _, elt := range kv.Elts {
										if kv, ok := elt.(*ast.KeyValueExpr); ok {
											if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "name" {
												if value, ok := kv.Value.(*ast.BasicLit); ok && value.Kind == token.STRING {
													testCase.Name = strings.Trim(value.Value, "\"")
												}
											}
										}
									}
									// Get the string representation of the test case
									var sb strings.Builder
									printer.Fprint(&sb, fset, kv)
									testCase.Str = sb.String()
									testCases[fn.Name.Name] = append(testCases[fn.Name.Name], testCase)
								}
							}
						}
					}
					return true
				})
			}
		}
		return true
	})
	return testCases, nil
}

// Function to find the implementation of a function in a Go file
func findFunctionImplementation(filePath, functionName string) (string, error) {
	// Read the file content
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// Create a new file set
	fset := token.NewFileSet()
	// Parse the file content
	node, err := parser.ParseFile(fset, "", content, parser.AllErrors)
	if err != nil {
		return "", err
	}
	// Inspect the AST to find the function implementation
	var functionCode string
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Name.Name == functionName {
				var sb strings.Builder
				printer.Fprint(&sb, fset, fn)
				functionCode = sb.String()
				return false
			}
		}
		return true
	})
	if functionCode == "" {
		return "", fmt.Errorf("function %s not found in file %s", functionName, filePath)
	}
	return functionCode, nil
}

// Function to find the function name that populates the 'got' variable
func findTestedFunctionName(fn *ast.FuncDecl) string {
	var functionName string
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if assignStmt, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assignStmt.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok && ident.Name == "got" {
					if callExpr, ok := assignStmt.Rhs[0].(*ast.CallExpr); ok {
						if selExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
							functionName = selExpr.Sel.Name
							return false
						}
					}
				}
			}
		}
		return true
	})
	return functionName
}

func main() {
	// Define a flag for the file path
	filePath := flag.String("path", "", "Path to the Go test file")
	flag.Parse()
	// Check if the file path is provided
	if *filePath == "" {
		fmt.Println("Please provide a path to the Go test file using the -path flag.")
		os.Exit(1)
	}
	// Find test cases in the specified file
	testCases, err := findTestCases(*filePath)
	if err != nil {
		log.Fatalf("Error finding test cases: %v", err)
	}
	// Determine the implementation file path
	implFilePath := strings.Replace(*filePath, "_test.go", ".go", 1)
	// Parse the test file to find the tested function names
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, *filePath, nil, parser.AllErrors)
	if err != nil {
		log.Fatalf("Error parsing test file: %v", err)
	}
	// Find the implementation of the tested functions
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "Test") {
			testedFunctionName := findTestedFunctionName(fn)
			if testedFunctionName != "" {
				functionCode, err := findFunctionImplementation(implFilePath, testedFunctionName)
				if err != nil {
					log.Printf("Error finding function implementation: %v", err)
				} else {
					for i, _ := range testCases[fn.Name.Name] {
						testCases[fn.Name.Name][i].TestedFunction = functionCode
					}
				}
			}
		}
		return true
	})
	// Print the test cases
	fmt.Println("Test cases found:")
	for testName, subTests := range testCases {
		fmt.Printf("%s:\n", testName)
		for _, subTest := range subTests {
			if subTest.TestedFunction != "" {
				p := createPrompt(subTest.Name, subTest.Str, subTest.TestedFunction)
				fmt.Printf("%s\n\n", p)
			}
		}
	}
}

func createPrompt(testName, testData, fnImplementation string) string {
	res := "Given this implementation: \n```" + fnImplementation + "```\n\n"

	res += "And this test case: \n```" + testData + "```\n\n"

	res += "The name '" + testName + "' is a good choice?"

	return res
}
