func getGoroutineTarget(
	call *ast.CallExpr,
	caller string,
	anonymousIndex int,
) (string, *ast.FuncLit) {
	if literal, ok := call.Fun.(*ast.FuncLit); ok {
		name := fmt.Sprintf("%s$goroutine%d", caller, anonymousIndex)
		return name, literal
	}

	return getCallName(call.Fun), nil
}
