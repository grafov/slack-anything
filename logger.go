package main

import (
	"fmt"
	"os"
	"time"
)

func runLogger() (chan string, chan string) {
	logErr := make(chan string, 32)
	logInfo := make(chan string, 32)
	go logger(logErr, logInfo)
	return logErr, logInfo
}

func logger(logErr chan string, logInfo chan string) {
	for {
		select {
		case record := <-logInfo:
			fmt.Fprintf(os.Stderr, "%s %s\n", time.Now(), record)
		case record := <-logErr:
			fmt.Fprintf(os.Stdout, "%s %s\n", time.Now(), record)
		}
	}
}
