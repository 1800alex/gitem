package gitm

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

var (
	shellColorReset   = "\033[0m"
	shellColorRed     = "\033[91m"
	shellColorGreen   = "\033[92m"
	shellColorYellow  = "\033[93m"
	shellColorBlue    = "\033[94m"
	shellColorMagenta = "\033[95m"
	shellColorCyan    = "\033[96m"
	shellColorWhite   = "\033[97m"

	shellColorMap = map[string]string{
		"red":     shellColorRed,
		"green":   shellColorGreen,
		"yellow":  shellColorYellow,
		"blue":    shellColorBlue,
		"magenta": shellColorMagenta,
		"cyan":    shellColorCyan,
		"white":   shellColorWhite,
		"reset":   shellColorReset,
	}
)

func ColorStringToShellEscape(colorString string) string {
	val, ok := shellColorMap[colorString]
	if ok {
		return val
	}

	return ""
}

func (gitm *Gitm) LogTimestamp() string {
	msg := ""

	if gitm.config.Logging.Timestamps != nil && false == *gitm.config.Logging.Timestamps {
		return msg
	}

	if gitm.config.Logging.Color == nil || true == *gitm.config.Logging.Color {
		if gitm.config.Logging.TimestampColor != nil {
			msg += ColorStringToShellEscape(*gitm.config.Logging.TimestampColor)
		} else {
			msg += shellColorBlue
		}
	}

	msg += fmt.Sprintf("[%s]", time.Now().Format("2006-01-02 15:04:05.000"))

	if gitm.config.Logging.Color == nil || true == *gitm.config.Logging.Color {
		msg += shellColorReset
	}

	return msg + " "
}

func (gitm *Gitm) LogCommand(command string, args []string) string {
	msg := ""

	if gitm.config.Logging.Commands != nil && false == *gitm.config.Logging.Commands {
		return msg
	}

	if gitm.config.Logging.Color == nil || true == *gitm.config.Logging.Color {
		if gitm.config.Logging.CommandColor != nil {
			msg += ColorStringToShellEscape(*gitm.config.Logging.CommandColor)
		} else {
			msg += shellColorCyan
		}
	}

	// join the command and args together as a single string
	// for use in the output formatting
	joined := append([]string{command}, args...)
	msg += "[" + strings.Join(joined, " ") + "]"

	if gitm.config.Logging.Color == nil || true == *gitm.config.Logging.Color {
		msg += shellColorReset
	}

	return msg + " "
}

func (gitm *Gitm) RunCommandWithOutputFormatting(ctx context.Context, command string, args []string) error {
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

	// kill the process if the context is cancelled
	done := make(chan error)
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
		case <-done:
		}
	}()

	commandString := gitm.LogCommand(command, args)

	if gitm.config.Logging.Begin == nil || true == *gitm.config.Logging.Begin {
		beginString := ""
		if gitm.config.Logging.Color == nil || true == *gitm.config.Logging.Color {
			if gitm.config.Logging.BeginColor != nil {
				beginString += ColorStringToShellEscape(*gitm.config.Logging.BeginColor)
			} else {
				beginString += shellColorWhite
			}
		}

		beginString += "[running]"

		if gitm.config.Logging.Color == nil || true == *gitm.config.Logging.Color {
			beginString += shellColorReset
		}

		gitm.logMutex.Lock()
		fmt.Printf("%s%s%s\n", gitm.LogTimestamp(), commandString, beginString)
		gitm.logMutex.Unlock()
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		gitm.FormatAndPrintLines(1, commandString, stdout)
	}()

	go func() {
		defer wg.Done()
		gitm.FormatAndPrintLines(2, commandString, stderr)
	}()

	err = cmd.Wait()
	wg.Wait()

	duration := time.Since(startTime)
	var durationString string

	if gitm.config.Logging.Duration == nil || true == *gitm.config.Logging.Duration {
		if gitm.config.Logging.Color == nil || true == *gitm.config.Logging.Color {
			if gitm.config.Logging.DurationColor != nil {
				durationString += ColorStringToShellEscape(*gitm.config.Logging.DurationColor)
			} else {
				durationString += shellColorMagenta
			}
		}

		if err != nil {
			durationString += "[failed after "
		} else {
			durationString += "[completed after "
		}

		if duration < 1*time.Second {
			durationString += fmt.Sprintf("%dms", duration.Milliseconds())
		} else if duration < 1*time.Minute {
			durationString += fmt.Sprintf("%d.%03ds", duration.Seconds(), duration.Milliseconds())
		} else if duration < 1*time.Hour {
			durationString += fmt.Sprintf("%dm%02ds", int(duration.Minutes()), int(duration.Seconds())%60)
		} else {
			durationString += fmt.Sprintf("%dh%02dm", int(duration.Hours()), int(duration.Minutes())%60)
		}

		durationString += "]"

		if gitm.config.Logging.Color == nil || true == *gitm.config.Logging.Color {
			durationString += shellColorReset
		}
	}

	gitm.logMutex.Lock()
	fmt.Printf("%s%s%s\n", gitm.LogTimestamp(), commandString, durationString)
	gitm.logMutex.Unlock()

	if err != nil {
		return err
	}

	return nil
}

func (gitm *Gitm) FormatAndPrintLines(streamType int, command string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		var lineColor string
		var colorReset string
		var streamPrefix string

		if streamType == 1 {
			if gitm.config.Logging.StdoutPrefix != nil {
				streamPrefix = *gitm.config.Logging.StdoutPrefix
			} else {
				streamPrefix = "[stdout] "
			}

			if gitm.config.Logging.Color == nil || true == *gitm.config.Logging.Color {
				if gitm.config.Logging.StdoutColor != nil {
					lineColor = ColorStringToShellEscape(*gitm.config.Logging.StdoutColor)
				} else {
					lineColor = shellColorWhite
				}

				colorReset = shellColorReset
			}
		} else if streamType == 2 {
			if gitm.config.Logging.StdoutPrefix != nil {
				streamPrefix = *gitm.config.Logging.StdoutPrefix
			} else {
				streamPrefix = "[stderr] "
			}

			if gitm.config.Logging.Color == nil || true == *gitm.config.Logging.Color {
				if gitm.config.Logging.StderrColor != nil {
					lineColor = ColorStringToShellEscape(*gitm.config.Logging.StderrColor)
				} else {
					lineColor = shellColorRed
				}

				colorReset = shellColorReset
			}

		}

		gitm.logMutex.Lock()
		if gitm.config.Logging.StderrRedirect != nil && true == *gitm.config.Logging.StderrRedirect {
			fmt.Printf("%s%s%s%s%s%s\n", gitm.LogTimestamp(), command, lineColor, streamPrefix, line, colorReset)
		} else {
			fmt.Fprintf(os.Stderr, "%s%s%s%s%s%s\n", gitm.LogTimestamp(), command, lineColor, streamPrefix, line, colorReset)
		}
		gitm.logMutex.Unlock()
	}
}

type CommandOutput struct {
	Stdout   []string
	Duration time.Duration
}

func (gitm *Gitm) RunCommandAndCaptureOutput(ctx context.Context, command string, args []string) (CommandOutput, error) {
	result := CommandOutput{}
	cmd := exec.Command(command, args...)

	startTime := time.Now()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return result, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return result, err
	}

	if err := cmd.Start(); err != nil {
		return result, err
	}

	// kill the process if the context is cancelled
	done := make(chan error)
	defer close(done)
	go func() {
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
		case <-done:
		}
	}()

	commandString := gitm.LogCommand(command, args)

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		result.Stdout = gitm.CaptureLines(stdout)
	}()

	go func() {
		defer wg.Done()
		gitm.FormatAndPrintLines(2, commandString, stderr)
	}()

	err = cmd.Wait()
	wg.Wait()

	result.Duration = time.Since(startTime)

	if err != nil {
		return result, err
	}

	return result, nil
}

func (gitm *Gitm) CaptureLines(r io.Reader) []string {
	var output []string
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		output = append(output, scanner.Text())
	}

	return output
}
