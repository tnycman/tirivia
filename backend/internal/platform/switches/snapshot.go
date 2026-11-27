package switches

// Snapshot is an immutable view of all switch values.
type Snapshot struct {
	Global  map[Key]bool
	PerShow map[string]map[Key]bool
}

// Enabled resolves show scope, then global, then the default.
func (s Snapshot) Enabled(key Key, showID string) bool {
	if showID != "" {
		if v, ok := s.PerShow[showID][key]; ok {
			return v
		}
	}
	if v, ok := s.Global[key]; ok {
		return v
	}
	return defaults[key]
}
