package commands

import (
	"fmt"
	"os"
	"strconv"

	"github.com/integralist/go-fastly-cli/common"
	"github.com/integralist/go-fastly-cli/flags"

	fastly "github.com/sethvargo/go-fastly"
)

// List all VCL files found in the remote service version
func List(f flags.Flags, client *fastly.Client) {
	var serviceVersion string
	var selectedVersion int

	fl := *f.Sub.VclListVersion
	fastlyServiceID = *f.Top.Service

	if fl != "" {
		serviceVersion = fl
	} else {
		var err error
		selectedVersion, err = common.GetLatestVCLVersion(fastlyServiceID, client)
		if err != nil {
			fmt.Println("Sorry, we were unable to acquire the latest service version")
			fmt.Println("Please try again, or provide a specific version by using the -version flag")
			os.Exit(2)
		}
	}

	if selectedVersion != 0 {
		fmt.Println("You didn't provide a specific service version, so we'll use the latest one")
	} else {
		var err error
		selectedVersion, err = strconv.Atoi(serviceVersion)
		if err != nil {
			fmt.Printf("Unable to convert provided version:\n\t%+v\n", err)
			os.Exit(1)
		}
	}

	vclFiles, err := client.ListVCLs(&fastly.ListVCLsInput{
		Service: fastlyServiceID,
		Version: selectedVersion,
	})
	if err != nil {
		fmt.Printf("Unable to retrieve list of VCL files for version: %s", common.Yellow(selectedVersion))
		os.Exit(1)
	}

	fmt.Printf("VCL files found for service version: %s\n\n", common.Yellow(selectedVersion))
	for _, f := range vclFiles {
		fmt.Printf("  * %v\n", f.Name)
	}

	os.Exit(1)
}
