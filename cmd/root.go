// Copyright © 2019 IBM Corporation and others.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	// for logging
	"k8s.io/klog"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var VERSION string

const APIVersionV1 = "v1"

type RootCommandConfig struct {
	CfgFile          string
	Dryrun           bool
	Verbose          bool
	CliConfig        *viper.Viper
	Buildah          bool
	ProjectConfig    *ProjectConfig
	ProjectDir       string
	UnsupportedRepos []string

	// package scoped, these are mostly for caching
	setupConfigRun bool
	imagePulled    map[string]bool
}

// Regular expression to match ANSI terminal commands so that we can remove them from the log
const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[-a-zA-Z\\d\\/#&.:=?%@~_\\s]*)*)?(\u0007|^G))|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PR-TZcf-ntqry=><~]))"

func homeDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", errors.Errorf("%v", err)

	}
	return home, nil
}

const operatorHome = "https://github.com/appsody/appsody-operator/releases/latest/download"

func newRootCmd(projectDir string, args []string) (*cobra.Command, error) {
	rootConfig := &RootCommandConfig{}
	rootConfig.ProjectDir = projectDir
	rootCmd := &cobra.Command{
		Use:           "appsody",
		SilenceErrors: true,
		SilenceUsage:  true,
		Short:         "Appsody CLI",
		Long: `The Appsody command-line tool (CLI) enables the rapid development of cloud native applications.

Complete documentation is available at https://appsody.dev`,
		//Run: no run action for the root command
	}

	rootCmd.PersistentFlags().StringVar(&rootConfig.CfgFile, "config", "", "config file (default is $HOME/.appsody/.appsody.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&rootConfig.Verbose, "verbose", "v", false, "Turns on debug output and logging to a file in $HOME/.appsody/logs")
	rootCmd.PersistentFlags().BoolVar(&rootConfig.Dryrun, "dryrun", false, "Turns on dry run mode")

	// parse the root flags and init logging before adding all the other commands in case those log messages
	rootCmd.SetArgs(args)
	_ = rootCmd.ParseFlags(args) // ignore flag errors here because we haven't added all the commands
	initLogging(rootConfig)

	rootCmd.AddCommand(
		newInitCmd(rootConfig),
		newBuildCmd(rootConfig),
		newExtractCmd(rootConfig),
		newCompletionCmd(rootCmd),
		newDebugCmd(rootConfig),
		newDeployCmd(rootConfig),
		newDocsCmd(rootConfig, rootCmd),
		newListCmd(rootConfig),
		newOperatorCmd(rootConfig),
		newPsCmd(rootConfig),
		newRepoCmd(rootConfig),
		newRunCmd(rootConfig),
		newStackCmd(rootConfig),
		newStopCmd(rootConfig),
		newTestCmd(rootConfig),
		newVersionCmd(rootCmd),
	)

	setupErr := setupConfig(args, rootConfig)
	if setupErr != nil {
		return rootCmd, setupErr
	}
	return rootCmd, nil
}

func setupConfig(args []string, config *RootCommandConfig) error {
	if config.setupConfigRun {
		return nil
	}
	err := InitConfig(config)
	if err != nil {
		return err
	}
	err = ensureConfig(config)
	if err != nil {
		return err
	}

	checkTime(config)
	setNewIndexURL(config)
	config.setupConfigRun = true
	return nil
}

func InitConfig(config *RootCommandConfig) error {

	cliConfig := viper.New()
	homeDirectory, dirErr := homeDir()
	if dirErr != nil {
		return dirErr
	}
	cliConfig.SetDefault("home", filepath.Join(homeDirectory, ".appsody"))
	cliConfig.SetDefault("images", "index.docker.io")
	cliConfig.SetDefault("operator", operatorHome)
	cliConfig.SetDefault("tektonserver", "")
	cliConfig.SetDefault("lastversioncheck", "none")
	if config.CfgFile != "" {
		// Use config file from the flag.
		cliConfig.SetConfigFile(config.CfgFile)
	} else {
		// Search config in home directory with name ".hello-cobra" (without extension).
		cliConfig.AddConfigPath(cliConfig.GetString("home"))
		cliConfig.SetConfigName(".appsody")
	}

	cliConfig.SetEnvPrefix("appsody")
	cliConfig.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	// Ignore errors, if the config isn't found, we will create a default later
	_ = cliConfig.ReadInConfig()
	config.CliConfig = cliConfig
	return nil
}

func getDefaultConfigFile(config *RootCommandConfig) string {
	return filepath.Join(config.CliConfig.GetString("home"), ".appsody.yaml")
}

func Execute(version string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory: ", err)
		os.Exit(1)
	}
	if err := ExecuteE(version, dir, os.Args); err != nil {
		os.Exit(1)
	}
}

func ExecuteE(version string, projectDir string, args []string) error {
	VERSION = version
	rootCmd, err := newRootCmd(projectDir, args[1:])
	if err != nil {
		Error.log(err)
	}
	Debug.log("Running with command line args: appsody ", strings.Join(args[1:], " "))
	err = rootCmd.Execute()
	if err != nil {
		Error.log(err)
	}
	return err
}

type appsodylogger struct {
	name            string
	verbose         bool
	klogInitialized bool
}
type stackTracer interface {
	StackTrace() errors.StackTrace
}

// define the logging levels
var (
	Info       = appsodylogger{name: "Info"}
	Warning    = appsodylogger{name: "Warning"}
	Error      = appsodylogger{name: "Error"}
	Debug      = appsodylogger{name: "Debug"}
	Container  = appsodylogger{name: "Container"}
	InitScript = appsodylogger{name: "InitScript"}
	DockerLog  = appsodylogger{name: "Docker"}
)

var allLoggers = []*appsodylogger{&Info, &Warning, &Error, &Debug, &Container, &InitScript, &DockerLog}

func (l appsodylogger) log(args ...interface{}) {
	msgString := fmt.Sprint(args...)
	l.internalLog(msgString, false, args...)
}

func (l appsodylogger) logf(fmtString string, args ...interface{}) {
	msgString := fmt.Sprintf(fmtString, args...)
	l.internalLog(msgString, false, args...)
}

func (l appsodylogger) Log(args ...interface{}) {
	msgString := fmt.Sprint(args...)
	l.internalLog(msgString, false, args...)
}

func (l appsodylogger) Logf(fmtString string, args ...interface{}) {
	msgString := fmt.Sprintf(fmtString, args...)
	l.internalLog(msgString, false, args...)
}

func (l appsodylogger) LogSkipConsole(args ...interface{}) {
	msgString := fmt.Sprint(args...)
	l.internalLog(msgString, true, args...)
}

func (l appsodylogger) LogfSkipConsole(fmtString string, args ...interface{}) {
	msgString := fmt.Sprintf(fmtString, args...)
	l.internalLog(msgString, true, args...)
}

func (l appsodylogger) internalLog(msgString string, skipConsole bool, args ...interface{}) {
	if l == Debug && !l.verbose {
		return
	}

	if l.verbose || l != Info {
		msgString = "[" + string(l.name) + "] " + msgString
	}

	// if verbose and any of the args are of type error, print the stack traces
	if l.verbose {
		for _, arg := range args {
			st, ok := arg.(stackTracer)
			if ok {
				msgString = fmt.Sprintf("%s\n\n%s%+v", msgString, st, st.StackTrace())
			}
		}
	}

	if !skipConsole {
		// Print to console
		if l == Info {
			fmt.Fprintln(os.Stdout, msgString)
		} else if l == Container {
			fmt.Fprint(os.Stdout, msgString)
		} else {
			fmt.Fprintln(os.Stderr, msgString)
		}
	}

	// Print to log file
	if l.verbose && l.klogInitialized {
		// Remove ansi commands
		ansiRegexp := regexp.MustCompile(ansi)
		msgString = ansiRegexp.ReplaceAllString(msgString, "")
		klog.InfoDepth(2, msgString)
		klog.Flush()
	}
}

func initLogging(config *RootCommandConfig) {
	if config.Verbose {
		for _, l := range allLoggers {
			l.verbose = true
		}

		homeDirectory, dirErr := homeDir()
		if dirErr != nil {
			os.Exit(1)
		}

		logDir := filepath.Join(homeDirectory, ".appsody", "logs")

		_, errPath := os.Stat(logDir)
		if errPath != nil {
			Debug.log("Creating log dir ", logDir)
			if err := os.MkdirAll(logDir, 0755); err != nil {
				Error.logf("Could not create %s: %s", logDir, err)
			}
		}

		currentTimeValues := strings.Split(time.Now().Local().String(), " ")
		fileName := strings.ReplaceAll("appsody"+currentTimeValues[0]+"T"+currentTimeValues[1]+".log", ":", "-")
		pathString := filepath.Join(homeDirectory, ".appsody", "logs", fileName)
		klogFlags := flag.NewFlagSet("klog", flag.ExitOnError)
		klog.InitFlags(klogFlags)
		_ = klogFlags.Set("v", "4")
		_ = klogFlags.Set("skip_headers", "false")
		_ = klogFlags.Set("skip_log_headers", "true")
		_ = klogFlags.Set("log_file", pathString)
		_ = klogFlags.Set("logtostderr", "false")
		_ = klogFlags.Set("alsologtostderr", "false")

		for _, l := range allLoggers {
			l.klogInitialized = true
		}
		Debug.log("Logging to file ", pathString)
	}
}
