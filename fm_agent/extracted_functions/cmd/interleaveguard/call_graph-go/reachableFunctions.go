func reachableFunctions(start string, callGraph []CallEdge) []string {
	// Build an adjacency list from the call graph.
	adjacency := make(map[string][]string)

	for _, edge := range callGraph {
		adjacency[edge.Caller] = append(
			adjacency[edge.Caller],
			edge.Callee,
		)
	}

	visited := make(map[string]bool)
	var reachable []string

	// DFS traversal.
	var visit func(string)

	visit = func(function string) {
		if visited[function] {
			return
		}

		visited[function] = true
		reachable = append(reachable, function)

		for _, callee := range adjacency[function] {
			visit(callee)
		}
	}

	visit(start)

	return reachable
}
