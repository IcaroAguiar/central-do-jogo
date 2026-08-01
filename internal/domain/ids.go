package domain

// ID is a stable opaque identifier for public and persisted entities.
type ID string

// String returns the raw identifier.
func (id ID) String() string {
	return string(id)
}

// Empty reports whether the identifier is unset.
func (id ID) Empty() bool {
	return id == ""
}
