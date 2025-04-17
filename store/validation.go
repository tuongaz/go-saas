package store

// ValidTableName checks if a table name is valid to prevent SQL injection
// It should only contain alphanumeric characters, underscores, and should not start with a number
func ValidTableName(table string) bool {
	if len(table) == 0 {
		return false
	}

	// First character should be a letter or underscore
	if !(('a' <= table[0] && table[0] <= 'z') ||
		('A' <= table[0] && table[0] <= 'Z') ||
		table[0] == '_') {
		return false
	}

	// Rest of the characters should be alphanumeric or underscore
	for i := 1; i < len(table); i++ {
		c := table[i]
		if !(('a' <= c && c <= 'z') ||
			('A' <= c && c <= 'Z') ||
			('0' <= c && c <= '9') ||
			c == '_') {
			return false
		}
	}

	return true
}

// ValidIdentifierName checks if a database identifier (column/field name) is valid
func ValidIdentifierName(name string) bool {
	if len(name) == 0 {
		return false
	}

	// First character should be a letter or underscore
	if !(('a' <= name[0] && name[0] <= 'z') ||
		('A' <= name[0] && name[0] <= 'Z') ||
		name[0] == '_') {
		return false
	}

	// Rest of the characters should be alphanumeric or underscore
	for i := 1; i < len(name); i++ {
		c := name[i]
		if !(('a' <= c && c <= 'z') ||
			('A' <= c && c <= 'Z') ||
			('0' <= c && c <= '9') ||
			c == '_') {
			return false
		}
	}

	return true
}
