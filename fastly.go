package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/integralist/go-fastly-cli/commands"
	"github.com/integralist/go-fastly-cli/flags"
	"github.com/integralist/go-fastly-cli/standalone"

	fastly "github.com/sethvargo/go-fastly"
	"github.com/sirupsen/logrus"
)

// appVersion is the application version
const appVersion = "0.0.3"

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

func success() {
	os.Exit(0)
}

func failure() {
	os.Exit(1)
}

func settings(version, service string, client *fastly.Client) {
	if version == "latest" {
		standalone.PrintLatestSettings(service, client)
		success()
	}

	if version != "" {
		settingsVersion, err := strconv.Atoi(version)
		if err != nil {
			fmt.Println(err)
			failure()
		}

		standalone.PrintSettingsFor(service, settingsVersion, client)
		success()
	}
}

func main() {
	f := flags.New()

	if *f.Top.Debug == true {
		logrus.SetLevel(logrus.DebugLevel)
	}
	logger.Debug("flags initialised, application starting")

	if showHelp(f) {
		f.Help()
	}

	if *f.Top.Version == true {
		fmt.Println(appVersion)
		os.Exit(1)
	}

	client, err := fastly.NewClient(*f.Top.Token)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *f.Top.Activate != "" {
		standalone.ActivateVersion(*f.Top.Activate, *f.Top.Service, client)
		return
	}

	if *f.Top.Validate != "" {
		standalone.ValidateVersion(*f.Top.Validate, *f.Top.Service, client)
		return
	}

	if *f.Top.Status != "" && *f.Top.Status == "latest" {
		status, err := standalone.GetLatestServiceVersionStatus(*f.Top.Service, client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println(status)
		return
	}

	if *f.Top.Status != "" {
		statusVersion, err := strconv.Atoi(*f.Top.Status)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		status, err := standalone.GetStatusForVersion(*f.Top.Service, statusVersion, client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(status)
		return
	}

	settings(*f.Top.Settings, *f.Top.Service, client)

	args := os.Args[1:] // strip first arg `fastly`
	arg, counter := f.Check(args)

	switch arg {
	case "delete":
		f.Top.Delete.Parse(args[counter:])
		commands.Delete(f, client)
	case "diff":
		f.Top.Diff.Parse(args[counter:])
		commands.Diff(f, client)
	case "list":
		f.Top.List.Parse(args[counter:])
		commands.List(f, client)
	case "upload":
		f.Top.Upload.Parse(args[counter:])
		commands.Upload(f, client)
	default:
		fmt.Printf("%v is not valid command.\n", arg)
		os.Exit(1)
	}
}
