package main

import (
	"bufio"
	"flag"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/kyokomi/emoji"
	log "github.com/sirupsen/logrus"
)

func main() {
	var testFileFlag = flag.String("testFile", "", "help message for flag n")
	var logLevelFlag = flag.String("logLevel", "info", "log level for the amount of information (debug, info)")
	var commandPrefixFlag = flag.String("cmdPrefix", ">", "line prefix for commands passed to the application")
	var regexPrefixFlag = flag.String("regexPrefix", "$", "prefix that indicates that the following line should be interpreted as a regex")
	var commentPrefixFlag = flag.String("commentPrefix", "#", "prefix used for comments that are not interpreted")

	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableQuote:     true,
		DisableTimestamp: true,
	})

	switch *logLevelFlag {
	case "debug":
		log.SetLevel(log.DebugLevel)
		break
	case "info":
	default:
		log.SetLevel(log.InfoLevel)
		break
	}

	// open the file specified with the flag
	log.Info("Testing file: ", filepath.Base(*testFileFlag))

	file, err := os.Open(*testFileFlag)
	if err != nil {
		log.Fatal("Error reading file:", err)
	}

	// close the file after finished
	defer func() {
		if err = file.Close(); err != nil {
			log.Fatal("Error closing file:", err)
		}
	}()

	// start the program
	command := flag.Arg(0)
	log.Debug("Executing:", flag.Args())

	cmd := exec.Command(command, flag.Args()[1:]...)

	inPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	outScanner := bufio.NewScanner(outPipe)

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	errScanner := bufio.NewScanner(errPipe)

	go func() {
		for errScanner.Scan() {
			log.Fatal(errScanner.Text())
		}
	}()

	// run the program
	if err := cmd.Start(); err != nil {
		log.Fatal("Execution error:", err)
	}

	lineNumber := 1

	// read lines
	fileScanner := bufio.NewScanner(file)
	for fileScanner.Scan() {

		// line from the file
		var line = fileScanner.Text()

		log.Debug("Reading line:", line)

		// ignore comment files
		if strings.HasPrefix(line, *commentPrefixFlag) {
			continue
		}

		// check if the line is a command
		if strings.HasPrefix(line, *commandPrefixFlag) {
			var lineWithoutPrefix = line[len(*commandPrefixFlag):]
			lineWithoutPrefix = strings.TrimSpace(lineWithoutPrefix)

			log.Debug("Passing command:", lineWithoutPrefix)

			_, err = io.WriteString(inPipe, lineWithoutPrefix+"\n")
			if err != nil {
				log.Fatal("Error passing command:", err)
			}

		} else {

			log.Debug("Trying to read response from std out...")

			if outScanner.Scan() {
				// read the response line from the program
				response := outScanner.Text()

				log.Debug("Reading response:", response)

				if strings.HasPrefix(line, *regexPrefixFlag) {
					regexWithoutPrefix := line[len(*regexPrefixFlag):]

					matched, err := regexp.MatchString(regexWithoutPrefix, response)
					if err != nil {
						log.Fatal("Error matching regex: ", err)
					}

					if !matched {
						logMismatch(regexWithoutPrefix, response, lineNumber)
					}

				} else {
					// check if the line equals the passed line
					if response != line {
						logMismatch(line, response, lineNumber)
					}
				}

			} else {
				log.Fatal("No response from the program, line:", lineNumber)
			}
		}

		lineNumber++
	}

	if err := fileScanner.Err(); err != nil {
		log.Fatal(err)
	}

	log.Info(emoji.Sprint(":check_mark: All tests passed"))
}

func logMismatch(expected, value string, lineNumber int) {
	log.Fatal(emoji.Sprint(":white_exclamation_mark: Error on line ", lineNumber, " expected: ", expected, " got: ", value))
}
