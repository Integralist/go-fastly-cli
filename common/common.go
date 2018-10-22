// Common is a package containing functions that used across mutliple packages.

package common

import (
	"fmt"
	"os"
	"sort"

	"github.com/fatih/color"
	fastly "github.com/sethvargo/go-fastly"
)

// useful colour settings for printing messages...

// Yellow colours stdout to be yellow
var Yellow = color.New(color.FgYellow).SprintFunc()

// Red colours stdout to be red
var Red = color.New(color.FgRed).SprintFunc()

// Green colours stdout to be green
var Green = color.New(color.FgGreen).SprintFunc()

// fastly API doesn't return sorted data
// so we have to manually sort the data ourselves
type version struct {
	Number  int
	Version *fastly.Version
}
type wrappedVersions []version

// satisfy the Sort interface
func (v wrappedVersions) Len() int      { return len(v) }
func (v wrappedVersions) Swap(i, j int) { v[i], v[j] = v[j], v[i] }
func (v wrappedVersions) Less(i, j int) bool {
	return v[i].Number < v[j].Number
}

// Success stops processing and returns a zero exit code
func Success() {
	os.Exit(0)
}

// Failure stops processing and returns an error exit code
func Failure() {
	os.Exit(1)
}

// GetLatestVCLVersion returns latest fastly service version
// This service version isn't necessarily the currently active version
func GetLatestVCLVersion(serviceID string, client *fastly.Client) (int, error) {
	// we have to get all the versions and then sort them to find the actual latest
	listVersions, err := client.ListVersions(&fastly.ListVersionsInput{
		Service: serviceID,
	})
	if err != nil {
		return 0, fmt.Errorf("There was a problem getting the version list:\n\n%s", Red(err))
	}

	wv := wrappedVersions{}
	for _, v := range listVersions {
		wv = append(wv, version{v.Number, v})
	}
	sort.Sort(wv)

	return wv[len(wv)-1].Number, nil
}
