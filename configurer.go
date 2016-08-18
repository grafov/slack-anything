package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"
)

const waitForValidConfigFile = 6 * time.Second

const (
	isFilter int8 = iota
	isAction
)

type configActionCode uint8

const (
	setUserSettings configActionCode = iota
	setConfigBlock
	dropConfigBlock
)

type configTask struct {
	ConfigAction configActionCode
}

func initConfigurer(logErr, logInfo chan string, configName string) chan configTask {
	config := make(chan configTask, 1)
	go configurer(logErr, logInfo, config, configName)
	return config
}

func configurer(errLog, infoLog chan string, config chan configTask, configName string) {
	var (
		fd           *os.File
		err          error
		settings     settingsBlock
		configBlocks []*configBlock
	)
	for {
		if fd, err = os.Open(configName); err != nil {
			errLog <- fmt.Sprintf("can't read configuration from %s", configName)
			time.Sleep(waitForValidConfigFile)
		}
		settings, configBlocks, err = parseConfigFile(fd)
		// XXX compare with prev parsed blocks (или делать это в контроллере)
		fmt.Println(settings, configBlocks, err)
	}
}

type (
	settingsBlock struct {
		Token string
		Me    []string
	}
	configBlock struct {
		Checks  []checkWithArgs
		Actions []execWithArgs
	}
	checkWithArgs struct {
		Type checkCode
		Not  bool
		Name string
		Args []string
	}
	execWithArgs struct {
		Type execCode
		Not  bool
		Name string
		Args []string
	}
	checkCode uint8
	execCode  uint8
)

const (
	checkUser checkCode = iota
	checkChannel
	execNothing execCode = iota
	execExtCmd
)

var (
	channelNameMask = regexp.MustCompile("[a-zA-Z0-9_.]+")
	userNameMask    = regexp.MustCompile("[a-zA-Z0-9_.]+")
)

// переводить в промежуточное представление из actionCode + checkCode
// оба в константах из списков выше
// в контроллере обрабатывать это и запускать воркеры соответствующих хендлеров
func parseConfigFile(file *os.File) (settingsBlock, []*configBlock, error) {
	var (
		settings settingsBlock
		blocks   []*configBlock
		errs     parsingErrors
	)
	defer file.Close()
	var (
		line     string
		inBlock  bool
		badBlock bool
		block    *configBlock
	)
	scanner := bufio.NewScanner(file)
	var (
		lineNo  uint
		blockNo uint
	)
	for scanner.Scan() {
		if !inBlock {
			block = &configBlock{}
			blockNo++
		}
		lineNo++
		line = strings.TrimSpace(scanner.Text())
		// skip empty lines and comments
		// empty line means that a new configuration block begins
		switch {
		case line == "":
			switch {
			case inBlock:
				blocks = append(blocks, block)
				inBlock = false
			case badBlock:
				badBlock = false
			}
			continue
		case strings.HasPrefix(line, ";"): // the comments
			continue
		}
		if badBlock {
			inBlock = false
			continue
		}
		var (
			not bool
			cmd string
		)
		switch cmd = line[1:]; line[0] {
		case '<': // the checks
			inBlock = true
			if cmd[0] == '!' {
				not = true
				cmd = cmd[1:]
			}
			switch args := cmd[1:]; cmd[0] {
			case '#': // check for a channel
				name := strings.TrimSpace(args)
				if !channelNameMask.MatchString(name) {
					errs.append(lineNo, blockNo, "invalid channel name", name)
					continue
				}
				block.Checks = append(block.Checks, checkWithArgs{Type: checkChannel, Not: not, Name: name})
				continue
			case '@': // check for a user
				//block.Checks = append(block.Checks, newUserCheck(not, args))
				continue
			}
			switch cmd {
			case "?": // содержание подстроки
			case "~": // регулярка
			case "search": // полнотекстовый поиск
			case "if": // проверка условия
			}
		case '>': // actions
			inBlock = true
		default:
			badBlock = true
		}
	}
	errs.ScanError = scanner.Err()
	return settings, blocks, errs
}

type (
	parsingErrors struct {
		ScanError   error
		BlockErrors []parsingError
	}
	parsingError struct {
		LineNo  uint
		BlockNo uint
		Msg     string
		Args    []interface{}
	}
)

func (e parsingErrors) Error() string {
	var buf bytes.Buffer
	for _, err := range e.BlockErrors {
		buf.WriteString(fmt.Sprintf("line: %d block: %d\t%s: %s\n", err.LineNo, err.BlockNo, err.Msg, err.Args))
	}
	buf.WriteString(e.ScanError.Error())
	return buf.String()
}

func (e *parsingErrors) append(lineNo, blockNo uint, msg string, args ...interface{}) {
	e.BlockErrors = append(e.BlockErrors, parsingError{lineNo, blockNo, msg, args})
}