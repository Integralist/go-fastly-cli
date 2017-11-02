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
	Help, HelpShort, Debug, Version                                              *bool
	Token, Service, Directory, Match, Skip, Status, Activate, Validate, Settings *string
	Delete, Diff, List, Upload                                                   *flag.FlagSet
}

// SubCommandFlags defines the settings for the subcommands
type SubCommandFlags struct {
	CloneVersion     *string
	UploadVersion    *string
	UseLatestVersion *bool
	VclDeleteVersion *string
	VclListVersion   *string
	VclName          *string
	VclVersion       *string
}

// Flags defines type of structure returned to user
type Flags struct {
	Top TopLevelFlags
	Sub SubCommandFlags
}

// New returns defined flags
func New() Flags {
	topLevelFlags := TopLevelFlags{
		Activate:  flag.String("activate", "", "specify Fastly service version to activate"),
		Debug:     flag.Bool("debug", false, "show any error/diff output + debug logs"),
		Delete:    flag.NewFlagSet("delete", flag.ExitOnError),
		Diff:      flag.NewFlagSet("diff", flag.ExitOnError),
		Directory: flag.String("dir", os.Getenv("VCL_DIRECTORY"), "vcl directory to compare files against"),
		Help:      flag.Bool("help", false, "show available flags"),
		HelpShort: flag.Bool("h", false, "show available flags"),
		List:      flag.NewFlagSet("list", flag.ExitOnError),
		Match:     flag.String("match", "", "regex for matching vcl directories (will also try: VCL_MATCH_PATH)"),
		Service:   flag.String("service", os.Getenv("FASTLY_SERVICE_ID"), "your service id (fallback: FASTLY_SERVICE_ID)"),
		Settings:  flag.String("settings", "", "get settings (Default TTL & Host) for specified Fastly service version (version number or latest)"),
		Skip:      flag.String("skip", "^____", "regex for skipping vcl directories (will also try: VCL_SKIP_PATH)"),
		Status:    flag.String("status", "", "retrieve status for the specified Fastly service 'version' (try: 'latest')"),
		Token:     flag.String("token", os.Getenv("FASTLY_API_TOKEN"), "your fastly api token (fallback: FASTLY_API_TOKEN)"),
		Upload:    flag.NewFlagSet("upload", flag.ExitOnError),
		Validate:  flag.String("validate", "", "specify Fastly service version to validate"),
		Version:   flag.Bool("version", false, "show application version"),
	}

	flag.Parse()

	return Flags{
		Top: topLevelFlags,
		Sub: subCommands(topLevelFlags),
	}
}

func subCommands(t TopLevelFlags) SubCommandFlags {
	return SubCommandFlags{
		CloneVersion:     t.Upload.String("clone", "", "specify Fastly service version to clone from before uploading to"),
		UploadVersion:    t.Upload.String("version", "", "specify non-active Fastly service 'version' to upload to"),
		UseLatestVersion: t.Upload.Bool("latest", false, "use latest Fastly service version to upload to (presumes not activated)"),
		VclDeleteVersion: t.Delete.String("version", "", "specify Fastly service version to delete VCL file from"),
		VclListVersion:   t.List.String("version", "", "specify Fastly service version to list VCL files from"),
		VclName:          t.Delete.String("name", "", "specify VCL filename to delete"),
		VclVersion:       t.Diff.String("version", "", "specify Fastly service version to verify against"),
	}
}

// Check determines if a flag was specified before the subcommand
// then returns the subcommand argument value based on the correct index
// followed by the index of where the subcommand's flags start in the args list
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

		if arg == "delete" || arg == "diff" || arg == "list" || arg == "upload" {
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
