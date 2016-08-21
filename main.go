package main

import (
	"flag"
	"os"
	"os/signal"
	"time"
)

const (
	reinitTime = 3 * time.Second
)

// registry for workers that need finalization before app exit
var registry []chan bool

func main() {
	configName := flag.String("config", ".config/slack-anything", "configuration file")
	flag.Parse()
	log := runController(*configName)
	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt, os.Kill)
	<-terminate
	log <- "Slack Anything will exit in ~2 sec."
	for _, done := range registry {
		done <- true
	}
	time.Sleep(2 * time.Second)
}

// addToRegistry is helper for workers for add them to global registry
func addToRegistry() chan bool {
	var workerDone = make(chan bool)
	registry = append(registry, workerDone)
	return workerDone
}
