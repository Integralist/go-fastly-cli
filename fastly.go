package main

import (
	"flag"
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
const appVersion = "0.0.1"

var logger *logrus.Entry

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)
	logger = logrus.WithFields(logrus.Fields{
		"package": "main",
	})
}

func printSubCommands() {
	diff := "\n  fastly diff\n\tview a diff between your local files and the remote versions\n\te.g. fastly diff -version 123"
	list := "\n\n  fastly list\n\tlist all vcl files found within specified remote service version\n\te.g. fastly list -version 123"
	upload := "\n\n  fastly upload\n\tupload local files to your remote service version\n\te.g. fastly upload -version 123"
	divider := "\n\n -------------------------------------------------------------------\n\n"
	fmt.Printf("%s%s%s%s", diff, list, upload, divider)
}

func main() {
	f := flags.New()

	if len(os.Args) < 2 {
		printSubCommands()
		flag.PrintDefaults()
		os.Exit(1)
	}

	logger.Debug("flags initialised, application starting")

	if *f.Top.Help == true || *f.Top.HelpShort == true {
		printSubCommands()
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

	if *f.Top.Settings == "latest" {
		standalone.PrintLatestSettings(*f.Top.Service, client)
		return
	}

	if *f.Top.Settings != "" {
		settingsVersion, err := strconv.Atoi(*f.Top.Settings)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		standalone.PrintSettingsFor(*f.Top.Service, settingsVersion, client)
		return
	}

	args := os.Args[1:] // strip first arg `fastly`
	arg, counter := flags.Check(args)

	switch arg {
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
