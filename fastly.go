package main

import (
	"commands"
	"flag"
	"flags"
	"fmt"
	"os"

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

	logger.Debug("application starting")

	args := os.Args[1:] // strip first arg `fastly`
	arg, counter := flags.Check(args)

	switch arg {
	case "diff":
		f.Top.Diff.Parse(args[counter:])
		commands.Diff(f)
	case "upload":
		f.Top.Upload.Parse(args[counter:])
		commands.Upload(f)
	default:
		fmt.Printf("%v is not valid command.\n", arg)
		os.Exit(1)
	}
}
