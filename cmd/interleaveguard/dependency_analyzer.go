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

	// Identify function pairs that may execute concurrently,
	// including functions indirectly reachable from goroutine targets.
	concurrentPairs := buildConcurrentFunctionPairs(program)

	// Compare events between potentially concurrent functions.
	// Also compare a function against itself when multiple goroutines
	// may execute that function concurrently.
	for i := 0; i < len(program.Functions); i++ {
		for j := i; j < len(program.Functions); j++ {

			functionA := program.Functions[i]
			functionB := program.Functions[j]

			if !concurrentPairs[functionA.Name][functionB.Name] {
				continue
			}

			for eventIndexA, eventA := range functionA.Events {
				for eventIndexB, eventB := range functionB.Events {

					// Avoid comparing an event with earlier events
					// when both belong to the same function.
					if i == j && eventIndexB < eventIndexA {
						continue
					}

					if eventA.Resource == "" ||
						eventA.Resource != eventB.Resource {
						continue
					}

					// Same lock protection prevents simultaneous
					// conflicting access to this resource.
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
