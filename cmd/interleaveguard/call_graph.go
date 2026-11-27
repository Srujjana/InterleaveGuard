package main

import (
	"go/ast"
	"go/types"
)

// Extract direct calls between functions declared in the analyzed file.

func extractCallGraph(
	file *ast.File,
	info *types.Info,
	anonymousGoroutines map[string]*ast.FuncLit,
) []CallEdge {

	// Map declared functions and methods to their names.
	declaredFunctions := make(map[*types.Func]string)

	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		function, ok := info.Defs[fn.Name].(*types.Func)
		if !ok {
			continue
		}

		declaredFunctions[function] = fn.Name.Name
	}

	var callGraph []CallEdge

	// Find calls made inside a function body.
	inspectCalls := func(caller string, body *ast.BlockStmt) {
		ast.Inspect(body, func(n ast.Node) bool {
			// Anonymous goroutines are analyzed separately.
			if _, ok := n.(*ast.FuncLit); ok {
				return false
			}

			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			var calledFunction *types.Func

			switch expression := call.Fun.(type) {
			case *ast.Ident:
				calledFunction, _ = info.Uses[expression].(*types.Func)

			case *ast.SelectorExpr:
				if selection := info.Selections[expression]; selection != nil {
					calledFunction, _ = selection.Obj().(*types.Func)
				} else {
					calledFunction, _ = info.Uses[expression.Sel].(*types.Func)
				}
			}

			callee, exists := declaredFunctions[calledFunction]
			if !exists {
				return true
			}

			callGraph = append(callGraph, CallEdge{
				Caller: caller,
				Callee: callee,
			})

			return true
		})
	}

	// Analyze normal functions and methods.
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}

		inspectCalls(fn.Name.Name, fn.Body)
	}

	// Analyze anonymous goroutine bodies separately.
	for name, literal := range anonymousGoroutines {
		inspectCalls(name, literal.Body)
	}

	return callGraph
}

// Find all functions reachable from a given starting function.
func reachableFunctions(start string, callGraph []CallEdge) []string {
	// Build an adjacency list from the call graph.
	adjacency := make(map[string][]string)

	for _, edge := range callGraph {
		adjacency[edge.Caller] = append(
			adjacency[edge.Caller],
			edge.Callee,
		)
	}

	visited := make(map[string]bool)
	var reachable []string

	// DFS traversal.
	var visit func(string)

	visit = func(function string) {
		if visited[function] {
			return
		}

		visited[function] = true
		reachable = append(reachable, function)

		for _, callee := range adjacency[function] {
			visit(callee)
		}
	}

	visit(start)

	return reachable
}

// Add a potentially concurrent relationship between two functions.
func addConcurrentPair(
	pairs map[string]map[string]bool,
	first string,
	second string,
) {
	if pairs[first] == nil {
		pairs[first] = make(map[string]bool)
	}

	if pairs[second] == nil {
		pairs[second] = make(map[string]bool)
	}

	pairs[first][second] = true
	pairs[second][first] = true
}

// Build function pairs that may execute concurrently.
func buildConcurrentFunctionPairs(
	program *ProgramModel,
) map[string]map[string]bool {

	pairs := make(map[string]map[string]bool)

	// Store the reachable functions for each goroutine launch,
	// grouped by the function that launches them.
	launchesByCaller := make(map[string][][]string)

	for _, launch := range program.GoroutineLaunches {
		reachable := reachableFunctions(
			launch.Target,
			program.CallGraph,
		)

		launchesByCaller[launch.Caller] = append(
			launchesByCaller[launch.Caller],
			reachable,
		)

		// The launching function may continue executing
		// while the new goroutine runs.
		for _, function := range reachable {
			// The launching function may continue while the goroutine runs.
			addConcurrentPair(pairs, launch.Caller, function)

			// Multiple goroutines may execute the same function concurrently.
			addConcurrentPair(pairs, function, function)
		}
	}

	// Different goroutines launched by the same function
	// may execute concurrently with one another.
	for _, launches := range launchesByCaller {
		for i := 0; i < len(launches); i++ {
			for j := i + 1; j < len(launches); j++ {
				for _, first := range launches[i] {
					for _, second := range launches[j] {
						addConcurrentPair(pairs, first, second)
					}
				}
			}
		}
	}

	return pairs
}
