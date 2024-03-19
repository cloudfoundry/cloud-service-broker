package main

import (
	"fmt"
	"os"
	"path"
	"time"

	cp "github.com/otiai10/copy"
)

var InvocationStore = ""

func main() {
	if InvocationStore == "" {
		panic("InvocationStore not set")
	}
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	targetDir := path.Join(InvocationStore, os.Args[1]+"-"+timestamp)
	_ = os.Mkdir(targetDir, 0700)
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	_ = cp.Copy(pwd, targetDir)
	responseTFPath := path.Join(InvocationStore, "mock_tf_state.json")
	if _, err := os.Stat(responseTFPath); err == nil {
		_ = cp.Copy(responseTFPath, path.Join(pwd, "terraform.tfstate"))
	}
}
