func extractCriticalSections(
	fn *ast.FuncDecl,
	info *types.Info,
	fset *token.FileSet,
) []CriticalSection {
	return extractCriticalSectionsFromBody(
		fn.Body,
		info,
		fset,
	)
}
