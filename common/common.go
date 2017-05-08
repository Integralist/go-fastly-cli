package common

import (
	"sort"
	"strconv"

	fastly "github.com/sethvargo/go-fastly"
)

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

// GetLatestVCLVersion returns latest fastly service version
// This service version isn't necessarily the currently active version
func GetLatestVCLVersion(serviceVersion string, client *fastly.Client) (string, error) {
	// we have to get all the versions and then sort them to find the actual latest
	listVersions, err := client.ListVersions(&fastly.ListVersionsInput{
		Service: serviceVersion,
	})
	if err != nil {
		return "", err
	}

	wv := wrappedVersions{}
	for _, v := range listVersions {
		wv = append(wv, version{v.Number, v})
	}
	sort.Sort(wv)

	return strconv.Itoa(wv[len(wv)-1].Number), nil
}
