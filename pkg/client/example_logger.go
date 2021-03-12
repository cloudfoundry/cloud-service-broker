package client

import (
	"fmt"
	"log"
)

type exampleLogger struct {
	id string
}

func newExampleLogger(id string) *exampleLogger {
	return &exampleLogger{id: id}
}

func (l *exampleLogger) Println(s string) {
	log.Println(l.format(s))
}

func (l *exampleLogger) Printf(format string, v ...interface{}) {
	log.Printf(l.format(format), v...)
}

func (l *exampleLogger) Fatalf(format string, v ...interface{}) {
	log.Fatalf(l.format(format), v...)
}

func (l *exampleLogger) format(format string) string {
	return fmt.Sprintf("(id=%s): %s", l.id, format)
}
