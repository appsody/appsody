package cmd

import (
	"bufio"
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

		cmdReader, err := execCmd.StdoutPipe()
		if err != nil {
			Error.log("Error creating StdoutPipe for docker Cmd ", err)
			return nil, err
		}

		errReader, err := execCmd.StderrPipe()
		if err != nil {
			Error.log("Error creating StderrPipe for docker Cmd ", err)
			return nil, err
		}

		outScanner := bufio.NewScanner(cmdReader)
		outScanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		go func() {
			for outScanner.Scan() {
				logger.log(outScanner.Text())
			}
		}()

		errScanner := bufio.NewScanner(errReader)
		errScanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		go func() {
			for errScanner.Scan() {
				logger.log(errScanner.Text())
			}
		}()

		err = execCmd.Start()
		if err != nil {
			Debug.log("Error running ", command, " command: ", errScanner.Text(), err)
			return nil, err
		}

	}
	return execCmd, err
}
