package main

import (
	"fmt"
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"sort"
)

// Represents one concurrency-related operation.
type Event struct {
	ID          int    `json:"id"`
	Function    string `json:"function"`
	Kind        string `json:"kind"`
	Resource    string `json:"resource,omitempty"`
	ProtectedBy string `json:"protected_by,omitempty"`
	Line        int    `json:"line"`
	pos         token.Pos
}

// Represents a region protected by a synchronization lock.
type CriticalSection struct {
	Lock      string   `json:"lock"`
	StartLine int      `json:"start_line"`
	EndLine   int      `json:"end_line"`
	Resources []string `json:"resources"`

	startPos token.Pos
	endPos   token.Pos
}

// Represents all events found inside a function.
type FunctionModel struct {
	Name             string            `json:"name"`
	Events           []Event           `json:"events"`
	CriticalSections []CriticalSection `json:"critical_sections"`
}

// Represents a dependency between two operations.
type Dependency struct {
	FromEventID int    `json:"from_event_id"`
	ToEventID   int    `json:"to_event_id"`
	Kind        string `json:"kind"`
	Resource    string `json:"resource,omitempty"`
}

// Represents a goroutine launch relationship.
type GoroutineLaunch struct {
	Caller string `json:"caller"`
	Target string `json:"target"`
}

// Represents a direct function-call relationship.
type CallEdge struct {
	Caller string `json:"caller"`
	Callee string `json:"callee"`
}

type GoroutineExecution struct {
	Caller             string   `json:"caller"`
	Entry              string   `json:"entry"`
	ReachableFunctions []string `json:"reachable_functions"`
}

// Represents the analyzed Go program.
type ProgramModel struct {
	File                string               `json:"file"`
	Functions           []FunctionModel      `json:"functions"`
	Dependencies        []Dependency         `json:"dependencies"`
	GoroutineLaunches   []GoroutineLaunch    `json:"goroutine_launches"`
	CallGraph           []CallEdge           `json:"call_graph"`
	GoroutineExecutions []GoroutineExecution `json:"goroutine_executions"`
}

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

func getGoroutineTarget(
	call *ast.CallExpr,
	caller string,
	anonymousIndex int,
) (string, *ast.FuncLit) {
	if literal, ok := call.Fun.(*ast.FuncLit); ok {
		name := fmt.Sprintf("%s$goroutine%d", caller, anonymousIndex)
		return name, literal
	}

	return getCallName(call.Fun), nil
}

// Collect package-level variables.
// Collect package-level variables by their actual declarations.
func getGlobalVariables(
	file *ast.File,
	info *types.Info,
) map[*types.Var]bool {
	globals := make(map[*types.Var]bool)

	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}

		for _, spec := range gen.Specs {
			valueSpec := spec.(*ast.ValueSpec)

			for _, name := range valueSpec.Names {
				if variable, ok := info.Defs[name].(*types.Var); ok {
					globals[variable] = true
				}
			}
		}
	}

	return globals
}

// Returns the package-level variable an identifier refers to, or nil.
func asGlobal(
	id *ast.Ident,
	globals map[*types.Var]bool,
	info *types.Info,
) *types.Var {
	if v, ok := info.Uses[id].(*types.Var); ok && globals[v] {
		return v
	}
	return nil
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

	// Resolve identifiers to their actual declarations.
	info := &types.Info{
		Defs:       make(map[*ast.Ident]types.Object),
		Uses:       make(map[*ast.Ident]types.Object),
		Types:      make(map[ast.Expr]types.TypeAndValue),
		Selections: make(map[*ast.SelectorExpr]*types.Selection),
	}

	config := types.Config{
		Importer: importer.Default(),
		//importer.ForCompiler(fset, "source", nil)
	}

	_, err = config.Check(file.Name.Name, fset, []*ast.File{file}, info)
	if err != nil {
		fmt.Println("Type-checking error:", err)
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
	// Identify named and anonymous goroutine launches.
	fmt.Println("\nGoroutine targets:")

	var goroutineLaunches []GoroutineLaunch
	anonymousGoroutines := make(map[string]*ast.FuncLit)
	anonymousIndex := 0

	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			goStmt, ok := n.(*ast.GoStmt)
			if !ok {
				return true
			}

			target, literal := getGoroutineTarget(
				goStmt.Call,
				fn.Name.Name,
				anonymousIndex+1,
			)

			if literal != nil {
				anonymousIndex++
				anonymousGoroutines[target] = literal
			}

			fmt.Printf("- %s launches %s()\n", fn.Name.Name, target)

			goroutineLaunches = append(goroutineLaunches, GoroutineLaunch{
				Caller: fn.Name.Name,
				Target: target,
			})

			return true
		})
	}

	// Identify package-level shared variables.
	globals := getGlobalVariables(file, info)

	// Build the structured program model.
	program := ProgramModel{
		File:              filePath,
		GoroutineLaunches: goroutineLaunches,
	}
	// Extract direct function-call relationships.
	program.CallGraph = extractCallGraph(file, info)

	// Display functions reachable from each goroutine launch.
	fmt.Println("\nReachable functions per goroutine:")

	for _, launch := range program.GoroutineLaunches {
		reachable := reachableFunctions(launch.Target, program.CallGraph)

		fmt.Printf("\nGoroutine launched by %s -> %s:\n",
			launch.Caller, launch.Target)

		for _, function := range reachable {
			fmt.Println("  ", function)
		}
	}

	// Display direct function-call relationships.
	fmt.Println("\nFunction call graph:")

	for _, edge := range program.CallGraph {
		fmt.Printf("%s -> %s\n", edge.Caller, edge.Callee)
	}

	eventID := 1

	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		events := extractEvents(fn, globals, info, fset, eventID)
		eventID += len(events)

		criticalSections := extractCriticalSections(fn, info, fset)
		functionModel := FunctionModel{
			Name:             fn.Name.Name,
			Events:           events,
			CriticalSections: criticalSections,
		}

		annotateProtectedAccesses(&functionModel)

		program.Functions = append(program.Functions, functionModel)
	}

	// Extract events from anonymous goroutine bodies.
	var anonymousNames []string

	for name := range anonymousGoroutines {
		anonymousNames = append(anonymousNames, name)
	}

	sort.Strings(anonymousNames)

	for _, name := range anonymousNames {
		literal := anonymousGoroutines[name]

		events := extractEventsFromBody(
			name,
			literal.Body,
			globals,
			info,
			fset,
			eventID,
		)
		eventID += len(events)

		criticalSections := extractCriticalSectionsFromBody(
			literal.Body,
			info,
			fset,
		)

		functionModel := FunctionModel{
			Name:             name,
			Events:           events,
			CriticalSections: criticalSections,
		}

		annotateProtectedAccesses(&functionModel)

		program.Functions = append(program.Functions, functionModel)
	}

	// Analyze dependencies between extracted operations.
	analyzeDependencies(&program)

	fmt.Println("\nStructured concurrency events:")

	fmt.Println("\nCritical sections:")

	for _, function := range program.Functions {
		for _, section := range function.CriticalSections {
			fmt.Printf(
				"Function %s | Lock: %s | Start: Line %d | End: Line %d | Resources: %v\n",
				function.Name,
				section.Lock,
				section.StartLine,
				section.EndLine,
				section.Resources,
			)
		}
	}

	for _, function := range program.Functions {
		for _, event := range function.Events {
			fmt.Printf(
				"Event %d | %s | %s | %s | Line %d | Protected by: %s\n",
				event.ID,
				event.Function,
				event.Kind,
				event.Resource,
				event.Line,
				event.ProtectedBy,
			)
		}
	}

	fmt.Println("\nOperation dependencies:")

	for _, dependency := range program.Dependencies {
		fmt.Printf(
			"Event %d -> Event %d | %s | Resource: %s\n",
			dependency.FromEventID,
			dependency.ToEventID,
			dependency.Kind,
			dependency.Resource,
		)
	}

	fmt.Println("\nPackage-level variables:")

	globalNames := make([]string, 0, len(globals))
	for variable := range globals {
		globalNames = append(globalNames, variable.Name())
	}
	sort.Strings(globalNames)

	for _, name := range globalNames {
		fmt.Println("-", name)
	}
}
