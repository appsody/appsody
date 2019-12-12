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
	"io"
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
var CONTROLLERVERSION string

const APIVersionV1 = "v1"

type LoggingConfig struct {
	// define the logging levels
	Info       appsodylogger
	Warning    appsodylogger
	Error      appsodylogger
	Debug      appsodylogger
	Container  appsodylogger
	InitScript appsodylogger
	DockerLog  appsodylogger
	BuildahLog appsodylogger
}

type RootCommandConfig struct {
	*LoggingConfig

	CfgFile          string
	Dryrun           bool
	Verbose          bool
	CliConfig        *viper.Viper
	Buildah          bool
	ProjectConfig    *ProjectConfig
	ProjectDir       string
	UnsupportedRepos []string
	StackRegistry    string

	// package scoped, these are mostly for caching
	setupConfigRun bool
	imagePulled    map[string]bool
	cachedEnvVars  map[string]string
}

// Regular expression to match ANSI terminal commands so that we can remove them from the log
const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[-a-zA-Z\\d\\/#&.:=?%@~_\\s]*)*)?(\u0007|^G))|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PR-TZcf-ntqry=><~]))"

const operatorHome = "https://github.com/appsody/appsody-operator/releases/latest/download"

func newRootCmd(projectDir string, outWriter, errWriter io.Writer, args []string) (*cobra.Command, *RootCommandConfig, error) {
	loggingConfig := &LoggingConfig{}
	loggingConfig.InitLogging(outWriter, errWriter)
	rootConfig := &RootCommandConfig{LoggingConfig: loggingConfig}

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
	// ignore errors on unknown flags at this point and continue parsing the root flags
	// later the Execute func will parse the flags again
	rootCmd.FParseErrWhitelist = cobra.FParseErrWhitelist{UnknownFlags: true}
	_ = rootCmd.ParseFlags(args) // ignore flag errors here because we haven't added all the commands
	err := rootConfig.initLogging()
	if err != nil {
		return rootCmd, rootConfig, err
	}
	setupErr := setupConfig(args, rootConfig)
	if setupErr != nil {
		return rootCmd, rootConfig, setupErr
	}

	rootCmd.AddCommand(
		newInitCmd(rootConfig),
		newBuildCmd(rootConfig),
		newExtractCmd(rootConfig),
		newCompletionCmd(rootConfig.LoggingConfig, rootCmd),
		newDebugCmd(rootConfig),
		newDeployCmd(rootConfig),
		newDocsCmd(rootConfig.LoggingConfig, rootCmd),
		newListCmd(rootConfig),
		newOperatorCmd(rootConfig),
		newPsCmd(rootConfig.LoggingConfig),
		newRepoCmd(rootConfig),
		newRunCmd(rootConfig),
		newStackCmd(rootConfig),
		newStopCmd(rootConfig),
		newTestCmd(rootConfig),
		newVersionCmd(rootConfig.LoggingConfig, rootCmd),
	)

	appsodyOnK8S := os.Getenv("APPSODY_K8S_EXPERIMENTAL")
	if appsodyOnK8S == "TRUE" {
		rootConfig.Buildah = true
	}
	//Invalidate the cache
	rootConfig.ProjectConfig = nil
	return rootCmd, rootConfig, nil
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
	setNewRepoName(config)
	config.setupConfigRun = true
	return nil
}

func InitConfig(config *RootCommandConfig) error {

	cliConfig := viper.New()
	homeDirectory, dirErr := homedir.Dir()
	if dirErr != nil {
		return dirErr
	}
	cliConfig.SetDefault("home", filepath.Join(homeDirectory, ".appsody"))
	cliConfig.SetDefault("images", "docker.io")
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

func Execute(version string, controllerVersion string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory: ", err)
		os.Exit(1)
	}
	if err := ExecuteE(version, controllerVersion, dir, os.Stdout, os.Stderr, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}

func ExecuteE(version string, controllerVersion string, projectDir string, outWriter, errWriter io.Writer, args []string) error {
	VERSION = version
	CONTROLLERVERSION = controllerVersion

	rootCmd, rootConfig, err := newRootCmd(projectDir, outWriter, errWriter, args)
	if err != nil {
		rootConfig.Error.log(err)
	}
	rootConfig.Debug.log("Running with command line args: appsody ", strings.Join(args, " "))
	err = rootCmd.Execute()
	if err != nil {
		rootConfig.Error.log(err)
	}
	return err
}

type appsodylogger struct {
	name      string
	verbose   bool
	outWriter io.Writer
	errWriter io.Writer
}
type stackTracer interface {
	StackTrace() errors.StackTrace
}

var klogInitialized = false

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
	if l.name == "Debug" && !l.verbose {
		return
	}

	if l.verbose || l.name != "Info" {
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
		if l.name == "Info" {
			fmt.Fprintln(l.outWriter, msgString)
		} else if l.name == "Container" {
			fmt.Fprint(l.outWriter, msgString)
		} else {
			fmt.Fprintln(l.errWriter, msgString)
		}
	}

	// Print to log file
	if l.verbose && klogInitialized {
		// Remove ansi commands
		ansiRegexp := regexp.MustCompile(ansi)
		msgString = ansiRegexp.ReplaceAllString(msgString, "")
		klog.InfoDepth(2, msgString)
		klog.Flush()
	}
}

// InitLogging initializes the logging configuration for a given RootCommandConfig.
// The initialization of klog is global and will only be performed once.
func (config *LoggingConfig) InitLogging(outWriter, errWriter io.Writer) {
	config.Info = appsodylogger{name: "Info"}
	config.Warning = appsodylogger{name: "Warning"}
	config.Error = appsodylogger{name: "Error"}
	config.Debug = appsodylogger{name: "Debug"}
	config.Container = appsodylogger{name: "Container"}
	config.InitScript = appsodylogger{name: "InitScript"}
	config.DockerLog = appsodylogger{name: "Docker"}
	config.BuildahLog = appsodylogger{name: "Buildah"}

	var allLoggers = []*appsodylogger{&config.Info, &config.Warning, &config.Error, &config.Debug, &config.Container, &config.InitScript, &config.DockerLog, &config.BuildahLog}

	for _, l := range allLoggers {
		l.outWriter = outWriter
		l.errWriter = errWriter
	}
}

func (config *RootCommandConfig) initLogging() error {
	var allLoggers = []*appsodylogger{&config.Info, &config.Warning, &config.Error, &config.Debug, &config.Container, &config.InitScript, &config.DockerLog}
	if config.Verbose {
		for _, l := range allLoggers {
			l.verbose = true
		}

		homeDirectory, dirErr := homedir.Dir()
		if dirErr != nil {
			return errors.Errorf("Error getting home directory: %v", dirErr)
		}

		logDir := filepath.Join(homeDirectory, ".appsody", "logs")

		_, errPath := os.Stat(logDir)
		if errPath != nil {
			config.Debug.log("Creating log dir ", logDir)
			if err := os.MkdirAll(logDir, 0755); err != nil {
				config.Error.logf("Could not create %s: %s", logDir, err)
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

		klogInitialized = true
		config.Debug.log("Logging to file ", pathString)
	}
	return nil
}
