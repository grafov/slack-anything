package main

import (
	"bufio"
	"os"
	"strings"

	"github.com/grafov/slackwatcher/check"
	"github.com/grafov/slackwatcher/slackapi"
)

var (
	settings struct {
		Token string
		Me    []string
	}
	configBlocks []*configBlock
)

type configBlock struct {
	Type    sourceType
	Sources []string
	Checks  []checker
	Actions []executor
}

type sourceType uint8

const (
	channel sourceType = iota
	user
	any
)

type checker interface {
	Do(*slackapi.Message) bool
}

type executor interface {
	Do(*slackapi.Message) error
}

func parseConfigFile(filename string) error {
	var (
		file *os.File
		err  error
	)

	file, err = os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var (
		line                         string
		inChanBlock, inSettingsBlock bool
		skipBlock                    bool
		block                        *configBlock
	)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line = strings.TrimSpace(scanner.Text())

		// skip empty lines and comments
		// empty line means that a new configuration block begins
		switch {
		case line == "":
			switch {
			case inChanBlock:
				blocks = append(blocks, block)
				inChanBlock = false
			case inSettingsBlock:
				inSettingsBlock = false
			case skipBlock:
				blocks = append(blocks, block)
				skipBlock = false
			}
			continue
		case strings.HasPrefix(line, "#"):
			continue
		}

		// check for head line of a config block for determine its type and sources
		// creates appropriate config block
		if !inChanBlock && !inSettingsBlock {
			splittedLine := strings.Split(line, " ")
			switch splittedLine[0] {
			// these blocks consits of settings for application
			case "settings":
				// TODO
				inSettingsBlock = true
			case "channel":
				// TODO set up block with all channels list
				if len(splittedLine) == 1 {
					// log error
					skipBlock = true
					continue
				}
				block = &configBlock{Type: channel, Sources: splittedLine[1:]}
				inChanBlock = true
				continue
			case "user":
				if len(splittedLine) == 1 {
					// log error
					skipBlock = true
					continue
				}
				block = &configBlock{Type: user, Sources: splittedLine[1:]}
				inChanBlock = true
			}
		}

		if inSettingsBlock {
			// TODO
			continue
		}

		// parse a check or an action for a channel
		if inChanBlock {
			cmdWithArgs := strings.SplitN(line, " ", 2)
			if len(cmdWithArgs) <= 1 {
				// TODO log error
				skipBlock = true
				continue
			}
			cmd := strings.ToLower(cmdWithArgs[0])
			switch cmd {
			case "contains":
				block.Checks = append(block.Checks, check.NewContains(cmdWithArgs[1]))
			case "regexp":
				if newCheck, err := check.NewRegexp(cmdWithArgs[1]); err != nil {
					// TODO log error
					continue
				}
				block.Checks = append(block.Checks, newCheck)
			case "log":
				if newCheck, err := check.NewLog(cmdWithArgs[1]); err != nil {
					continue
				}
				block.Actions = append(block.Actions, newAction)
			case "run":
			case "save":
			}
		}

	}
	return scanner.Err()
}
