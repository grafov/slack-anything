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

func runConfigurer(logErr, logInfo chan string, configName string) chan configTask {
	config := make(chan configTask, 1)
	done := addToRegistry()
	go configurer(done, logErr, logInfo, config, configName)
	return config
}

func configurer(done chan bool, errLog, infoLog chan string, config chan configTask, configName string) {
	var (
		fd           *os.File
		err          error
		settings     settingsBlock
		configBlocks []*configBlock
	)
	for {
		select {
		case <-done:
			return
		default:
			if fd, err = os.Open(configName); err != nil {
				errLog <- fmt.Sprintf("can't read configuration from %s", configName)
				time.Sleep(waitForValidConfigFile)
			}
			settings, configBlocks, err = parseConfigFile(fd)
			// XXX compare with prev parsed blocks (или делать это в контроллере)
			fmt.Println(settings, configBlocks, err)
			for _, b := range configBlocks {
				fmt.Printf("%+v %+v\n", b, b.Checks)
			}
			time.Sleep(1 * time.Second)
		}
	}
}

type (
	settingsBlock struct {
		Token string
		Me    []string
	}
	configBlock struct {
		No      uint16
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
	checkString
	checkRegexp
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
		lineNo  uint16
		blockNo uint16
	)
	for scanner.Scan() {
		if !inBlock {
			block = new(configBlock)
		}
		lineNo++
		line = strings.TrimSpace(scanner.Text())
		// skip empty lines and comments
		// empty line means that a new configuration block begins
		switch {
		case line == "":
			switch {
			case inBlock:
				blockNo++
				block.No = blockNo
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
		switch cmd = strings.TrimSpace(line[1:]); line[0] {
		case '<': // the checks
			inBlock = true
			if cmd[0] == '!' {
				not = true
				cmd = cmd[1:]
			}
			arg := strings.TrimSpace(cmd[1:])
			switch cmd[0] {
			case '#': // check for a channel
				if !channelNameMask.MatchString(arg) {
					errs.append(lineNo, blockNo, "invalid channel name", arg)
					continue
				}
				block.Checks = append(block.Checks, checkWithArgs{Type: checkChannel, Not: not, Name: arg})
			case '@': // check for a user
				if !channelNameMask.MatchString(arg) {
					errs.append(lineNo, blockNo, "invalid username", arg)
					continue
				}
				block.Checks = append(block.Checks, checkWithArgs{Type: checkUser, Not: not, Name: arg})
			case '?': // содержание подстроки
				block.Checks = append(block.Checks, checkWithArgs{Type: checkString, Not: not, Name: arg})
			case '~': // регулярка
				block.Checks = append(block.Checks, checkWithArgs{Type: checkRegexp, Not: not, Name: arg})
			default:
				// проверка многосимвольных команд
				switch arg {
				case "search": // полнотекстовый поиск
				case "if": // проверка условия
				}
			}
		case '>': // actions
			inBlock = true
			switch cmd {

			}
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
		LineNo  uint16
		BlockNo uint16
		Msg     string
		Args    []interface{}
	}
)

func (e parsingErrors) Error() string {
	var buf bytes.Buffer
	for _, err := range e.BlockErrors {
		buf.WriteString(fmt.Sprintf("line: %d block: %d\t%s: %s\n", err.LineNo, err.BlockNo, err.Msg, err.Args))
	}
	if e.ScanError != nil {
		buf.WriteString(e.ScanError.Error())
	}
	return buf.String()
}

func (e *parsingErrors) append(lineNo, blockNo uint16, msg string, args ...interface{}) {
	e.BlockErrors = append(e.BlockErrors, parsingError{lineNo, blockNo, msg, args})
}
