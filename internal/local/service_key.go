package local

import (
	"encoding/json"
	"fmt"
	"log"
)

func ServiceKey(serviceInstanceName, serviceKeyName string) {
	credentials, err := store().GetServiceBindingCredentials(nameToID(serviceKeyName), nameToID(serviceInstanceName))
	if err != nil {
		log.Fatal(err)
	}
	bytes, _ := json.Marshal(credentials.Credentials)
	fmt.Printf("\nService Key: %s\n", bytes)
}
