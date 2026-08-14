package migration

import "github.com/google/uuid"

// derefString is a helper function to safely dereference a string pointer.
func derefString(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}

// parseNullableUUID parses a nullable UUID pointer into a UUID.
// It returns uuid.Nil if the pointer is nil or the string is empty.
func parseNullableUUID(s *string) (uuid.UUID, error) {
	if s == nil || *s == "" {
		return uuid.Nil, nil
	}
	return uuid.Parse(*s)
}