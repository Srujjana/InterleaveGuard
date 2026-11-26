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
