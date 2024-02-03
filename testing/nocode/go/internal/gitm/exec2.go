package gitm

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"
)

type ExecOptions struct {
	// StdoutPrefix is the prefix to use when logging stdout
	StdoutPrefix *string

	// LogStdout is a flag to log stdout
	LogStdout bool

	// Stdout is the buffer to write stdout to
	Stdout *bytes.Buffer

	// LogStderr is a flag to log stderr
	LogStderr bool

	// StderrPrefix is the prefix to use when logging stderr
	StderrPrefix *string

	// Stderr is the buffer to write stderr to
	Stderr *bytes.Buffer

	// Dir is the directory to run the command in
	Dir *string

	// Env is the environment variables to set
	Env []string

	// Command is the command to run
	Command string

	// Args are the arguments to pass to the command
	Args []string
}

// Exec runs a command with the given options
func (gitm *Gitm) Exec(ctx context.Context, opts ExecOptions) error {
	cmd := exec.Command(opts.Command, opts.Args...)

	if opts.Dir != nil {
		cmd.Dir = *opts.Dir
	}

	if len(opts.Env) > 0 {
		cmd.Env = append(os.Environ(), opts.Env...)
	}

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

	logOpts := gitm.LogOptions(opts)

	commandString := gitm.LogCommand(opts.Command, opts.Args)

	if logOpts.Begin != nil {
		beginString := logOpts.Begin.Color + "[running]" + logOpts.Begin.ColorReset

		gitm.logMutex.Lock()
		fmt.Fprintf(logOpts.Begin.FileNumber, "%s%s%s\n", gitm.LogTimestamp(), commandString, beginString)
		gitm.logMutex.Unlock()
	}

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		gitm.HandleLines(logOpts.Stdout, commandString, stdout, opts.Stdout)
	}()

	go func() {
		defer wg.Done()
		gitm.HandleLines(logOpts.Stderr, commandString, stderr, opts.Stderr)
	}()

	err = cmd.Wait()
	wg.Wait()

	if logOpts.Duration != nil {
		duration := time.Since(startTime)
		durationString := logOpts.Duration.Color

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
		durationString += logOpts.Duration.ColorReset

		gitm.logMutex.Lock()
		fmt.Fprintf(logOpts.Duration.FileNumber, "%s%s%s\n", gitm.LogTimestamp(), commandString, durationString)
		gitm.logMutex.Unlock()
	}

	if err != nil {
		return err
	}

	return nil
}

// HandleLines handles the lines from a reader
func (gitm *Gitm) HandleLines(log *LogPipeOptions, command string, r io.Reader, buf *bytes.Buffer) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if buf != nil {
			buf.WriteString(line + "\n")
		}

		if log == nil {
			continue
		}

		gitm.logMutex.Lock()
		fmt.Fprintf(log.FileNumber, "%s%s%s%s%s%s\n", gitm.LogTimestamp(), command, log.Color, log.Prefix, line, log.ColorReset)
		gitm.logMutex.Unlock()
	}
}

type LogPipeOptions struct {
	// Prefix is the prefix to use when logging
	Prefix string

	// FileNumber is the file to write to
	FileNumber io.Writer

	// Color is the color to use when logging
	Color string

	// ColorReset is the color reset to use when logging
	ColorReset string
}

type LogOptions struct {
	Begin    *LogPipeOptions
	Duration *LogPipeOptions
	Stdout   *LogPipeOptions
	Stderr   *LogPipeOptions
}

func (gitm *Gitm) LogOptions(opts ExecOptions) LogOptions {
	logOpts := LogOptions{}

	if gitm.Config.Logging.Begin == nil || true == *gitm.Config.Logging.Begin {
		logPipeOpts := LogPipeOptions{}
		logPipeOpts.FileNumber = os.Stdout

		logPipeOpts.Prefix = "[running]"

		if gitm.Config.Logging.Color == nil || true == *gitm.Config.Logging.Color {
			if gitm.Config.Logging.BeginColor != nil {
				logPipeOpts.Color = ColorStringToShellEscape(*gitm.Config.Logging.BeginColor)
			} else {
				logPipeOpts.Color = shellColorWhite
			}
		}

		if gitm.Config.Logging.Color == nil || true == *gitm.Config.Logging.Color {
			logPipeOpts.ColorReset = shellColorReset
		}

		logOpts.Begin = &logPipeOpts
	}

	if gitm.Config.Logging.Duration == nil || true == *gitm.Config.Logging.Duration {
		logPipeOpts := LogPipeOptions{}
		logPipeOpts.FileNumber = os.Stdout

		if gitm.Config.Logging.Color == nil || true == *gitm.Config.Logging.Color {
			if gitm.Config.Logging.DurationColor != nil {
				logPipeOpts.Color = ColorStringToShellEscape(*gitm.Config.Logging.DurationColor)
			} else {
				logPipeOpts.Color = shellColorMagenta
			}
		}

		if gitm.Config.Logging.Color == nil || true == *gitm.Config.Logging.Color {
			logPipeOpts.ColorReset = shellColorReset
		}

		logOpts.Duration = &logPipeOpts
	}

	if opts.LogStdout {
		logPipeOpts := LogPipeOptions{}
		logPipeOpts.FileNumber = os.Stdout

		if gitm.Config.Logging.StdoutPrefix != nil {
			logPipeOpts.Prefix = *gitm.Config.Logging.StdoutPrefix
		} else {
			logPipeOpts.Prefix = "[stdout] "
		}

		if opts.StdoutPrefix != nil {
			logPipeOpts.Prefix += *opts.StdoutPrefix
		}

		if gitm.Config.Logging.Color == nil || true == *gitm.Config.Logging.Color {
			if gitm.Config.Logging.StdoutColor != nil {
				logPipeOpts.Color = ColorStringToShellEscape(*gitm.Config.Logging.StdoutColor)
			} else {
				logPipeOpts.Color = shellColorWhite
			}

			logPipeOpts.ColorReset = shellColorReset
		}

		logOpts.Stdout = &logPipeOpts
	}

	if opts.LogStderr {
		logPipeOpts := LogPipeOptions{}

		if gitm.Config.Logging.StderrRedirect != nil && true == *gitm.Config.Logging.StderrRedirect {
			logPipeOpts.FileNumber = os.Stdout
		} else {
			logPipeOpts.FileNumber = os.Stderr
		}

		if gitm.Config.Logging.StderrPrefix != nil {
			logPipeOpts.Prefix = *gitm.Config.Logging.StderrPrefix
		} else {
			logPipeOpts.Prefix = "[stderr] "
		}

		if opts.StderrPrefix != nil {
			logPipeOpts.Prefix += *opts.StderrPrefix
		}

		if gitm.Config.Logging.Color == nil || true == *gitm.Config.Logging.Color {
			if gitm.Config.Logging.StderrColor != nil {
				logPipeOpts.Color = ColorStringToShellEscape(*gitm.Config.Logging.StderrColor)
			} else {
				logPipeOpts.Color = shellColorRed
			}

			logPipeOpts.ColorReset = shellColorReset
		}

		logOpts.Stderr = &logPipeOpts
	}

	return logOpts
}
