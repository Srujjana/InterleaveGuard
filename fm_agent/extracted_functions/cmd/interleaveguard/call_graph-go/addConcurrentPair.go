func addConcurrentPair(
	pairs map[string]map[string]bool,
	first string,
	second string,
) {
	if pairs[first] == nil {
		pairs[first] = make(map[string]bool)
	}

	if pairs[second] == nil {
		pairs[second] = make(map[string]bool)
	}

	pairs[first][second] = true
	pairs[second][first] = true
}
