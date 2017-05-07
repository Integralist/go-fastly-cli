package commands

import (
	"flags"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	fastly "github.com/sethvargo/go-fastly"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"package": "commands",
	})
}

// regex used to define user specific filtering
var dirSkipRegex *regexp.Regexp
var dirMatchRegex *regexp.Regexp

// globals needed for sharing between functions
var fastlyServiceID string
var latestVersion string

// the WaitGroup is used when processing files with multiple goroutine
var wg sync.WaitGroup

// list of VCL files to process
var vclFiles []string

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

// data structure for Fastly API response
type vclResponse struct {
	Path    string
	Name    string
	Content string
	Error   bool
}

type fileProcessor func(string, string, *fastly.Client, chan vclResponse)
type responseProcessor func(vclResponse, bool, string)

// function called by filepath.Walk
func aggregate(path string, f os.FileInfo, err error) error {
	if validPathDefaults(path) && validPathUserDefined(path) && !invalidPathUserDefined(path) {
		vclFiles = append(vclFiles, path)
	}

	return nil
}

func validPathDefaults(path string) bool {
	return !strings.Contains(path, ".git") && strings.Contains(path, ".vcl")
}

func validPathUserDefined(path string) bool {
	return dirMatchRegex.MatchString(path)
}

func invalidPathUserDefined(path string) bool {
	return dirSkipRegex.MatchString(path)
}

func extractName(path string) string {
	_, file := filepath.Split(path)
	return strings.Split(file, ".")[0]
}

func getLatestVCLVersion(client *fastly.Client) (string, error) {
	// we have to get all the versions and then sort them to find the actual latest
	listVersions, err := client.ListVersions(&fastly.ListVersionsInput{
		Service: fastlyServiceID,
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

func configureSkipMatch(f flags.Flags) {
	// check if env vars are defined. If so, use them to override default values
	envSkipDir := os.Getenv("VCL_SKIP_DIRECTORY")
	envMatchDir := os.Getenv("VCL_MATCH_DIRECTORY")

	if envSkipDir != "" {
		*f.Top.Skip = envSkipDir
	}

	if envMatchDir != "" {
		*f.Top.Match = envMatchDir
	}

	// compile regex with provided values or the defaults (see vars.go for usage)
	dirSkipRegex, _ = regexp.Compile(*f.Top.Skip)
	dirMatchRegex, _ = regexp.Compile(*f.Top.Match)
}

// processFiles first aggregates all available local VCL files
// lookup is based on `-dir` or `VCL_DIRECTORY`
// then for each VCL file it spins up new goroutine
// the goroutine behaviour is provided by the caller
// finally, it ranges over the buffered channel of data
// each item in the channel is processed dependant on the caller provided function
func processFiles(selectedVersion string, fp fileProcessor, rp responseProcessor, f flags.Flags, client *fastly.Client) {
	walkError := filepath.Walk(*f.Top.Directory, aggregate)
	if walkError != nil {
		fmt.Printf("filepath.Walk() returned an error: %v\n", walkError)
	}

	logger.WithFields(logrus.Fields{
		"files":  vclFiles,
		"length": len(vclFiles),
	}).Debug("aggregated files")

	ch := make(chan vclResponse, len(vclFiles))

	for _, vclPath := range vclFiles {
		wg.Add(1)
		go fp(selectedVersion, vclPath, client, ch)
	}
	wg.Wait()

	close(ch)

	// reset slice so no data shared between subcommands
	vclFiles = []string{}

	for vclFile := range ch {
		rp(vclFile, *f.Top.Debug, selectedVersion)
	}
}
