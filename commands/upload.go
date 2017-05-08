package commands

import (
	"common"
	"flags"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"

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
	// 		A. clone the specified version before uploading files: `-clone-version`
	// 		B. upload files to the specified version: `-upload-version`
	// 		C. upload files to the latest version: `-use-latest-version`
	// 		D. clone the latest version if it's already activated
	selectedVersion, err := acquireVersion(f, client)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	processFiles(selectedVersion, uploadVCL, handleResponse, f, client)
}

func checkIncorrectFlagConfiguration(f flags.Flags) {
	if *f.Sub.CloneVersion != "" && *f.Sub.UploadVersion != "" {
		fmt.Println("Please do not provide both -clone-version and -upload-version flags")
		os.Exit(1)
	}
}

func cloneFromVersion(version string, client *fastly.Client) (*fastly.Version, error) {
	v, err := strconv.Atoi(version)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	clonedVersion, err := client.CloneVersion(&fastly.CloneVersionInput{
		Service: fastlyServiceID,
		Version: v,
	})
	if err != nil {
		return nil, err
	}

	return clonedVersion, nil
}

func acquireVersion(f flags.Flags, client *fastly.Client) (string, error) {
	// clone from specified version and upload to that
	if *f.Sub.CloneVersion != "" {
		clonedVersion, err := cloneFromVersion(*f.Sub.CloneVersion, client)
		if err != nil {
			return "", err
		}

		fmt.Printf("Successfully created new version %d from existing version %s\n\n", clonedVersion.Number, *f.Sub.CloneVersion)
		return strconv.Itoa(clonedVersion.Number), nil
	}

	// upload to the specified version (it can't be activated)
	if *f.Sub.UploadVersion != "" {
		v, err := strconv.Atoi(*f.Sub.UploadVersion)
		if err != nil {
			return "", err
		}

		getVersion, err := client.GetVersion(&fastly.GetVersionInput{
			Service: fastlyServiceID,
			Version: v,
		})
		if err != nil {
			return "", err
		}

		if getVersion.Active {
			return "", fmt.Errorf("Sorry, the specified version is already activated")
		}

		return *f.Sub.UploadVersion, nil
	}

	// upload to the latest version
	// note: latest version must not be activated already
	if *f.Sub.UseLatestVersion {
		latestVersion, err := common.GetLatestVCLVersion(*f.Top.Service, client)
		if err != nil {
			return "", err
		}

		v, err := strconv.Atoi(latestVersion)
		if err != nil {
			return "", err
		}

		getVersion, err := client.GetVersion(&fastly.GetVersionInput{
			Service: fastlyServiceID,
			Version: v,
		})
		if err != nil {
			return "", err
		}

		if getVersion.Active {
			fmt.Println("Sorry, the latest version is already activated")
			return "", err
		}

		return latestVersion, nil
	}

	// otherwise clone the latest version and upload to that
	clonedVersion, err := cloneFromVersion(latestVersion, client)
	if err != nil {
		return "", err
	}

	fmt.Printf("Successfully created new version %d from latest version %s\n\n", clonedVersion.Number, latestVersion)
	return strconv.Itoa(clonedVersion.Number), nil
}

func uploadVCL(selectedVersion, path string, client *fastly.Client, ch chan vclResponse) {
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
		v, err := strconv.Atoi(selectedVersion)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		vclFile, err := client.CreateVCL(&fastly.CreateVCLInput{
			Service: fastlyServiceID,
			Version: v,
			Name:    name,
			Content: content,
		})

		if err != nil {
			fmt.Printf("There was an error creating the file '%s':\n%s\nWe'll now try updating this file instead of creating it\n\n", common.Yellow(name), common.Red(err))

			vclFileUpdate, updateErr := client.UpdateVCL(&fastly.UpdateVCLInput{
				Service: fastlyServiceID,
				Version: v,
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
		} else {
			ch <- vclResponse{
				Path:    path,
				Name:    name,
				Content: vclFile.Content,
				Error:   false,
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

func handleResponse(vr vclResponse, debug bool, selectedVersion string) {
	if vr.Error {
		fmt.Printf("Whoops, the file '%s' didn't upload to version '%s' because of the following error:\n\t%s\n\n", common.Yellow(vr.Name), selectedVersion, common.Red(vr.Content))
	} else {
		fmt.Printf("Yay, the file '%s' in version '%s' was updated successfully\n", common.Green(vr.Name), common.Yellow(selectedVersion))
	}
}
