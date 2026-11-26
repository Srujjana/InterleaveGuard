func extractCriticalSectionsFromBody(
	body *ast.BlockStmt,
	info *types.Info,
	fset *token.FileSet,
) []CriticalSection {
	type lockOperation struct {
		kind     string
		resource string
		pos      token.Pos
		deferred bool
	}

	type lockStart struct {
		resource string
		pos      token.Pos
	}

	var operations []lockOperation

	// Record calls that are deferred.
	deferredCalls := make(map[token.Pos]bool)

	inspectFunctionBody(body, func(n ast.Node) bool {
		if stmt, ok := n.(*ast.DeferStmt); ok {
			deferredCalls[stmt.Call.Pos()] = true
		}
		return true
	})

	// Collect recognized mutex operations.
	inspectFunctionBody(body, func(n ast.Node) bool {
		call, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}

		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		kind := synchronizationMethod(selector, info)

		switch kind {
		case "MUTEX_Lock", "MUTEX_Unlock",
			"RW_MUTEX_Lock", "RW_MUTEX_Unlock",
			"RW_MUTEX_RLock", "RW_MUTEX_RUnlock":

			operations = append(operations, lockOperation{
				kind:     kind,
				resource: resourceName(selector.X),
				pos:      call.Pos(),
				deferred: deferredCalls[call.Pos()],
			})
		}

		return true
	})

	// Process operations in source order.
	sort.Slice(operations, func(i, j int) bool {
		return operations[i].pos < operations[j].pos
	})

	active := make(map[string][]lockStart)
	var deferredUnlocks []string
	var sections []CriticalSection

	for _, operation := range operations {
		switch operation.kind {
		case "MUTEX_Lock", "RW_MUTEX_Lock", "RW_MUTEX_RLock":
			active[operation.resource] = append(
				active[operation.resource],
				lockStart{
					resource: operation.resource,
					pos:      operation.pos,
				},
			)

		case "MUTEX_Unlock", "RW_MUTEX_Unlock", "RW_MUTEX_RUnlock":
			if operation.deferred {
				deferredUnlocks = append(
					deferredUnlocks,
					operation.resource,
				)
				continue
			}

			starts := active[operation.resource]
			if len(starts) == 0 {
				continue
			}

			start := starts[len(starts)-1]
			active[operation.resource] = starts[:len(starts)-1]

			sections = append(sections, CriticalSection{
				Lock:      start.resource,
				StartLine: fset.Position(start.pos).Line,
				EndLine:   fset.Position(operation.pos).Line,
				startPos:  start.pos,
				endPos:    operation.pos,
			})
		}
	}

	// Deferred unlocks execute when the function returns,
	// in last-in-first-out order.
	for i := len(deferredUnlocks) - 1; i >= 0; i-- {
		resource := deferredUnlocks[i]
		starts := active[resource]

		if len(starts) == 0 {
			continue
		}

		start := starts[len(starts)-1]
		active[resource] = starts[:len(starts)-1]

		sections = append(sections, CriticalSection{
			Lock:      start.resource,
			StartLine: fset.Position(start.pos).Line,
			EndLine:   fset.Position(body.Rbrace).Line,
			startPos:  start.pos,
			endPos:    body.Rbrace,
		})
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].StartLine < sections[j].StartLine
	})

	return sections
}
