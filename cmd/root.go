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
	"strings"
	"time"

	// for logging
	"k8s.io/klog"

	//  homedir "github.com/mitchellh/go-homedir"

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

func homeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		Error.log(err)
		os.Exit(1)
	}
	return home
}

var rootCmd = &cobra.Command{
	Use:           "appsody",
	SilenceErrors: true,
	SilenceUsage:  true,
	Short:         "Appsody CLI",
	Long: `The Appsody command-line tool (CLI) enables the rapid development of cloud native applications.

Complete documentation is available at https://appsody.dev`,
	//Run: no run action for the root command
}

func init() {
	// Don't run this on help commands
	// TODO - instead of the isHelpCommand() check, we should delay the config init/ensure until we really need the config
	if !isHelpCommand() {
		cobra.OnInitialize(initLogging)
		cobra.OnInitialize(initConfig)
		cobra.OnInitialize(ensureConfig)
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

func initConfig() {
	Debug.log("Running with command line args: appsody ", strings.Join(os.Args[1:], " "))
	cliConfig = viper.New()

	cliConfig.SetDefault("home", filepath.Join(homeDir(), ".appsody"))
	cliConfig.SetDefault("images", "index.docker.io")
	cliConfig.SetDefault("tektonserver", "")
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
}

func getDefaultConfigFile() string {
	return filepath.Join(cliConfig.GetString("home"), ".appsody.yaml")
}

func Execute(version string) {
	VERSION = version

	if err := rootCmd.Execute(); err != nil {
		Error.logE(err)
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

func (l appsodylogger) logE(err error) {
	//msgString := fmt.Sprint(args...)
	if verbose && klogInitialized {

		st := err.(stackTracer).StackTrace()

		klog.InfoDepth(2, st[0:2])
		klog.Flush()
	}
	l.log(err)
}
func (l appsodylogger) log(args ...interface{}) {
	msgString := fmt.Sprint(args...)
	l.internalLog(msgString)
}

func (l appsodylogger) logf(fmtString string, args ...interface{}) {
	msgString := fmt.Sprintf(fmtString, args...)
	l.internalLog(msgString)
}

func (l appsodylogger) internalLog(msgString string) {
	if l == Debug && !verbose {
		return
	}

	if verbose || l != Info {
		msgString = "[" + string(l) + "] " + msgString
	}

	// Print to console
	if l == Info {
		fmt.Fprintln(os.Stdout, msgString)
	} else {
		fmt.Fprintln(os.Stderr, msgString)
	}

	// Print to log file
	if verbose && klogInitialized {
		klog.InfoDepth(2, msgString)
		klog.Flush()
	}
}

func initLogging() {

	if verbose {

		logDir := filepath.Join(homeDir(), ".appsody", "logs")

		_, errPath := os.Stat(logDir)
		if errPath != nil {
			Debug.log("Creating log dir ", logDir)
			if err := os.MkdirAll(logDir, 0755); err != nil {
				Error.logf("Could not create %s: %s", logDir, err)
			}
		}

		currentTimeValues := strings.Split(time.Now().Local().String(), " ")
		fileName := strings.ReplaceAll("appsody"+currentTimeValues[0]+"T"+currentTimeValues[1]+".log", ":", "-")
		pathString := filepath.Join(homeDir(), ".appsody", "logs", fileName)
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
