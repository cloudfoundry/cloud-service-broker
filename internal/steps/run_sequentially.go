// Package steps implements a "stepper" which steps through a list of callbacks, running each one sequentially
package steps

// RunSequentially will sequentially call the specified callback functions, stopping on the first one that returns an error
// By using the common error handling, code can easier to read and less prone to typos in the error handling
func RunSequentially(callbacks ...func() error) error {
	for _, c := range callbacks {
		if err := c(); err != nil {
			return err
		}
	}
	return nil
}
