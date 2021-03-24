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

var testFileFlag = flag.String("testFile", "", "help message for flag n")
var commandPrefixFlag = flag.String("cmdPrefix", ">", "line prefix for commands passed to the application")
var regexPrefixFlag = flag.String("regexPrefix", "$", "prefix that indicates that the following line should be interpreted as a regex")
var commentPrefixFlag = flag.String("commentPrefix", "#", "prefix used for comments that are not interpreted")
var argsPrefixFlag = flag.String("argsPrefix", "$$", "define how the first line starts when passing args")
var logLevelFlag = flag.String("logLevel", "info", "log level for the amount of information (debug, info)")

func main() {

	flag.Parse()

	log.SetFormatter(&log.TextFormatter{
		ForceColors:      true,
		DisableQuote:     true,
		DisableTimestamp: true,
	})

	if *logLevelFlag == "debug" {
		log.SetLevel(log.DebugLevel)
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

	firstLine := ""

	// read lines
	fileScanner := bufio.NewScanner(file)
	if fileScanner.Scan() {
		firstLine = fileScanner.Text()
	} else {
		log.Fatal("File is empty.")
	}

	args := make([]string, 0)

	// if the file starts with $$ in the first line
	// read it as additional arguments
	if strings.HasPrefix(firstLine, *argsPrefixFlag) {
		argString := firstLine[len(*argsPrefixFlag):]
		additionalFileArgs := strings.Split(argString, " ")

		log.Debug("Args read from file: ", additionalFileArgs)

		args = append(args, additionalFileArgs...)
		firstLine = ""
	}

	// command to execute on the shell
	command := flag.Arg(0)
	args = append(flag.Args()[1:], args...)

	log.Debug("Executing:", args)
	// start the program
	cmd := exec.Command(command, args...)

	outScanner, errScanner, inPipe := applicationStdInterface(cmd)

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

	if len(firstLine) > 0 {
		checkLine(firstLine, lineNumber, outScanner, &inPipe)
		lineNumber++
	}

	for fileScanner.Scan() {

		// line from the file
		var line = fileScanner.Text()
		log.Debug("Reading line:", line)

		checkLine(line, lineNumber, outScanner, &inPipe)

		lineNumber++
	}

	if err := fileScanner.Err(); err != nil {
		log.Fatal(err)
	}

	log.Info(emoji.Sprint(":check_mark: All tests passed"))
}

func applicationStdInterface(cmd *exec.Cmd) (outScanner, errScanner *bufio.Scanner, inPipe io.WriteCloser) {
	inPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Fatal(err)
	}

	outPipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	outScanner = bufio.NewScanner(outPipe)

	errPipe, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	errScanner = bufio.NewScanner(errPipe)
	return
}

func logMismatch(expected, value string, lineNumber int) {
	log.Fatal(emoji.Sprint(":white_exclamation_mark: Error on line ", lineNumber, " expected: ", expected, " got: ", value))
}

func checkLine(line string, lineNumber int, outScanner *bufio.Scanner, inPipe *io.WriteCloser) {
	// ignore comment files
	if strings.HasPrefix(line, *commentPrefixFlag) {
		return
	}

	// check if the line is a command
	// then passing it to the running application
	if strings.HasPrefix(line, *commandPrefixFlag) {
		var lineWithoutPrefix = line[len(*commandPrefixFlag):]

		log.Debug("Passing command:", lineWithoutPrefix)

		_, err := io.WriteString(*inPipe, lineWithoutPrefix+"\n")
		if err != nil {
			log.Fatal("Error passing command:", err)
		}

		return
	}

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
