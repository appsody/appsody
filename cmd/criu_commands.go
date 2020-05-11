package cmd

import (
	"os/exec"
	"time"
	"strings"
)

func StopAndRemoveCriuTempContainer (config *RootCommandConfig, checkpointContainerName string, logger appsodylogger) error {
	stoperr := StopDockerContainer(config, checkpointContainerName, logger)
	if stoperr != nil {
		return stoperr
	}
	removeerr := RemoveDockerContainer(config, checkpointContainerName, logger)
	if removeerr != nil {
		return removeerr
	}

	return nil 
}

func RunToCreateCheckpoint(config *RootCommandConfig, imageName string, logger appsodylogger) error {
	projectName, perr := getProjectName(config)
	checkpointContainerName := projectName + "-criu-checkpoint-runner"
	if perr != nil {
		return perr
	}

	stopAndRemoveContErr := StopAndRemoveCriuTempContainer (config, checkpointContainerName, logger)

	if stopAndRemoveContErr != nil {
		return stopAndRemoveContErr
	}
	
	capabilities := []string {
		"--cap-add AUDIT_CONTROL",
		"--cap-add DAC_READ_SEARCH",
        "--cap-add NET_ADMIN",
        "--cap-add SYS_ADMIN",
        "--cap-add SYS_PTRACE",
        "--cap-add SYS_RESOURCE",
	}
	command_args := []string {"run", "-d", "--name", checkpointContainerName}
	command_args = append(command_args, capabilities...)
	command_args = append(command_args, imageName)

	checkpointRunErr := RunDockerCommandAndWait(config, command_args, logger)

	if checkpointRunErr != nil {
		return checkpointRunErr
	}

	return nil
}

func WaitAndCheckForSuccessfulCheckpoint (config *RootCommandConfig) (error, bool) {
	projectName, perr := getProjectName(config)
	checkpointContainerName := projectName + "-criu-checkpoint-runner"

	args := []string {"exec", "-u", "0", checkpointContainerName, "cat", "criu-log.txt"}
	for i := 0; i < 60; i++ { // Wait for 2 mins and check for checkpoint for every 2 secs 
		out, err := exec.Command("docker", args...).Output()

		if err != nil {
			return err, false
		}

		output := string(out[:])
		if strings.Contains(output, "checkpoint successfull") {
			return nil, true
		}
		time.Sleep(2000 * time.Millisecond)
	}

	if perr != nil {
		return perr, false
	}
	return nil, false
}

func CreateRestorableImage (config *RootCommandConfig, buildImageName string, logger appsodylogger) error {
	projectName, perr := getProjectName(config)
	checkpointContainerName := projectName + "-criu-checkpoint-runner"

	if perr != nil {
		return perr
	}
	commitErr := CommitDockerContainer(config, checkpointContainerName, buildImageName, logger)
	if commitErr != nil {
		return commitErr
	}

	return nil
}
