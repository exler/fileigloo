package cmd

import (
	"errors"
	"os"

	"github.com/exler/fileigloo/storage"
	"github.com/urfave/cli/v2"
)

var Cmd = &cli.App{
	Name:     "fileigloo",
	Usage:    "Small and simple online file sharing & pastebin",
	Commands: []*cli.Command{versionCmd, serverCmd, filesCmd},
}

func GetStorage(cCtx *cli.Context) (chosenStorage storage.Storage, err error) {
	switch storageProvider := cCtx.String("storage"); storageProvider {
	case "local":
		udir := cCtx.String("upload-directory")
		if udir == "" {
			return nil, errors.New("no upload directory specified")
		}

		chosenStorage, err = storage.NewLocalStorage(udir)
	case "s3":
		bucket := cCtx.String("s3-bucket")
		region := cCtx.String("s3-region")
		accessKey := cCtx.String("aws-access-key")
		secretKey := cCtx.String("aws-secret-key")
		sessionToken := cCtx.String("aws-session-token")
		endpointUrl := cCtx.String("aws-endpoint-url")

		chosenStorage, err = storage.NewS3Storage(accessKey, secretKey, sessionToken, endpointUrl, region, bucket)
	default:
		return nil, errors.New("wrong storage provider")
	}

	return
}

func Run() error {
	return Cmd.Run(os.Args)
}
