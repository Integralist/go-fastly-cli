// Standalone is a package that defines behaviour which relies on a single top
// level flag, and not nested flags or sub commands.

package standalone

import (
	"fmt"
	"strconv"

	"github.com/integralist/go-fastly-cli/common"
	"github.com/sethvargo/go-fastly/fastly"
)

// ActivateVersion activates the specified Fastly service version
func ActivateVersion(version, service string, client *fastly.Client) {
	v, err := strconv.Atoi(version)
	if err != nil {
		fmt.Println(err)
		common.Failure()
	}

	_, err = client.ActivateVersion(&fastly.ActivateVersionInput{
		Service: service,
		Version: v,
	})
	if err != nil {
		fmt.Printf("\nThere was a problem activating version %s\n\n%s", common.Yellow(version), common.Red(err))
		common.Failure()
	}

	fmt.Printf("\nService '%s' now has version '%s' activated\n\n", common.Yellow(service), common.Green(version))
}

// ValidateVersion validates the specified Fastly service version
func ValidateVersion(version, service string, client *fastly.Client) {
	v, err := strconv.Atoi(version)
	if err != nil {
		fmt.Println(err)
		common.Failure()
	}

	valid, msg, err := client.ValidateVersion(&fastly.ValidateVersionInput{
		Service: service,
		Version: v,
	})
	if err != nil {
		fmt.Printf("\nThere was a problem validating version %s\n\n%s", common.Yellow(version), common.Red(err))
		common.Failure()
	}

	var validColour, details string

	validColour = common.Green(valid)

	if valid == false {
		validColour = common.Red(valid)
		details = common.Red(msg)
	}

	fmt.Printf("\nService '%s' valid? %s\n\n%s", common.Yellow(service), validColour, details)
}

// PrintLatestSettings sends the sepecified service version settings to stdout
func PrintLatestSettings(serviceID string, client *fastly.Client) {
	latestVersion, err := common.GetLatestVCLVersion(serviceID, client)
	if err != nil {
		fmt.Println(err)
		common.Failure()
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
		common.Failure()
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
