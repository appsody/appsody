package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
)

//DockerRunAndListen runs a Docker command with arguments in args
//This function does NOT override the image registry (uses args as is)
func DockerRunAndListen(config *RootCommandConfig, args []string, logger appsodylogger, interactive bool) (*exec.Cmd, error) {
	var runArgs = []string{"run"}
	runArgs = append(runArgs, args...)
	return RunDockerCommandAndListen(config, runArgs, logger, interactive)
}

func DockerBuild(config *RootCommandConfig, args []string, logger appsodylogger) error {
	var buildArgs = []string{"build"}
	buildArgs = append(buildArgs, args...)
	return RunDockerCommandAndWait(config, buildArgs, logger)
}
func BuildahBuild(config *RootCommandConfig, args []string, logger appsodylogger) error {
	var buildArgs = []string{"bud"}
	buildArgs = append(buildArgs, args...)
	cmd, err := RunBuildahCommandAndListen(config, buildArgs, logger, false)
	if err != nil {
		return err
	}
	if config.Dryrun {
		config.Info.log("Dry Run - Skipping : cmd.Wait")
		return nil
	}
	return cmd.Wait()
}

func RunDockerCommandAndWait(config *RootCommandConfig, args []string, logger appsodylogger) error {

	cmd, err := RunDockerCommandAndListen(config, args, logger, false)
	if err != nil {
		return err
	}
	if config.Dryrun {
		config.Info.log("Dry Run - Skipping : cmd.Wait")
		return nil
	}
	return cmd.Wait()

}

// RunDockerInspect -TODO - this function should be removed. No one uses it, except the test.
// We are using inspectImage
func RunDockerInspect(log *LoggingConfig, imageName string) (string, error) {
	cmdName := "docker"
	cmdArgs := []string{"image", "inspect", imageName}
	log.Debug.Logf("About to run %s with args %s ", cmdName, cmdArgs)
	inspectCmd := exec.Command(cmdName, cmdArgs...)
	output, err := SeparateOutput(inspectCmd)
	return output, err
}

// RunDockerVolumeList lists all the volumes containing a certain string
func RunDockerVolumeList(log *LoggingConfig, volName string) (string, error) {
	cmdName := "docker"
	cmdArgs := []string{"volume", "ls", "--format", "{{.Name}}"}
	if volName != "" {
		volNameArg := fmt.Sprintf("name=%s", volName)
		cmdArgs = append(cmdArgs, "-f", volNameArg)
	}
	log.Debug.Logf("About to run %s with args %s ", cmdName, cmdArgs)
	inspectCmd := exec.Command(cmdName, cmdArgs...)
	output, err := SeparateOutput(inspectCmd)
	return output, err
}
func RunKubeCommandAndListen(config *RootCommandConfig, args []string, logger appsodylogger, interactive bool) (*exec.Cmd, error) {
	command := "kubectl"
	return RunCommandAndListen(config, command, args, logger, interactive)
}
func RunDockerCommandAndListen(config *RootCommandConfig, args []string, logger appsodylogger, interactive bool) (*exec.Cmd, error) {
	command := "docker"
	return RunCommandAndListen(config, command, args, logger, interactive)
}
func RunBuildahCommandAndListen(config *RootCommandConfig, args []string, logger appsodylogger, interactive bool) (*exec.Cmd, error) {
	command := "buildah"
	return RunCommandAndListen(config, command, args, logger, interactive)
}

func RunCommandAndListen(config *RootCommandConfig, commandValue string, args []string, logger appsodylogger, interactive bool) (*exec.Cmd, error) {
	var execCmd *exec.Cmd
	var command = commandValue
	var err error
	if config.Dryrun {
		config.Info.log("Dry Run - Skipping command: ", command, " ", ArgsToString(args))
	} else {
		config.Info.log("Running command: ", command, " ", ArgsToString(args))
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
				if lastByteNewline && (config.Verbose || logger != config.Info) {
					_, _ = logger.outWriter.Write([]byte("[" + logger.name + "] "))
				}
				_, _ = logger.outWriter.Write([]byte(text))
				lastByteNewline = text == "\n"
			}
		}()

		err = execCmd.Start()
		if err != nil {
			config.Debug.log("Error running ", command, " command: ", logScanner.Text(), err)
			return nil, err
		}

	}
	return execCmd, err
}
