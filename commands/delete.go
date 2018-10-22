package commands

import (
	"fmt"
	"strconv"

	"github.com/integralist/go-fastly-cli/common"
	"github.com/integralist/go-fastly-cli/flags"

	fastly "github.com/sethvargo/go-fastly"
)

// Delete specified VCL file in the remote service version
func Delete(f flags.Flags, client *fastly.Client) {
	var serviceVersion string
	var selectedVersion int

	deleteVCL := *f.Sub.VclName

	if deleteVCL == "" {
		fmt.Println("You must provide a VCL name\n  e.g. -name test_file")
		common.Failure()
	}

	deleteVersion := *f.Sub.VclDeleteVersion
	fastlyServiceID = *f.Top.Service

	if deleteVersion != "" {
		serviceVersion = deleteVersion
	} else {
		var err error
		selectedVersion, err = common.GetLatestVCLVersion(fastlyServiceID, client)
		if err != nil {
			fmt.Println("Sorry, we were unable to acquire the latest service version")
			fmt.Println("Please try again, or provide a specific version by using the -version flag")
			common.Failure()
		}
	}

	// If we're not the type default, then we have the latest service version
	if selectedVersion != 0 {
		fmt.Printf("You didn't provide a specific service version, so we'll use the latest one: %s\n", common.Yellow(selectedVersion))
	} else {
		// Otherwise the user provided a service version, which we need to convert
		var err error
		selectedVersion, err = strconv.Atoi(serviceVersion)
		if err != nil {
			fmt.Printf("Unable to convert provided version:\n\t%+v\n", err)
			common.Failure()
		}
	}

	err := client.DeleteVCL(&fastly.DeleteVCLInput{
		Service: fastlyServiceID,
		Version: selectedVersion,
		Name:    deleteVCL,
	})

	if err != nil {
		fmt.Printf("\nUnable to delete the specified VCL file from version: %s\n\n", common.Yellow(selectedVersion))
		fmt.Printf("Error:\n%s", common.Red(err))
	}

	common.Success()
}
