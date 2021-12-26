<p align="center">
    <img src="public/logo.svg" width="128">
    <p align="center">❄️ Small and simple temporary file sharing & pastebin </p>
    <p align="center">
      <img alt="GitHub Test Workflow Status" src="https://img.shields.io/github/workflow/status/exler/fileigloo/Test">
      <img alt="MIT License" src="https://img.shields.io/github/license/exler/fileigloo?color=lightblue">
    </p>
</p>



## Requirements

* Go >= 1.16

## Usage

Common configuration is in the [config/fileigloo.yaml](config/fileigloo.yaml) file. All of the configuration there, as well as flags passed to the program can be overriden by setting an environment variable (in uppercase). For example:

```bash
# Overrides the storage provider
$ export STORAGE=s3
```

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

## License

Copyright (c) 2021 by ***Kamil Marut***

`Fileigloo` is under the terms of the [MIT License](https://www.tldrlegal.com/l/mit), following all clarifications stated in the [license file](LICENSE).
