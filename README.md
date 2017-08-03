# go-fastly-cli

CLI tool for uploading and diffing local/remote Fastly VCL

> Note: this supersedes the individual tools: [ero](https://github.com/Integralist/ero) and [lataa](https://github.com/Integralist/lataa)

## Install

```bash
go get github.com/integralist/go-fastly-cli
```

## Usage

```bash
fastly <flags> [diff <options>]
fastly <flags> [upload <options>]
```

Flags:

```bash
fastly -help

  -activate string
        specify Fastly service version to activate
  -debug
        show any debug logs and subcommand specific information
  -dir string
        the directory where your vcl files are located
  -help
        show available flags
  -match string
        regex for matching vcl directories (fallback: VCL_MATCH_PATH)
  -service string
        your Fastly service id (fallback: FASTLY_SERVICE_ID) 
  -settings string
        get settings for the specified Fastly service version (try: 'latest')
  -skip string
        regex for skipping vcl directories (will also try: VCL_SKIP_PATH) 
  -status string
        get status for the specified Fastly service version (try: 'latest')
  -token string
        your fastly api token (fallback: FASTLY_API_TOKEN) 
  -version
        show application version
```

Diff Options:

```bash
fastly diff -help

Usage of diff:
  -version string
        specify Fastly service version to verify against
```

Upload Options:

```bash
fastly upload -help

Usage of upload:
  -clone string
        specify a Fastly service version to clone from (files will upload to it)
  -latest
        use latest Fastly service version to upload to (presumes not activated)
  -version string
        specify non-active Fastly service version to upload to
```

## Environment Variables

The use of environment variables help to reduce the amount of flags required by the `fastly` CLI tool.

For example, I always diff against a 'stage' service in our Fastly account and so I don't want to have to put in the same credentials all the time.

Below is a list of environment variables this tool supports:

* `FASTLY_API_TOKEN` (`-token`)
* `FASTLY_SERVICE_ID` (`-service`)
* `VCL_DIRECTORY` (`-dir`)
* `VCL_MATCH_PATH` (`-match`)
* `VCL_SKIP_PATH` (`-skip`)

> Use the relevant CLI flags to override these values

## Makefile

To compile binaries for multiple OS architectures:

```bash
make compile
```

To start up a dockerized development environment (inc. Vim):

```bash
make dev
```

To install a local binary for testing (darwin):

```bash
make install
```

To remove all compiled binaries, vim files and containers:

```bash
make clean
```

## Examples

> Note: all examples presume `FASTLY_API_TOKEN`/`FASTLY_SERVICE_ID` env vars set

```bash
# view status for the latest service version
fastly -status latest

# view status for the specified service version
fastly -status 123

# view settings for the latest service version
fastly -settings latest

# view settings for the specified service version
fastly -settings 123

# activate specified service version
fastly -activate 123

# diff local vcl files against the lastest remote versions
fastly diff

# diff local vcl files against the specific remote versions
fastly diff -version 123

# enable debug mode
# this will mean debug logs are displayed
# for 'diff' subcommand: also display per file diff
fastly -debug diff -version 123

# upload local files to remote service version
fastly upload -version 123

# token and service explicitly set to override env vars
fastly -service xxx -token xxx upload -version 123

# clone specified service version and upload local files to it
fastly upload -clone 123

# upload local files to the latest remote service version
fastly upload -latest

# clone latest service version available and upload local files to it
fastly upload
```

## TODO

* Ability to 'dry run' a command (to see what files are affected, e.g. what files will be uploaded and where)
* Ability to diff two remote services (not just local against a remote)
* Ability to upload an individual file (not just pattern matched list of files)
* Ability to display all available services (along with their ID)
* Test Suite
