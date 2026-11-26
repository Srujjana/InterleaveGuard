func analyzeDependencies(program *ProgramModel) {
	var dependencies []Dependency

	// Compare events within each function.
	// Source order establishes program-order relationships.
	for _, function := range program.Functions {
		events := function.Events

		for i := 0; i < len(events); i++ {
			for j := i + 1; j < len(events); j++ {
				dependencies = append(dependencies, Dependency{
					FromEventID: events[i].ID,
					ToEventID:   events[j].ID,
					Kind:        "PROGRAM_ORDER",
				})
			}
		}
	}

	// Identify function pairs that may execute concurrently,
	// including functions indirectly reachable from goroutine targets.
	concurrentPairs := buildConcurrentFunctionPairs(program)
	// Compare events only between potentially concurrent functions.
	for i := 0; i < len(program.Functions); i++ {
		for j := i + 1; j < len(program.Functions); j++ {

			functionA := program.Functions[i]
			functionB := program.Functions[j]

			if !concurrentPairs[functionA.Name][functionB.Name] {
				continue
			}

			for _, eventA := range functionA.Events {
				for _, eventB := range functionB.Events {

					if eventA.Resource == "" ||
						eventA.Resource != eventB.Resource {
						continue
					}
					// Both accesses use the same lock, so they cannot
					// execute inside their critical sections simultaneously.
					if eventA.ProtectedBy != "" &&
						eventA.ProtectedBy == eventB.ProtectedBy {
						continue
					}
					if !isWriteOperation(eventA.Kind) &&
						!isWriteOperation(eventB.Kind) {
						continue
					}

					dependencies = append(dependencies, Dependency{
						FromEventID: eventA.ID,
						ToEventID:   eventB.ID,
						Kind:        "SHARED_RESOURCE_CONFLICT",
						Resource:    eventA.Resource,
					})
				}
			}
		}
	}

	program.Dependencies = dependencies
}
