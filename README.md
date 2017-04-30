# go-fastly-cli

> CLI tool for uploading and diffing local/remote Fastly VCL

## Install

```bash
go get github.com/integralist/go-fastly-cli
```

## Usage

```bash
fastly <flags> diff <options>
fastly <flags> upload <options>
```

> flags: settings common to both commands  
> options: unique to the specific command

Flags:

```bash
fastly -help

  -debug
        show any error/diff output + debug logs
  -dir string
        vcl directory to compare files against 
  -help
        show available flags
  -match string
        regex for matching vcl directories (will also try: VCL_MATCH_DIRECTORY)
  -service string
        your service id (fallback: FASTLY_SERVICE_ID) 
  -skip string
        regex for skipping vcl directories (will also try: VCL_SKIP_DIRECTORY) 
  -token string
        your fastly api token (fallback: FASTLY_API_TOKEN) 
  -version
        show application version
```

Diff Options:

```bash
fastly diff -help

Usage of diff:
  -vcl-version string
        specify Fastly service 'version' to verify against
```

Upload Options:

```bash
fastly upload -help

Usage of upload:
  -activate-version string
        specify Fastly service 'version' to activate
  -clone-version string
        specify Fastly service 'version' to clone from before uploading to
  -get-latest-version
        get latest Fastly service version and its active status
  -get-settings string
        get settings (Default TTL & Host) for specified Fastly service version (version number or latest)
  -get-version-status string
        retrieve status for the specified Fastly service 'version'
  -upload-version string
        specify non-active Fastly service 'version' to upload to
  -use-latest-version
        use latest Fastly service version to upload to (presumes not activated)
```

## Environment Variables

The use of environment variables help to reduce the amount of flags required by the `fastly` CLI tool.

For example, I always diff against a 'stage' service in our Fastly account and so I don't want to have to put in the same credentials all the time.

Below is a list of environment variables this tool supports:

* `FASTLY_API_TOKEN` (`-token`)
* `FASTLY_SERVICE_ID` (`-service`)
* `VCL_DIRECTORY` (`-dir`)
* `VCL_MATCH_DIRECTORY` (`-match`)
* `VCL_SKIP_DIRECTORY` (`-skip`)

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
