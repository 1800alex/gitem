package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

func (gitm *Gitm) runCommandWithOutputFormatting(command string, args []string) error {
	cmd := exec.Command(command, args...)

	startTime := time.Now()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// join the command and args together as a single string
	// for use in the output formatting
	joined := append([]string{command}, args...)

	gitm.logMutex.Lock()
	fmt.Printf("\033[94m[%s]\033[0m \033[93m[%s]\033[0m %s\n", time.Now().Format("2006-01-02 15:04:05.000"), strings.Join(joined, " "), "\033[97m[running]\033[0m")
	gitm.logMutex.Unlock()

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		gitm.formatAndPrintLines("stdout", strings.Join(joined, " "), stdout)
	}()

	go func() {
		defer wg.Done()
		gitm.formatAndPrintLines("stderr", strings.Join(joined, " "), stderr)
	}()

	err = cmd.Wait()
	wg.Wait()

	duration := time.Since(startTime)
	var durationString string

	if duration < 1*time.Second {
		durationString = fmt.Sprintf("%dms", duration.Milliseconds())
	} else if duration < 1*time.Minute {
		durationString = fmt.Sprintf("%d.%03ds", duration.Seconds(), duration.Milliseconds())
	} else if duration < 1*time.Hour {
		durationString = fmt.Sprintf("%dm%02ds", int(duration.Minutes()), int(duration.Seconds())%60)
	} else {
		durationString = fmt.Sprintf("%dh%02dm", int(duration.Hours()), int(duration.Minutes())%60)
	}

	if err != nil {
		gitm.logMutex.Lock()
		fmt.Printf("\033[94m[%s]\033[0m \033[93m[%s]\033[0m %s\n", time.Now().Format("2006-01-02 15:04:05.000"), strings.Join(joined, " "), fmt.Sprintf("\033[91m[failed after %s]\033[0m", durationString))
		gitm.logMutex.Unlock()
		return err
	}

	gitm.logMutex.Lock()
	fmt.Printf("\033[94m[%s]\033[0m \033[93m[%s]\033[0m %s\n", time.Now().Format("2006-01-02 15:04:05.000"), strings.Join(joined, " "), fmt.Sprintf("\033[97m[completed after %s]\033[0m", durationString))
	gitm.logMutex.Unlock()

	return nil
}

func (gitm *Gitm) formatAndPrintLines(streamName string, command string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		colorReset := "\033[0m"

		var lineColor string
		var streamPrefix string

		if streamName == "stdout" {
			lineColor = "\033[97m" // White for stdout
			streamPrefix = "[stdout] "
		} else if streamName == "stderr" {
			lineColor = "\033[91m" // Red for stderr
			streamPrefix = "[stderr] "
		}

		gitm.logMutex.Lock()
		fmt.Printf("\033[94m[%s]\033[0m \033[93m[%s]\033[0m %s%s%s%s\n", timestamp, command, lineColor, streamPrefix, line, colorReset)
		gitm.logMutex.Unlock()
	}
}
