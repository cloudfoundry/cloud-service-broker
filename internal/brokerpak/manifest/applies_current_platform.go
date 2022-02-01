package manifest

// AppliesToCurrentPlatform returns true if the one of the platforms in the
// manifest match the current GOOS and GOARCH.
func (m *Manifest) AppliesToCurrentPlatform() bool {
	for _, platform := range m.Platforms {
		if platform.MatchesCurrent() {
			return true
		}
	}

	return false
}
