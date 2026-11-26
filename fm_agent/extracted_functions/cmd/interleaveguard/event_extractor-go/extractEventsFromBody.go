func extractEventsFromBody(
	functionName string,
	body *ast.BlockStmt,
	globals map[*types.Var]bool,
	info *types.Info,
	fset *token.FileSet,
	startID int,
) []Event {
	type positionedEvent struct {
		event Event
		pos   token.Pos
	}

	var collected []positionedEvent

	writes := make(map[token.Pos]string)
	synchronizationPositions := make(map[token.Pos]bool)

	// Identify writes to package-level variables.
	inspectFunctionBody(body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				id, ok := lhs.(*ast.Ident)
				if !ok {
					continue
				}

				variable, ok := info.Uses[id].(*types.Var)
				if !ok || !globals[variable] {
					continue
				}

				kind := "WRITE"
				if node.Tok != token.ASSIGN {
					kind = "READ_WRITE"
				}
				writes[id.Pos()] = kind
			}

		case *ast.IncDecStmt:
			id, ok := node.X.(*ast.Ident)
			if !ok {
				return true
			}

			variable, ok := info.Uses[id].(*types.Var)
			if ok && globals[variable] {
				writes[id.Pos()] = "READ_WRITE"
			}
		}

		return true
	})

	// Collect synchronization operations.
	deferredCalls := make(map[token.Pos]bool)

	inspectFunctionBody(body, func(n ast.Node) bool {
		if stmt, ok := n.(*ast.DeferStmt); ok {
			deferredCalls[stmt.Call.Pos()] = true
		}
		return true
	})

	inspectFunctionBody(body, func(n ast.Node) bool {
		addEvent := func(kind string, resource ast.Expr, pos token.Pos) {
			collected = append(collected, positionedEvent{
				event: Event{
					Function: functionName,
					Kind:     kind,
					Resource: resourceName(resource),
					Line:     fset.Position(pos).Line,
				},
				pos: pos,
			})

			if id, ok := resource.(*ast.Ident); ok {
				synchronizationPositions[id.Pos()] = true
			}
		}

		switch node := n.(type) {
		case *ast.CallExpr:
			if selector, ok := node.Fun.(*ast.SelectorExpr); ok {
				if kind := synchronizationMethod(selector, info); kind != "" {
					eventPos := node.Pos()

					if deferredCalls[node.Pos()] {
						eventPos = body.Rbrace
					}

					addEvent(kind, selector.X, eventPos)
				}
			}

			if id, ok := node.Fun.(*ast.Ident); ok &&
				id.Name == "close" &&
				info.Uses[id] == types.Universe.Lookup("close") &&
				len(node.Args) == 1 &&
				isChannel(node.Args[0], info) {
				addEvent("CHANNEL_CLOSE", node.Args[0], node.Pos())
			}

		case *ast.SendStmt:
			if isChannel(node.Chan, info) {
				addEvent("CHANNEL_SEND", node.Chan, node.Pos())
			}

		case *ast.UnaryExpr:
			if node.Op == token.ARROW && isChannel(node.X, info) {
				addEvent("CHANNEL_RECEIVE", node.X, node.Pos())
			}

		case *ast.SelectStmt:
			addEvent("SELECT", nil, node.Pos())
		}

		return true
	})

	// Collect accesses to package-level variables.
	inspectFunctionBody(body, func(n ast.Node) bool {
		id, ok := n.(*ast.Ident)
		if !ok || synchronizationPositions[id.Pos()] {
			return true
		}

		variable, ok := info.Uses[id].(*types.Var)
		if !ok || !globals[variable] {
			return true
		}

		kind := "READ"
		if writeKind, exists := writes[id.Pos()]; exists {
			kind = writeKind
		}

		collected = append(collected, positionedEvent{
			event: Event{
				Function: functionName,
				Kind:     kind,
				Resource: variable.Name(),
				Line:     fset.Position(id.Pos()).Line,
				pos:      id.Pos(),
			},
			pos: id.Pos(),
		})

		return true
	})

	// Preserve exact source order.
	sort.Slice(collected, func(i, j int) bool {
		return collected[i].pos < collected[j].pos
	})

	events := make([]Event, 0, len(collected))

	for i, item := range collected {
		item.event.ID = startID + i
		events = append(events, item.event)
	}

	return events

}
