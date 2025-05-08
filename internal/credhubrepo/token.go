package credhubrepo

import "time"

type token struct {
	value  string
	expiry time.Time
}

func (t token) expired() bool {
	return t.value == "" || time.Now().After(t.expiry)
}
