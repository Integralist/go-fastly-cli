package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/integralist/go-fastly-cli/commands"
	"github.com/integralist/go-fastly-cli/common"
	"github.com/integralist/go-fastly-cli/flags"
	"github.com/integralist/go-fastly-cli/standalone"

	"github.com/sethvargo/go-fastly/fastly"
	"github.com/sirupsen/logrus"
)

// appVersion is the application version
const appVersion = "0.0.5"

var logger *logrus.Entry

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logger = logrus.WithFields(logrus.Fields{
		"package": "main",
	})
}

func showHelp(f flags.Flags) bool {
	if len(os.Args) < 2 || *f.Top.Help == true || *f.Top.HelpShort == true {
		return true
	}
	return false
}

func showSettings(version, service string, client *fastly.Client) {
	if version == "latest" {
		standalone.PrintLatestSettings(service, client)
		common.Success()
	}

	if version != "" {
		settingsVersion, err := strconv.Atoi(version)
		if err != nil {
			fmt.Println(err)
			common.Failure()
		}

		standalone.PrintSettingsFor(service, settingsVersion, client)
		common.Success()
	}
}

func handleStatus(status, service string, client *fastly.Client) {
	if status != "" && status == "latest" {
		status, err := standalone.GetLatestServiceVersionStatus(service, client)
		if err != nil {
			fmt.Println(err)
			common.Failure()
		}

		fmt.Println(status)
		common.Success()
	}

	if status != "" {
		statusVersion, err := strconv.Atoi(status)
		if err != nil {
			fmt.Println(err)
			common.Failure()
		}

		status, err := standalone.GetStatusForVersion(service, statusVersion, client)
		if err != nil {
			fmt.Println(err)
			common.Failure()
		}

		fmt.Println(status)
	}
}

func main() {
	f := flags.New()

	activate := *f.Top.Activate
	debug := *f.Top.Debug
	service := *f.Top.Service
	settings := *f.Top.Settings
	status := *f.Top.Status
	token := *f.Top.Token
	validate := *f.Top.Validate
	version := *f.Top.Version

	if debug == true {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logger.Debug("flags initialised, application starting")

	if version == true {
		fmt.Println(appVersion)
		return
	}

	if showHelp(f) {
		f.Help()
	}

	client, err := fastly.NewClient(token)
	if err != nil {
		fmt.Println(err)
		common.Failure()
	}

	if activate != "" {
		standalone.ActivateVersion(activate, service, client)
		return
	}

	if validate != "" {
		standalone.ValidateVersion(validate, service, client)
		return
	}

	handleStatus(status, service, client)
	showSettings(settings, service, client)

	args := os.Args[1:] // strip first arg `fastly`
	arg, counter := f.Check(args)
	subset := args[counter:]

	switch arg {
	case "delete":
		f.Top.Delete.Parse(subset)
		commands.Delete(f, client)
	case "diff":
		f.Top.Diff.Parse(subset)
		commands.Diff(f, client)
	case "list":
		f.Top.List.Parse(subset)
		commands.List(f, client)
	case "upload":
		f.Top.Upload.Parse(subset)
		commands.Upload(f, client)
	default:
		fmt.Printf("%v is not valid command.\n", arg)
		common.Failure()
	}
}
