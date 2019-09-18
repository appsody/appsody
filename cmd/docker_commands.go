package cmd

import (
	"bufio"
	"io"
	"os"
	"os/exec"
	"strings"
)

func DockerRunAndListen(args []string, logger appsodylogger) (*exec.Cmd, error) {
	var runArgs = []string{"run"}
	runArgs = append(runArgs, args...)
	return RunDockerCommandAndListen(runArgs, logger)
}

func DockerBuild(args []string, logger appsodylogger) error {
	var buildArgs = []string{"build"}
	buildArgs = append(buildArgs, args...)
	return RunDockerCommandAndWait(buildArgs, logger)
}

func RunDockerCommandAndWait(args []string, logger appsodylogger) error {

	cmd, err := RunDockerCommandAndListen(args, logger)
	if err != nil {
		return err
	}
	if dryrun {
		Info.log("Dry Run - Skipping : cmd.Wait")

		return nil
	}
	return cmd.Wait()

}

func RunDockerCommandAndListen(args []string, logger appsodylogger) (*exec.Cmd, error) {
	var execCmd *exec.Cmd
	var command = "docker"
	var err error
	if dryrun {
		Info.log("Dry Run - Skipping docker command: ", command, " ", strings.Join(args, " "))
	} else {
		Info.log("Running docker command: ", command, " ", strings.Join(args, " "))
		execCmd = exec.Command(command, args...)

		// Create io pipes for the command
		logReader, logWriter := io.Pipe()
		consoleReader, consoleWriter := io.Pipe()
		execCmd.Stdout = io.MultiWriter(logWriter, consoleWriter)
		execCmd.Stderr = io.MultiWriter(logWriter, consoleWriter)
		if interactive {
			execCmd.Stdin = os.Stdin
		}

		// Create a scanner for both the log and the console
		// The log will be written when a newline is encountered
		logScanner := bufio.NewScanner(logReader)
		logScanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		go func() {
			for logScanner.Scan() {
				logger.LogSkipConsole(logScanner.Text())
			}
		}()

		// The console will be written on every byte
		consoleScanner := bufio.NewScanner(consoleReader)
		consoleScanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		consoleScanner.Split(bufio.ScanBytes)
		go func() {
			lastByteNewline := true
			for consoleScanner.Scan() {
				text := consoleScanner.Text()
				if lastByteNewline && (verbose || logger != Info) {
					os.Stdout.WriteString("[" + string(logger) + "] ")
				}
				os.Stdout.WriteString(text)
				lastByteNewline = text == "\n"
			}
		}()

		err = execCmd.Start()
		if err != nil {
			Debug.log("Error running ", command, " command: ", logScanner.Text(), err)
			return nil, err
		}

	}
	return execCmd, err
}
