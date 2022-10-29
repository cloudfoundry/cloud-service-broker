package steps

func Sequentially(callbacks ...func() error) error {
	for _, c := range callbacks {
		if err := c(); err != nil {
			return err
		}
	}
	return nil
}
