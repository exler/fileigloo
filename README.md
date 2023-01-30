<p align="center">
    <img src="public/logo.svg" width="128">
    <p align="center">❄️ Small and simple online file sharing & pastebin</p>
    <p align="center">
      <img alt="GitHub Test Workflow Status" src="https://img.shields.io/github/actions/workflow/status/exler/fileigloo/test.yml?branch=main">
      <img alt="MIT License" src="https://img.shields.io/github/license/exler/fileigloo?color=lightblue">
    </p>
</p>

## Requirements

* Go >= 1.16

## Configuration

All configuration is done through environment variables.

### Local storage

```bash
# Override storage provider
$ export STORAGE=local

# Provide directory where uploaded files should be stored
$ export UPLOAD_DIRECTORY=uploads/
```

### S3

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

### Storj

```bash
# Override storage provider
$ export STORAGE=storj

# Specify Storj bucket and its access key
$ export STORJ_BUCKET=storage-bucket
$ export STORJ_ACCESS=
```

## Usage

### Program usage

```bash
Usage:
  fileigloo [command]

Available Commands:
  help        Help about any command
  runserver   Run web server
  version     Show current version

Flags:
  -h, --help   help for fileigloo

Use "fileigloo [command] --help" for more information about a command.
```

### cURL examples

* Upload file

```bash
$ curl -F file=@example.txt http://localhost:8000
http://localhost:8000/M7JeqHRk3uw0
```

* Upload paste

```bash
$ curl -F text="Example request" http://localhost:8000
http://localhost:8000/view/6QZuThTz8U7d
```

* Download file

```bash
# Write to stdout
$ curl http://localhost:8000/M7JeqHRk3uw0

# Write to file
$ curl -o output.txt http://localhost:8000/M7JeqHRk3uw0
```

* Delete file

```
$ curl -X DELETE <value from Delete-Url header>
```

## License

Copyright (c) 2021-2022 by ***Kamil Marut***

`Fileigloo` is under the terms of the [MIT License](https://www.tldrlegal.com/l/mit), following all clarifications stated in the [license file](LICENSE).
