func getCallName(expr ast.Expr) string {
	switch fn := expr.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		return getCallName(fn.X) + "." + fn.Sel.Name
	}
	return "<unknown>"
}
