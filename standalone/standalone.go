package standalone

import (
	"fmt"
	"os"
	"strconv"

	"github.com/integralist/go-fastly-cli/common"
	"github.com/integralist/go-fastly-cli/flags"

	fastly "github.com/sethvargo/go-fastly"
)

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
		fmt.Printf("\nThere was a problem activating version %s\n\n%s", common.Yellow(*f.Top.Activate), common.Red(err))
		os.Exit(1)
	}
	fmt.Printf("\nService '%s' now has version '%s' activated\n\n", common.Yellow(*f.Top.Service), common.Green(*f.Top.Activate))
}

// ValidateVersion validates the specified Fastly service version
func ValidateVersion(f flags.Flags, client *fastly.Client) {
	v, err := strconv.Atoi(*f.Top.Validate)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	valid, msg, err := client.ValidateVersion(&fastly.ValidateVersionInput{
		Service: *f.Top.Service,
		Version: v,
	})
	if err != nil {
		fmt.Printf("\nThere was a problem validating version %s\n\n%s", common.Yellow(*f.Top.Validate), common.Red(err))
		os.Exit(1)
	}

	var validColour, details string

	validColour = common.Green(valid)

	if valid == false {
		validColour = common.Red(valid)
		details = common.Red(msg)
	}

	fmt.Printf("\nService '%s' valid? %s\n\n%s", common.Yellow(*f.Top.Service), validColour, details)
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
func PrintSettingsFor(serviceID string, serviceVersion int, client *fastly.Client) {
	settings, err := client.GetSettings(&fastly.GetSettingsInput{
		Service: serviceID,
		Version: serviceVersion,
	})
	if err != nil {
		fmt.Printf("\nThere was a problem getting the settings for version %s\n\n%s", common.Yellow(serviceVersion), common.Red(err))
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
func GetStatusForVersion(serviceID string, statusVersion int, client *fastly.Client) (string, error) {
	versionStatus, err := client.GetVersion(&fastly.GetVersionInput{
		Service: serviceID,
		Version: statusVersion,
	})
	if err != nil {
		msg := "\nThere was a problem getting the status for version %s\n\n%s\n\n"
		return "", fmt.Errorf(msg, common.Yellow(statusVersion), common.Red(err))
	}

	activated := common.Green("not activated")
	if versionStatus.Active {
		activated = common.Red("already activated")
	}

	msg := "\nService '%s' version '%s' is '%s'\n\n"
	status := fmt.Sprintf(msg, common.Yellow(serviceID), common.Yellow(statusVersion), activated)

	return status, nil
}
