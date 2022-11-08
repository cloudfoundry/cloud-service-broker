// Package local is an experimental mimic for the "cf create-service" command
package local

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

func nameToID(name string) string {
	if len(name) > 16 {
		log.Fatal("name too long")
	}

	for len(name) < 16 {
		name = name + "."
	}

	h := []byte(fmt.Sprintf("%x", name))
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}

func idToName(id string) (result string) {
	h := strings.Join(strings.Split(id, "-"), "")
	if len(h)%2 != 0 {
		log.Fatalf("expected even length")
	}
	for i := 0; i < len(h); i += 2 {
		b := h[i : i+2]
		r, err := strconv.ParseUint(b, 16, 8)
		if err != nil {
			log.Fatal(err)
		}
		result = fmt.Sprintf("%s%s", result, []byte{byte(r)})
	}
	return strings.Trim(result, ".")
}
