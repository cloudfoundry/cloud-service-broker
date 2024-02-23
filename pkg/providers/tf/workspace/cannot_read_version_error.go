package workspace

type CannotReadVersionError struct {
	message string
}

func (s CannotReadVersionError) Error() string {
	return s.message
}
