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

var (
	// VERSION is set during build
	VERSION         string
	cfgFile         string
	cliConfig       *viper.Viper
	APIVersionV1    = "v1"
	dryrun          bool
	verbose         bool
	klogInitialized = false
)

// Regular expression to match ANSI terminal commands so that we can remove them from the log
const ansi = "[\u001B\u009B][[\\]()#;?]*(?:(?:(?:[a-zA-Z\\d]*(?:;[-a-zA-Z\\d\\/#&.:=?%@~_\\s]*)*)?(\u0007|^G))|(?:(?:\\d{1,4}(?:;\\d{0,4})*)?[\\dA-PR-TZcf-ntqry=><~]))"

var ansiRegexp = regexp.MustCompile(ansi)

func homeDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", errors.Errorf("%v", err)

	}
	return home, nil
}

var operatorHome = "https://github.com/appsody/appsody-operator/releases/latest/download"

var rootCmd = &cobra.Command{
	Use:           "appsody",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "Appsody CLI",
	Long: `The Appsody command-line tool (CLI) enables the rapid development of cloud native applications.

Complete documentation is available at https://appsody.dev`,
	//Run: no run action for the root command
}

func setupConfig() error {
	err := initConfig()
	if err != nil {
		return err
	}
	err = ensureConfig()
	if err != nil {
		return err
	}

	checkTime()
	setNewIndexURL()
	return nil
}

func init() {
	// Don't run this on help commands
	// TODO - instead of the isHelpCommand() check, we should delay the config init/ensure until we really need the config
	if !isHelpCommand() {
		cobra.OnInitialize(initLogging)
		//cobra.OnInitialize(initConfig)
		//cobra.OnInitialize(ensureConfig)
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.appsody/.appsody.yaml)")
	// Added for logging
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Turns on debug output and logging to a file in $HOME/.appsody/logs")

	rootCmd.PersistentFlags().BoolVar(&dryrun, "dryrun", false, "Turns on dry run mode")

}

func isHelpCommand() bool {
	if len(os.Args) <= 1 {
		return true
	}
	if os.Args[1] == "help" {
		return true
	}
	for _, arg := range os.Args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

var initConfigRun = false

func initConfig() error {
	Debug.log("Running with command line args: appsody ", strings.Join(os.Args[1:], " "))
	if initConfigRun {
		return nil
	}
	cliConfig = viper.New()
	homeDirectory, dirErr := homeDir()
	if dirErr != nil {
		return dirErr
	}
	cliConfig.SetDefault("home", filepath.Join(homeDirectory, ".appsody"))
	cliConfig.SetDefault("images", "index.docker.io")
	cliConfig.SetDefault("operator", operatorHome)
	cliConfig.SetDefault("tektonserver", "")
	cliConfig.SetDefault("lastversioncheck", "none")
	if cfgFile != "" {
		// Use config file from the flag.
		cliConfig.SetConfigFile(cfgFile)
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
	initConfigRun = true
	return nil
}

func getDefaultConfigFile() string {
	return filepath.Join(cliConfig.GetString("home"), ".appsody.yaml")
}

func Execute(version string) {
	VERSION = version

	if err := rootCmd.Execute(); err != nil {
		Error.log(err)
		os.Exit(1)
	}
}

type appsodylogger string
type stackTracer interface {
	StackTrace() errors.StackTrace
}

// define the logging levels
var (
	Info       appsodylogger = "Info"
	Warning    appsodylogger = "Warning"
	Error      appsodylogger = "Error"
	Debug      appsodylogger = "Debug"
	Container  appsodylogger = "Container"
	InitScript appsodylogger = "InitScript"
	DockerLog  appsodylogger = "Docker"
)

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
	if l == Debug && !verbose {
		return
	}

	if verbose || l != Info {
		msgString = "[" + string(l) + "] " + msgString
	}

	// if verbose and any of the args are of type error, print the stack traces
	if verbose {
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
	if verbose && klogInitialized {
		// Remove ansi commands
		msgString = ansiRegexp.ReplaceAllString(msgString, "")
		klog.InfoDepth(2, msgString)
		klog.Flush()
	}
}

func initLogging() {

	if verbose {
		// this is an initizer method and currently you can not return an error from them

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
		klogInitialized = true
		Debug.log("Logging to file ", pathString)
	}
}
