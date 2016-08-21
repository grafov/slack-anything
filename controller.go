package main

import (
	"time"
)

func runController(configName string) chan string {
	logErr, logInfo := runLogger()
	runConfigurer(logErr, logInfo, configName) // XXX
	go controller()
	return logInfo
}

func controller() {
	for {
		time.Sleep(reinitTime)
	}
}
