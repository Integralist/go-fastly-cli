package commands

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/integralist/go-fastly-cli/common"
	"github.com/integralist/go-fastly-cli/flags"

	fastly "github.com/sethvargo/go-fastly"
)

// Upload takes specified list of files and creates new remote version
// if upload fails it'll attempt uploading over existing remote version
func Upload(f flags.Flags, client *fastly.Client) {
	checkIncorrectFlagConfiguration(f)
	configureSkipMatch(f)

	// store value rather than dereference pointer multiple times later
	fastlyServiceID = *f.Top.Service

	// the acquireVersion function checks if we should...
	//
	// 		A. clone the specified version before uploading files: `-clone`
	// 		B. upload files to the specified version: `-version`
	// 		C. upload files to the latest version: `-latest`
	// 		D. clone the latest version available
	selectedVersion, err := acquireVersion(f, client)
	if err != nil {
		fmt.Println(err)
		common.Failure()
	}

	processFiles(selectedVersion, uploadVCL, handleResponse, f, client)
}

func checkIncorrectFlagConfiguration(f flags.Flags) {
	if *f.Sub.CloneVersion != "" && *f.Sub.UploadVersion != "" {
		fmt.Println("Please do not provide both -clone-version and -upload-version flags")
		common.Failure()
	}
}

func cloneFromVersion(version int, client *fastly.Client) (*fastly.Version, error) {
	clonedVersion, err := client.CloneVersion(&fastly.CloneVersionInput{
		Service: fastlyServiceID,
		Version: version,
	})
	if err != nil {
		return nil, err
	}

	return clonedVersion, nil
}

func acquireVersion(f flags.Flags, client *fastly.Client) (int, error) {
	// clone from specified version and upload to that
	if *f.Sub.CloneVersion != "" {
		cloneVersion, err := strconv.Atoi(*f.Sub.CloneVersion)
		if err != nil {
			return 0, err
		}

		clonedVersion, err := cloneFromVersion(cloneVersion, client)
		if err != nil {
			return 0, err
		}

		fmt.Printf("Successfully created new version %d from existing version %s\n\n", clonedVersion.Number, *f.Sub.CloneVersion)
		return clonedVersion.Number, nil
	}

	// upload to the specified version (it can't be activated)
	if *f.Sub.UploadVersion != "" {
		uploadVersion, err := strconv.Atoi(*f.Sub.UploadVersion)
		if err != nil {
			return 0, err
		}

		getVersion, err := client.GetVersion(&fastly.GetVersionInput{
			Service: fastlyServiceID,
			Version: uploadVersion,
		})
		if err != nil {
			return 0, err
		}

		if getVersion.Active {
			return 0, fmt.Errorf("Sorry, the specified version is already activated")
		}

		return uploadVersion, nil
	}

	latestVersion, err := common.GetLatestVCLVersion(*f.Top.Service, client)
	if err != nil {
		return 0, err
	}

	// upload to the latest version
	// note: latest version must not be activated already
	if *f.Sub.UseLatestVersion {
		getVersion, err := client.GetVersion(&fastly.GetVersionInput{
			Service: fastlyServiceID,
			Version: latestVersion,
		})
		if err != nil {
			return 0, err
		}

		if getVersion.Active {
			return 0, fmt.Errorf("Sorry, the latest version is already activated")
		}

		return latestVersion, nil
	}

	// otherwise clone the latest version and upload to that
	clonedVersion, err := cloneFromVersion(latestVersion, client)
	if err != nil {
		return 0, err
	}

	fmt.Printf("Successfully created new version %d from latest version %d\n\n", clonedVersion.Number, latestVersion)
	return clonedVersion.Number, nil
}

func uploadVCL(selectedVersion int, path string, client *fastly.Client, ch chan vclResponse) {
	defer wg.Done()

	name := extractName(path)
	content, err := getLocalVCL(path)

	if err != nil {
		ch <- vclResponse{
			Path:    path,
			Name:    name,
			Content: fmt.Sprintf("get local vcl error: %s", err),
			Error:   true,
		}
	} else {
		// First check if the local file exists already on the remote
		_, err := client.GetVCL(&fastly.GetVCLInput{
			Service: fastlyServiceID,
			Version: selectedVersion,
			Name:    name,
		})

		if err != nil {
			// If the file DOESNT exist, then we'll create it
			fmt.Printf("\n%+v\n> will attempt to create the file %s.vcl\n\n", err, name)

			vclFile, err := client.CreateVCL(&fastly.CreateVCLInput{
				Service: fastlyServiceID,
				Version: selectedVersion,
				Name:    name,
				Content: content,
			})

			if err != nil {
				fmt.Printf("There was an error creating the file '%s':\n%s\n\n", common.Yellow(name), common.Red(err))
			} else {
				ch <- vclResponse{
					Path:    path,
					Name:    name,
					Content: vclFile.Content,
					Error:   false,
				}
			}
		} else {
			// If the file DOES exist, then we'll upload our version on top of it
			vclFileUpdate, updateErr := client.UpdateVCL(&fastly.UpdateVCLInput{
				Service: fastlyServiceID,
				Version: selectedVersion,
				Name:    name,
				Content: content,
			})
			if updateErr != nil {
				ch <- vclResponse{
					Path:    path,
					Name:    name,
					Content: fmt.Sprintf("error: %s", updateErr),
					Error:   true,
				}
			} else {
				ch <- vclResponse{
					Path:    path,
					Name:    name,
					Content: vclFileUpdate.Content,
					Error:   false,
				}
			}
		}
	}
}

func getLocalVCL(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func handleResponse(vr vclResponse, debug bool, selectedVersion int) {
	if vr.Error {
		fmt.Printf("The file '%s' didn't upload to version '%d' because of the following error:\n\t%s\n\n", common.Yellow(vr.Name), selectedVersion, common.Red(vr.Content))
	} else {
		fmt.Printf("The file '%s' in version '%s' was updated successfully\n", common.Green(vr.Name), common.Yellow(selectedVersion))
	}
}
