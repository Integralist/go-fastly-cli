package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/integralist/go-fastly-cli/flags"

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

// data structure for Fastly API response
type vclResponse struct {
	Path    string
	Name    string
	Content string
	Error   bool
}

type fileProcessor func(int, string, *fastly.Client, chan vclResponse)
type responseProcessor func(vclResponse, bool, int)

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

func configureSkipMatch(f flags.Flags) {
	skipDefault := "^____"
	matchDefault := ""

	skipRegex := *f.Top.Skip
	matchRegex := *f.Top.Match

	if skipRegex == skipDefault && os.Getenv("VCL_SKIP_PATH") != "" {
		skipRegex = os.Getenv("VCL_SKIP_PATH")
	}
	if matchRegex == matchDefault && os.Getenv("VCL_MATCH_PATH") != "" {
		matchRegex = os.Getenv("VCL_MATCH_PATH")
	}

	logger.WithFields(logrus.Fields{
		"skip":  skipRegex,
		"match": matchRegex,
	}).Debug("compile skip/match regexes")

	// compile regex with provided values or the defaults (see vars.go for usage)
	dirSkipRegex, _ = regexp.Compile(skipRegex)
	dirMatchRegex, _ = regexp.Compile(matchRegex)
}

// processFiles first aggregates all available local VCL files
// lookup is based on `-dir` or `VCL_DIRECTORY`
// then for each VCL file it spins up new goroutine
// the goroutine behaviour is provided by the caller
// finally, it ranges over the buffered channel of data
// each item in the channel is processed dependant on the caller provided function
func processFiles(selectedVersion int, fp fileProcessor, rp responseProcessor, f flags.Flags, client *fastly.Client) {
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
