package local

import (
	"fmt"
	"log"
)

func ServiceKeys(serviceInstanceName string) {
	keyGUIDs, err := store().GetServiceBindingIDsForServiceInstance(nameToID(serviceInstanceName))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Name")
	fmt.Println("----")
	for _, guid := range keyGUIDs {
		fmt.Println(idToName(guid))
	}
}
