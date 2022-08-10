package createservice

import (
	"fmt"
)

func NameToID(name string) string {
	if len(name) > 16 {
		panic("name too long")
	}

	for len(name) < 16 {
		name = name + "."
	}

	h := []byte(fmt.Sprintf("%x", name))
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}
