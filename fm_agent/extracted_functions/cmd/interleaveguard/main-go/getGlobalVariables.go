func getGlobalVariables(
	file *ast.File,
	info *types.Info,
) map[*types.Var]bool {
	globals := make(map[*types.Var]bool)

	for _, declaration := range file.Decls {
		gen, ok := declaration.(*ast.GenDecl)
		if !ok || gen.Tok != token.VAR {
			continue
		}

		for _, spec := range gen.Specs {
			valueSpec := spec.(*ast.ValueSpec)

			for _, name := range valueSpec.Names {
				if variable, ok := info.Defs[name].(*types.Var); ok {
					globals[variable] = true
				}
			}
		}
	}

	return globals
}
