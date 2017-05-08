package main

import (
	"commands"
	"flag"
	"flags"
	"fmt"
	"os"
	"standalone"

	"github.com/fatih/color"
	fastly "github.com/sethvargo/go-fastly"
	"github.com/sirupsen/logrus"
)

// appVersion is the application version
const appVersion = "0.0.1"

// useful colour settings for printing messages
var yellow = color.New(color.FgYellow).SprintFunc()

var logger *logrus.Entry

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logger = logrus.WithFields(logrus.Fields{
		"package": "main",
	})
}

func main() {
	f := flags.New()

	if *f.Top.Help == true {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *f.Top.Version == true {
		fmt.Println(appVersion)
		os.Exit(1)
	}

	if *f.Top.Debug == true {
		logrus.SetLevel(logrus.DebugLevel)
	}

	client, err := fastly.NewClient(*f.Top.Token)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if *f.Top.Activate != "" {
		standalone.ActivateVersion(f, client)
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
		status, err := standalone.GetStatusForVersion(*f.Top.Service, *f.Top.Status, client)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Println(status)
		return
	}

	if *f.Top.Settings == "latest" {
		standalone.PrintLatestSettings(*f.Top.Service, client)
		return
	}

	if *f.Top.Settings != "" {
		standalone.PrintSettingsFor(*f.Top.Service, *f.Top.Settings, client)
		return
	}

	logger.Debug("application starting")

	args := os.Args[1:] // strip first arg `fastly`
	arg, counter := flags.Check(args)

	switch arg {
	case "diff":
		f.Top.Diff.Parse(args[counter:])
		commands.Diff(f, client)
	case "upload":
		f.Top.Upload.Parse(args[counter:])
		commands.Upload(f, client)
	default:
		fmt.Printf("%v is not valid command.\n", arg)
		os.Exit(1)
	}
}
