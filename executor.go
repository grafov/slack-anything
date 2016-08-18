package main

import (
	"github.com/nlopes/slack"
)

// все actions должны это реализовать
type executor interface {
	Do(*slack.Message) error
}
