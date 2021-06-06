<p align="center">
    <img src="public/logo.svg" width="124">
    <p align="center">❄️ Small and simple temporary file sharing & pastebin </p>
    <p align="center">
      <img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/exler/fileigloo">
    </p>
</p>



## Requirements

* Go >= 1.15

## Usage

Common configuration is in the [config/fileigloo.yaml](config/fileigloo.yaml) file. All of the configuration there, as well as flags passed to the program can be overriden by setting an environment variable (in uppercase). For example:

```bash
# Overrides the storage provider
$ export STORAGE=s3
```

Program usage:

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

## License

Copyright (c) 2021 by ***Kamil Marut***

`Fileigloo` is under the terms of the [MIT License](https://www.tldrlegal.com/l/mit), following all clarifications stated in the [license file](LICENSE).
