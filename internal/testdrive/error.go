package testdrive

import "fmt"

type UnexpectedStatusError struct {
	StatusCode   int
	ResponseBody []byte
}

func (u *UnexpectedStatusError) Error() string {
	return fmt.Sprintf("unexpected status code %d: %s", u.StatusCode, u.ResponseBody)
}
