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
