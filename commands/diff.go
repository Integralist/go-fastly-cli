package commands

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/integralist/go-fastly-cli/common"
	"github.com/integralist/go-fastly-cli/flags"

	"github.com/fatih/color"
	fastly "github.com/sethvargo/go-fastly"
	"github.com/sirupsen/logrus"
)

// Diff compares local VCL to the specificed remote service vcl version
func Diff(f flags.Flags, client *fastly.Client) {
	configureSkipMatch(f)

	// store value rather than dereference pointer multiple times later
	fastlyServiceID = *f.Top.Service

	var (
		selectedVersion int
		err             error
	)

	if *f.Sub.VclVersion != "" {
		selectedVersion, err = strconv.Atoi(*f.Sub.VclVersion)
		if err != nil {
			fmt.Println(err)
			common.Failure()
		}
	} else {
		latestVersion, err := common.GetLatestVCLVersion(*f.Top.Service, client)
		if err != nil {
			fmt.Println(err)
			common.Failure()
		}
		selectedVersion = latestVersion
	}

	processFiles(selectedVersion, getVCL, processDiff, f, client)
}

func getVCL(selectedVersion int, path string, client *fastly.Client, ch chan vclResponse) {
	defer wg.Done()

	logger.WithFields(logrus.Fields{
		"channel":  fmt.Sprintf("%p", ch),
		"length":   len(ch),
		"capacity": cap(ch),
	}).Debug("channel view - start of getVCL")

	name := extractName(path)

	logger.WithFields(logrus.Fields{
		"path": path,
		"name": name,
	}).Debug("file processor")

	vclFile, err := client.GetVCL(&fastly.GetVCLInput{
		Service: fastlyServiceID,
		Version: selectedVersion,
		Name:    name,
	})

	if err != nil {
		errMsg := fmt.Sprintf("error: %s", err)

		logger.WithFields(logrus.Fields{
			"name":  name,
			"error": errMsg,
		}).Debug("error retrieving vcl file from fastly")

		ch <- vclResponse{
			Path:    path,
			Name:    name,
			Content: errMsg,
			Error:   true,
		}

		logger.WithFields(logrus.Fields{
			"channel":  fmt.Sprintf("%p", ch),
			"length":   len(ch),
			"capacity": cap(ch),
		}).Debug("channel view - failed retrieval")
	} else {
		logger.WithField("name", name).Debug("successfully retrieved vcl file from fastly")

		ch <- vclResponse{
			Path:    path,
			Name:    name,
			Content: vclFile.Content,
			Error:   false,
		}

		logger.WithFields(logrus.Fields{
			"channel":  fmt.Sprintf("%p", ch),
			"length":   len(ch),
			"capacity": cap(ch),
		}).Debug("channel view - successful retrieval")
	}
}

func processDiff(vr vclResponse, debug bool, selectedVersion int) {
	var (
		err    error
		cmdOut []byte
	)
	cmdName := "diff"
	cmdArgs := []string{
		"--ignore-all-space",
		"--ignore-blank-lines",
		"--ignore-matching-lines",
		"^[[:space:]]\\+#",
		"-", // the dash (-) indicates that the first file comes from stdin
		vr.Path,
	}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdin = strings.NewReader(vr.Content)

	if cmdOut, err = cmd.Output(); err != nil {
		color.Red("\nThere was a difference between the version (%d) of '%s' and the version found locally\n\t%s\n", selectedVersion, vr.Name, vr.Path)

		if debug == true {
			fmt.Printf("\n%s\n", string(cmdOut))
		}
	} else {
		color.Green("\nNo difference between the version (%d) of '%s' and the version found locally\n\t%s\n", selectedVersion, vr.Name, vr.Path)
	}
}
