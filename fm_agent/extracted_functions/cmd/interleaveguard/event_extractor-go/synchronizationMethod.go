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
