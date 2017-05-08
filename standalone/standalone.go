package standalone

import (
	"common"
	"flags"
	"fmt"
	"os"
	"strconv"

	"github.com/fatih/color"
	fastly "github.com/sethvargo/go-fastly"
)

// TODO: move to common package
// useful colour settings for printing messages
var yellow = color.New(color.FgYellow).SprintFunc()
var red = color.New(color.FgRed).SprintFunc()
var green = color.New(color.FgGreen).SprintFunc()

// ActivateVersion activates the specified Fastly service version
func ActivateVersion(f flags.Flags, client *fastly.Client) {
	v, err := strconv.Atoi(*f.Top.Activate)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = client.ActivateVersion(&fastly.ActivateVersionInput{
		Service: *f.Top.Service,
		Version: v,
	})
	if err != nil {
		fmt.Printf("\nThere was a problem activating version %s\n\n%s", yellow(*f.Top.Activate), red(err))
		os.Exit(1)
	}
	fmt.Printf("\nService '%s' now has version '%s' activated\n\n", yellow(*f.Top.Service), green(*f.Top.Activate))
}

// PrintLatestSettings sends the sepecified service version settings to stdout
func PrintLatestSettings(serviceID string, client *fastly.Client) {
	latestVersion, err := common.GetLatestVCLVersion(serviceID, client)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	PrintSettingsFor(serviceID, latestVersion, client)
}

// PrintSettingsFor sends the sepecified service version settings to stdout
func PrintSettingsFor(serviceID, serviceVersion string, client *fastly.Client) {
	v, err := strconv.Atoi(serviceVersion)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	settings, err := client.GetSettings(&fastly.GetSettingsInput{
		Service: serviceID,
		Version: v,
	})
	if err != nil {
		fmt.Printf("\nThere was a problem getting the settings for version %s\n\n%s", yellow(serviceVersion), red(err))
		os.Exit(1)
	}

	fmt.Printf(
		"\nDefault Host: %s\nDefault TTL: %d (seconds)\n\n",
		settings.DefaultHost,
		settings.DefaultTTL,
	)
}

// GetLatestServiceVersionStatus returns the latest Fastly service version and its status
func GetLatestServiceVersionStatus(serviceID string, client *fastly.Client) (string, error) {
	latestVersion, err := common.GetLatestVCLVersion(serviceID, client)
	if err != nil {
		return "", err
	}

	status, err := GetStatusForVersion(serviceID, latestVersion, client)
	if err != nil {
		return "", err
	}

	return status, nil
}

// GetStatusForVersion returns the status of the specified Fastly service version
func GetStatusForVersion(serviceID, statusVersion string, client *fastly.Client) (string, error) {
	v, err := strconv.Atoi(statusVersion)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	versionStatus, err := client.GetVersion(&fastly.GetVersionInput{
		Service: serviceID,
		Version: v,
	})
	if err != nil {
		msg := "\nThere was a problem getting the status for version %s\n\n%s\n\n"
		return "", fmt.Errorf(msg, yellow(statusVersion), red(err))
	}

	activated := green("not activated")
	if versionStatus.Active {
		activated = red("already activated")
	}

	msg := "\nService '%s' version '%s' is '%s'\n\n"
	status := fmt.Sprintf(msg, yellow(serviceID), yellow(statusVersion), activated)

	return status, nil
}
