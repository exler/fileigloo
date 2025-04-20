<p align="center">
    <img src="server/static/logo.svg" width="128">
    <p align="center">❄️ Small and simple online file sharing & pastebin</p>
    <p align="center">
      <img alt="GitHub Test Workflow Status" src="https://img.shields.io/github/actions/workflow/status/exler/fileigloo/test.yml?branch=main">
      <img alt="MIT License" src="https://img.shields.io/github/license/exler/fileigloo?color=lightblue">
    </p>
</p>

## Requirements

* Go >= 1.24

## Configuration

All configuration is done through environment variables or CLI flags.

### Local storage

```bash
# Override storage provider
$ export STORAGE=local

# Provide directory where uploaded files should be stored
$ export UPLOAD_DIRECTORY=uploads/
```

### S3-compatible storage

```bash
# Override storage provider
$ export STORAGE=s3

# Specify S3 bucket and region
$ export AWS_S3_BUCKET=storage-bucket
$ export AWS_S3_REGION=eu-central-1

# Specify AWS keys for accessing S3
$ export AWS_S3_ACCESS_KEY=
$ export AWS_S3_SECRET_KEY=

# Optionally, specify AWS session token for temporary credentials
$ export AWS_S3_SESSION_TOKEN=
```

## Usage

### Program usage

```bash
USAGE:
   fileigloo [global options] command [command options] [arguments...]

COMMANDS:
   version    Show current version
   runserver  Run web server
   files      Manage files in storage
   help, h    Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

## License

Copyright (c) 2021-2025 by ***Kamil Marut***

`Fileigloo` is under the terms of the [MIT License](https://www.tldrlegal.com/l/mit), following all clarifications stated in the [license file](LICENSE).
