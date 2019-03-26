# go-fastly-cli

CLI tool for:

* Uploading local VCL to Fastly.
* Diffing local VCL with remote Fastly VCL.
* Listing remote Fastly VCL files.
* Deleting remote Fastly VCL files.
* Creating, Activating and Validating Fastly service versions.

This tool is an abstraction layer built on top of "[go-fastly](https://github.com/sethvargo/go-fastly)".

## Install

```bash
go get github.com/integralist/go-fastly-cli
```

> GOOS=darwin GOARCH=386 go build fastly.go

## Usage

```bash
fastly <flags> [diff <options>]
fastly <flags> [upload <options>]
fastly <flags> [list <options>]
fastly <flags> [delete <options>]
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
  -help, -h
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
  -validate string
        specify Fastly service version to validate
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

List Options:

```bash
fastly list -help

Usage of list:
  -version string
        specify Fastly service version to list VCL files from
```

Delete Options:

```bash
fastly delete -help

Usage of delete:
  -name string
        specify VCL filename to delete
  -version string
        specify Fastly service version to delete VCL files from
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

You can quickly view the relevant environment variables in your current shell using the following bash command:

```bash
env | sort | grep -iE '(vcl|fastly)'

FASTLY_API_TOKEN=123
FASTLY_SERVICE_ID=456
VCL_DIRECTORY=/Users/integralist/code/organization/cdn
VCL_MATCH_PATH=stage|www
VCL_SKIP_PATH=utils
```

In the above output we can see that the files I'll upload will be only those that are located in either a `stage` or `www` sub directory of my repository (which is found at `VCL_DIRECTORY`). I'll also skip uploading any files stored in the `utils` sub directory.

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

# validate specified service version
fastly -validate 123

# activate specified service version
fastly -activate 123

# view latest version of remote service vcl files
fastly list

# view version 123 of remote service vcl files
fastly list -version 123

# delete specified vcl file from latest version of remote service
fastly delete -name test_file

# delete specified vcl file from specific version of remote service
fastly delete -name test_file -version 123

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

# modify VCL directory temporarily + use different token/service id
VCL_MATCH_PATH=foo fastly -token $FASTLY_API_TOKEN_FOO -service $FASTLY_SERVICE_ID_FOO diff

# clone specified service version and upload local files to it
fastly upload -clone 123

# upload local files to the latest remote service version
fastly upload -latest

# clone latest service version available and upload local files to it
fastly upload
```

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

## TODO

* Ability to 'dry run' a command (to see what files are affected, e.g. what files will be uploaded and where)
* Ability to diff two remote services (not just local against a remote)
* Ability to upload individual files (not just pattern matched list of files)
* Ability to display all available services (along with their ID)
* Better diffing tool than linux `diff` command
* Setup for homebrew install
* Test Suite
