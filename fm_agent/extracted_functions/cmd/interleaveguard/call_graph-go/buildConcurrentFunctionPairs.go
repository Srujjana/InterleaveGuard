func buildConcurrentFunctionPairs(
	program *ProgramModel,
) map[string]map[string]bool {

	pairs := make(map[string]map[string]bool)

	// Store the reachable functions for each goroutine launch,
	// grouped by the function that launches them.
	launchesByCaller := make(map[string][][]string)

	for _, launch := range program.GoroutineLaunches {
		reachable := reachableFunctions(
			launch.Target,
			program.CallGraph,
		)

		launchesByCaller[launch.Caller] = append(
			launchesByCaller[launch.Caller],
			reachable,
		)

		// The launching function may continue executing
		// while the new goroutine runs.
		for _, function := range reachable {
			addConcurrentPair(pairs, launch.Caller, function)
		}
	}

	// Different goroutines launched by the same function
	// may execute concurrently with one another.
	for _, launches := range launchesByCaller {
		for i := 0; i < len(launches); i++ {
			for j := i + 1; j < len(launches); j++ {
				for _, first := range launches[i] {
					for _, second := range launches[j] {
						addConcurrentPair(pairs, first, second)
					}
				}
			}
		}
	}

	return pairs
}
