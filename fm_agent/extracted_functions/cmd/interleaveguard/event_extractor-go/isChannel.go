func isChannel(expr ast.Expr, info *types.Info) bool {
	t := info.TypeOf(expr)
	if t == nil {
		return false
	}

	_, ok := types.Unalias(t).Underlying().(*types.Chan)
	return ok
}
