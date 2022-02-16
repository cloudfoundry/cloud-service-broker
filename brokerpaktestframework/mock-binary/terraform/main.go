package main

import (
	"fmt"
	cp "github.com/otiai10/copy"
	"os"
	"path"
	"time"
)

var InvocationStore = ""

func main() {
	if InvocationStore == "" {
		panic("InvocationStore not set")
	}
	timestamp := fmt.Sprintf("%d", time.Now().UnixNano())
	targetDir := path.Join(InvocationStore, os.Args[1]+"-"+timestamp)
	os.Mkdir(targetDir, 0700)
	pwd, err := os.Getwd()
	if err != nil {
		panic(err.Error())
	}
	cp.Copy(pwd, targetDir)
}
