func extractCallGraph(
	file *ast.File,
	info *types.Info,
) []CallEdge {

	// Map actual function declarations to their names.
	declaredFunctions := make(map[*types.Func]string)

	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Recv != nil {
			continue
		}

		function, ok := info.Defs[fn.Name].(*types.Func)
		if !ok {
			continue
		}

		declaredFunctions[function] = fn.Name.Name
	}

	var callGraph []CallEdge

	// Inspect calls inside each declared function.
	for _, declaration := range file.Decls {
		fn, ok := declaration.(*ast.FuncDecl)
		if !ok || fn.Body == nil || fn.Recv != nil {
			continue
		}

		caller := fn.Name.Name

		ast.Inspect(fn.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Resolve the called function using Go type information.
			var calledFunction *types.Func

			switch expression := call.Fun.(type) {
			case *ast.Ident:
				calledFunction, _ = info.Uses[expression].(*types.Func)

			case *ast.SelectorExpr:
				if selection := info.Selections[expression]; selection != nil {
					calledFunction, _ = selection.Obj().(*types.Func)
				}
			}

			// Keep only functions declared in this analyzed file.
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

	return callGraph
}
