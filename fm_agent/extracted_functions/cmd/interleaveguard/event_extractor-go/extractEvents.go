func extractEvents(
	fn *ast.FuncDecl,
	globals map[*types.Var]bool,
	info *types.Info,
	fset *token.FileSet,
	startID int,
) []Event {
	return extractEventsFromBody(
		fn.Name.Name,
		fn.Body,
		globals,
		info,
		fset,
		startID,
	)
}
