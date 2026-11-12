package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"sort"
)

// Extract a readable function name.
func getCallName(expr ast.Expr) string {
	switch fn := expr.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		return getCallName(fn.X) + "." + fn.Sel.Name
	}
	return "<unknown>"
}

// Collect package-level variables.
func getGlobalVariables(file *ast.File) map[string]bool {
	globals := make(map[string]bool)

	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}

		for _, spec := range gen.Specs {
			valueSpec := spec.(*ast.ValueSpec)

			for _, name := range valueSpec.Names {
				globals[name.Name] = true
			}
		}
	}

	return globals
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ./cmd/interleaveguard <go-file>")
		return
	}

	filePath := os.Args[1]
	fset := token.NewFileSet()

	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		fmt.Println("Error parsing file:", err)
		return
	}

	fmt.Println("Analyzing:", filePath)

	// Collect declared functions.
	declaredFunctions := make(map[string]bool)

	for _, declaration := range file.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok {
			declaredFunctions[fn.Name.Name] = true
		}
	}

	// Display functions.
	fmt.Println("\nFunctions and methods:")

	for _, declaration := range file.Decls {
		if fn, ok := declaration.(*ast.FuncDecl); ok {
			kind := "function"
			if fn.Recv != nil {
				kind = "method"
			}

			fmt.Printf("- %s (%s)\n", fn.Name.Name, kind)
		}
	}

	// Identify goroutine launches.
	fmt.Println("\nGoroutine targets:")

	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if goStmt, ok := n.(*ast.GoStmt); ok {
				fmt.Printf("- %s launches %s()\n",
					fn.Name.Name, getCallName(goStmt.Call.Fun))
			}
			return true
		})
	}

	// Identify package-level shared variables.
	globals := getGlobalVariables(file)

	fmt.Println("\nPackage-level variables:")

	globalNames := make([]string, 0, len(globals))
	for name := range globals {
		globalNames = append(globalNames, name)
	}
	sort.Strings(globalNames)

	for _, name := range globalNames {
		fmt.Println("-", name)
	}

	// Analyze reads and writes inside each function.
	fmt.Println("\nShared-state access analysis:")

	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		// Record global-variable writes by source position.
		writes := make(map[token.Pos]string)

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			switch node := n.(type) {

			case *ast.AssignStmt:
				accessType := "WRITE"

				// Compound assignments also read the old value.
				if node.Tok != token.ASSIGN {
					accessType = "READ + WRITE"
				}

				for _, lhs := range node.Lhs {
					if id, ok := lhs.(*ast.Ident); ok && globals[id.Name] {
						writes[id.Pos()] = accessType
					}
				}

			case *ast.IncDecStmt:
				if id, ok := node.X.(*ast.Ident); ok && globals[id.Name] {
					writes[id.Pos()] = "READ + WRITE"
				}
			}

			return true
		})

		// Collect accesses.
		accesses := make(map[string]map[string]bool)

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			id, ok := n.(*ast.Ident)
			if !ok || !globals[id.Name] {
				return true
			}

			accessType := "READ"

			if writeType, exists := writes[id.Pos()]; exists {
				accessType = writeType
			}

			if accesses[id.Name] == nil {
				accesses[id.Name] = make(map[string]bool)
			}

			accesses[id.Name][accessType] = true

			return true
		})

		// Print results in a predictable order.
		variableNames := make([]string, 0, len(accesses))
		for name := range accesses {
			variableNames = append(variableNames, name)
		}
		sort.Strings(variableNames)

		for _, name := range variableNames {
			var types []string

			for _, accessType := range []string{
				"READ", "WRITE", "READ + WRITE",
			} {
				if accesses[name][accessType] {
					types = append(types, accessType)
				}
			}

			fmt.Printf("- Function %s accesses %s: %v\n",
				fn.Name.Name, name, types)
		}
	}
}
