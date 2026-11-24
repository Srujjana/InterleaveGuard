package main

// Check whether an event modifies a resource.
func isWriteOperation(kind string) bool {
	return kind == "WRITE" || kind == "READ_WRITE"
}

// Analyze dependencies between operations.
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

	// Compare events across different functions.
	// Identify function pairs that may execute concurrently.
	concurrentPairs := make(map[string]bool)

	addConcurrentPair := func(a, b string) {
		if a == b {
			return
		}

		// Normalize the pair so A-B and B-A are equivalent.
		if a > b {
			a, b = b, a
		}

		concurrentPairs[a+"|"+b] = true
	}

	// A goroutine may execute concurrently with its caller.
	for _, launch := range program.GoroutineLaunches {
		addConcurrentPair(launch.Caller, launch.Target)
	}

	// Goroutines launched by the same caller may overlap.
	for i := 0; i < len(program.GoroutineLaunches); i++ {
		for j := i + 1; j < len(program.GoroutineLaunches); j++ {
			first := program.GoroutineLaunches[i]
			second := program.GoroutineLaunches[j]

			if first.Caller == second.Caller {
				addConcurrentPair(first.Target, second.Target)
			}
		}
	}

	// Compare events only between potentially concurrent functions.
	for i := 0; i < len(program.Functions); i++ {
		for j := i + 1; j < len(program.Functions); j++ {

			functionA := program.Functions[i]
			functionB := program.Functions[j]

			pair := functionA.Name + "|" + functionB.Name
			reversePair := functionB.Name + "|" + functionA.Name

			if !concurrentPairs[pair] && !concurrentPairs[reversePair] {
				continue
			}

			for _, eventA := range functionA.Events {
				for _, eventB := range functionB.Events {

					if eventA.Resource == "" ||
						eventA.Resource != eventB.Resource {
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
