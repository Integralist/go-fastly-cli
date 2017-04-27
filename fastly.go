package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func main() {
	logger := log.WithFields(log.Fields{
		"app":  "go-fastly-cli",
		"type": "cli",
	})
	logger.Debug("this is my debug log message")
	logger.Info("this is my info log message")
	fmt.Println("hello from fastly")
}
