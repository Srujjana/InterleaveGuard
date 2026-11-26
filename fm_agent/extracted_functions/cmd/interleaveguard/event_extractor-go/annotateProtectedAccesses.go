func annotateProtectedAccesses(function *FunctionModel) {
	for i := range function.Events {
		event := &function.Events[i]

		// Only shared-state accesses need protection annotations.
		if event.Kind != "READ" &&
			event.Kind != "WRITE" &&
			event.Kind != "READ_WRITE" {
			continue
		}

		var locks []string

		for j := range function.CriticalSections {
			section := &function.CriticalSections[j]

			// Check whether the access is inside this critical section.
			if event.pos > section.startPos &&
				event.pos < section.endPos {

				// Avoid duplicate lock names.
				alreadyAdded := false
				for _, lock := range locks {
					if lock == section.Lock {
						alreadyAdded = true
						break
					}
				}

				if !alreadyAdded {
					locks = append(locks, section.Lock)
				}

				// Record the resource protected by this section.
				found := false
				for _, resource := range section.Resources {
					if resource == event.Resource {
						found = true
						break
					}
				}

				if !found {
					section.Resources = append(
						section.Resources,
						event.Resource,
					)
				}
			}
		}

		event.ProtectedBy = strings.Join(locks, ",")
	}
}
