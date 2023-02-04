package cmd

import (
	"errors"
	"fmt"

	colors "github.com/logrusorgru/aurora/v4"
	"github.com/urfave/cli/v2"
)

func truncateText(text string, length int) string {
	if len(text) > length {
		return text[:length] + "..."
	}

	return text
}

var (
	flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "storage",
			Value:   "local",
			EnvVars: []string{"STORAGE"},
		},
		&cli.StringFlag{
			Name:    "upload-directory",
			Value:   "uploads/",
			EnvVars: []string{"UPLOAD_DIRECTORY"},
		},
		&cli.StringFlag{
			Name:    "s3-bucket",
			EnvVars: []string{"S3_BUCKET"},
		},
		&cli.StringFlag{
			Name:    "s3-region",
			EnvVars: []string{"S3_REGION"},
		},
		&cli.StringFlag{
			Name:    "aws-access-key",
			EnvVars: []string{"AWS_ACCESS_KEY"},
		},
		&cli.StringFlag{
			Name:    "aws-secret-key",
			EnvVars: []string{"AWS_SECRET_KEY"},
		},
		&cli.StringFlag{
			Name:    "aws-session-token",
			EnvVars: []string{"AWS_SESSION_TOKEN"},
		},
		&cli.StringFlag{
			Name:    "aws-endpoint-url",
			EnvVars: []string{"AWS_ENDPOINT_URL"},
		},
	}

	filesCmd = &cli.Command{
		Name:  "files",
		Usage: "Manage files in storage",
		Subcommands: []*cli.Command{
			{
				Name:  "list",
				Usage: "List files in storage",
				Flags: flags,
				Action: func(cCtx *cli.Context) error {
					s, err := GetStorage(cCtx)
					if err != nil {
						return err
					}

					files, metadata, err := s.List(cCtx.Context)
					if err != nil {
						return err
					}

					fmt.Println(colors.Blue("File ID | Size (bytes) | Filename"))
					for i, file := range files {
						fmt.Println(file, metadata[i].ContentLength, truncateText(metadata[i].Filename, 32))
					}
					return nil
				},
			},
			{
				Name:  "delete",
				Usage: "Delete given file from storage",
				Flags: flags,
				Action: func(cCtx *cli.Context) error {
					s, err := GetStorage(cCtx)
					if err != nil {
						return err
					}

					fileID := cCtx.Args().First()
					if fileID == "" {
						return errors.New("no file id provided")
					}

					err = s.Delete(cCtx.Context, fileID)
					if err != nil {
						return err
					}
					fmt.Println(colors.Blue(fmt.Sprintf("File deleted [fileId=%s]", fileID)))
					return nil
				},
			},
		},
	}
)
