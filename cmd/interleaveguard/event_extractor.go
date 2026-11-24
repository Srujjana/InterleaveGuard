package main

import (
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"
)

// Get a readable name for a resource expression.
func resourceName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		return resourceName(e.X) + "." + e.Sel.Name
	case *ast.StarExpr:
		return "*" + resourceName(e.X)
	case *ast.IndexExpr:
		return resourceName(e.X) + "[...]"
	}
	return "<expression>"
}

// Identify synchronization methods using their actual Go receiver types.
func synchronizationMethod(
	selector *ast.SelectorExpr,
	info *types.Info,
) string {
	selection := info.Selections[selector]
	if selection == nil {
		return ""
	}

	receiverType := info.TypeOf(selector.X)
	if receiverType == nil {
		return ""
	}

	receiverType = types.Unalias(receiverType)

	if pointer, ok := receiverType.(*types.Pointer); ok {
		receiverType = types.Unalias(pointer.Elem())
	}

	named, ok := receiverType.(*types.Named)
	if !ok || named.Obj().Pkg() == nil {
		return ""
	}

	pkg := named.Obj().Pkg().Path()
	typeName := named.Obj().Name()
	method := selection.Obj().Name()

	if pkg != "sync" {
		return ""
	}

	switch typeName {
	case "Mutex":
		switch method {
		case "Lock", "Unlock", "TryLock":
			return "MUTEX_" + method
		}

	case "RWMutex":
		switch method {
		case "Lock", "Unlock", "RLock", "RUnlock",
			"TryLock", "TryRLock":
			return "RW_MUTEX_" + method
		}

	case "WaitGroup":
		switch method {
		case "Add", "Done", "Wait":
			return "WAITGROUP_" + method
		}
	}

	return ""
}

// Check whether an expression represents a channel.
func isChannel(expr ast.Expr, info *types.Info) bool {
	t := info.TypeOf(expr)
	if t == nil {
		return false
	}

	_, ok := types.Unalias(t).Underlying().(*types.Chan)
	return ok
}

// Extract concurrency-related events from a function.
func extractEvents(
	fn *ast.FuncDecl,
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
	ast.Inspect(fn.Body, func(n ast.Node) bool {
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

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if stmt, ok := n.(*ast.DeferStmt); ok {
			deferredCalls[stmt.Call.Pos()] = true
		}
		return true
	})

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		addEvent := func(kind string, resource ast.Expr, pos token.Pos) {
			collected = append(collected, positionedEvent{
				event: Event{
					Function: fn.Name.Name,
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
						eventPos = fn.Body.Rbrace
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
	ast.Inspect(fn.Body, func(n ast.Node) bool {
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
				Function: fn.Name.Name,
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

// Extract lexically matched critical sections from a function.
func extractCriticalSections(
	fn *ast.FuncDecl,
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

	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if stmt, ok := n.(*ast.DeferStmt); ok {
			deferredCalls[stmt.Call.Pos()] = true
		}
		return true
	})

	// Collect recognized mutex operations.
	ast.Inspect(fn.Body, func(n ast.Node) bool {
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
			EndLine:   fset.Position(fn.Body.Rbrace).Line,
			startPos:  start.pos,
			endPos:    fn.Body.Rbrace,
		})
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].StartLine < sections[j].StartLine
	})

	return sections
}

// Associate shared-variable accesses with locks held at their source positions.
func annotateProtectedAccesses(function *FunctionModel) {
	for i := range function.Events {
		event := &function.Events[i]

		// Only shared-state accesses need protection annotations.
		if event.Kind != "READ" &&
			event.Kind != "WRITE" &&
			event.Kind != "READ_WRITE" {
			continue
		}

		var locks []string

		for j := range function.CriticalSections {
			section := &function.CriticalSections[j]

			// Check whether the access is inside this critical section.
			if event.pos > section.startPos &&
				event.pos < section.endPos {

				// Avoid duplicate lock names.
				alreadyAdded := false
				for _, lock := range locks {
					if lock == section.Lock {
						alreadyAdded = true
						break
					}
				}

				if !alreadyAdded {
					locks = append(locks, section.Lock)
				}

				// Record the resource protected by this section.
				found := false
				for _, resource := range section.Resources {
					if resource == event.Resource {
						found = true
						break
					}
				}

				if !found {
					section.Resources = append(
						section.Resources,
						event.Resource,
					)
				}
			}
		}

		event.ProtectedBy = strings.Join(locks, ",")
	}
}
