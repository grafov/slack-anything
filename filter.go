package main

import (
	"github.com/nlopes/slack"
)

type checker interface {
	Check(*slack.Message) bool
}

func filter(block *configBlock, in chan interface{}) {

}
