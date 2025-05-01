package brokerchapi

import "time"

type token struct {
	value  string
	expiry time.Time
}

func (t token) valid() bool {
	return t.value != "" && time.Now().Before(t.expiry)
}
