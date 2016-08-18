package main

import (
	"time"
)

func controller(configName string) {
	logErr, logInfo := initLogger()
	initConfigurer(logErr, logInfo, configName) // XXX
	for {

		time.Sleep(reinitTime)
	}
}
