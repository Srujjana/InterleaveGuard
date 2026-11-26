func isWriteOperation(kind string) bool {
	return kind == "WRITE" || kind == "READ_WRITE"
}
