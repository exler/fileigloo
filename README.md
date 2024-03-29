<p align="center">
    <img src="server/static/logo.svg" width="128">
    <p align="center">❄️ Small and simple online file sharing & pastebin</p>
    <p align="center">
      <img alt="GitHub Test Workflow Status" src="https://img.shields.io/github/actions/workflow/status/exler/fileigloo/test.yml?branch=main">
      <img alt="MIT License" src="https://img.shields.io/github/license/exler/fileigloo?color=lightblue">
    </p>
</p>

## Requirements

* Go >= 1.20

## Configuration

All configuration is done through environment variables or CLI flags.

### Local storage

```bash
# Override storage provider
$ export STORAGE=local

# Provide directory where uploaded files should be stored
$ export UPLOAD_DIRECTORY=uploads/
```

### Amazon S3

```bash
# Override storage provider
$ export STORAGE=s3

# Specify S3 bucket and region
$ export S3_BUCKET=storage-bucket
$ export S3_REGION=eu-central-1

# Specify AWS keys for accessing S3
$ export AWS_ACCESS_KEY=
$ export AWS_SECRET_KEY=

# Optionally, specify AWS session token for temporary credentials
$ export AWS_SESSION_TOKEN=
```

### Cloudflare R2

Cloudflare R2 implements S3 API, so we use the `s3` storage here as well.

```bash
# Override storage provider
$ export STORAGE=s3

# Specify R2 bucket
$ export S3_BUCKET=storage-bucket

# Region must be 'auto'
$ export S3_REGION=auto

# Specify R2 keys for accessing the bucket
$ export AWS_ACCESS_KEY=
$ export AWS_SECRET_KEY=

# Specify the bucket URL
$ export AWS_ENDPOINT_URL=https://${accountid}.r2.cloudflarestorage.com
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

Copyright (c) 2021-2023 by ***Kamil Marut***

`Fileigloo` is under the terms of the [MIT License](https://www.tldrlegal.com/l/mit), following all clarifications stated in the [license file](LICENSE).
