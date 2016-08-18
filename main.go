package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"
)

const (
	reinitTime = 3 * time.Second
)

func main() {
	configName := flag.String("config", ".config/slack-anything", "configuration file")
	flag.Parse()
	go controller(*configName)
	terminate := make(chan os.Signal)
	signal.Notify(terminate, os.Interrupt, os.Kill)
	<-terminate
	fmt.Println("Slack Anything exited.")
}
