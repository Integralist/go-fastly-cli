package main

import (
	"flag"
	"flags"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// AppVersion is the application version
const AppVersion = "0.0.1"

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

	if *f.Help == true {
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *f.Version == true {
		fmt.Println(AppVersion)
		os.Exit(1)
	}

	if *f.Debug == true {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logger.Debug("application starting")

	args := os.Args[1:] // strip first arg `fastly`
	arg, counter := flags.Check(args)

	switch arg {
	case "diff":
		f.Diff.Parse(args[counter:])
	case "upload":
		f.Upload.Parse(args[counter:])
	default:
		fmt.Printf("%v is not valid command.\n", arg)
		os.Exit(1)
	}
}
