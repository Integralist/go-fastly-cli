package flags

import (
	"flag"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

var logger *logrus.Entry

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"package": "flags",
	})
}

// TopLevelFlags defines the common settings across all commands
type TopLevelFlags struct {
	Help, Debug, Version                   *bool
	Token, Service, Directory, Match, Skip *string
	Diff, Upload                           *flag.FlagSet
}

// Flags defines type of structure returned to user
type Flags struct {
	TopLevelFlags
}

// New returns defined flags
func New() Flags {
	topLevelFlags := TopLevelFlags{
		Help:      flag.Bool("help", false, "show available flags"),
		Debug:     flag.Bool("debug", false, "show any error/diff output + debug logs"),
		Version:   flag.Bool("version", false, "show application version"),
		Token:     flag.String("token", os.Getenv("FASTLY_API_TOKEN"), "your fastly api token (fallback: FASTLY_API_TOKEN)"),
		Service:   flag.String("service", os.Getenv("FASTLY_SERVICE_ID"), "your service id (fallback: FASTLY_SERVICE_ID)"),
		Directory: flag.String("dir", os.Getenv("VCL_DIRECTORY"), "vcl directory to compare files against"),
		Match:     flag.String("match", "", "regex for matching vcl directories (will also try: VCL_MATCH_DIRECTORY)"),
		Skip:      flag.String("skip", "^____", "regex for skipping vcl directories (will also try: VCL_SKIP_DIRECTORY)"),
		Diff:      flag.NewFlagSet("diff", flag.ExitOnError),
		Upload:    flag.NewFlagSet("upload", flag.ExitOnError),
	}

	flag.Parse()

	subCommands(topLevelFlags)

	return Flags{
		topLevelFlags,
	}
}

func subCommands(t TopLevelFlags) {
	// TODO: might need to expose these flags in order to use their logic?
	// meaning main package might need to pass instance/data around?
	t.Diff.String("vcl-version", "", "specify Fastly service 'version' to verify against")
	t.Upload.Bool("use-latest-version", false, "use latest Fastly service version to upload to (presumes not activated)")
	t.Upload.Bool("get-latest-version", false, "get latest Fastly service version and its active status")
	t.Upload.String("get-settings", "", "get settings (Default TTL & Host) for specified Fastly service version (version number or latest)")
	t.Upload.String("clone-version", "", "specify Fastly service 'version' to clone from before uploading to")
	t.Upload.String("upload-version", "", "specify non-active Fastly service 'version' to upload to")
	t.Upload.String("activate-version", "", "specify Fastly service 'version' to activate")
	t.Upload.String("get-version-status", "", "retrieve status for the specified Fastly service 'version'")
}

// Check determines if a flag was specified before the subcommand
// Then returns the subcommand argument value based on the correct index
// Followed by the index of where the subcommand's flags start in the args list
func Check(args []string) (string, int) {
	counter := 0
	subcommandSeen := false

	for _, arg := range args {
		if subcommandSeen {
			break
		}

		if strings.HasPrefix(arg, "-") == true {
			counter++
			continue
		}

		if arg == "diff" || arg == "upload" {
			subcommandSeen = true
		} else {
			counter++
		}
	}

	subcommandFlagsIndex := counter + 1

	logger.WithFields(logrus.Fields{
		"args":       args,
		"counter":    counter,
		"subcommand": args[counter],
		"index":      subcommandFlagsIndex,
	}).Debug("subcommand selected")

	return args[counter], subcommandFlagsIndex
}
