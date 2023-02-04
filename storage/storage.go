package storage

import (
	"context"
	"io"
	"strconv"
)

type Metadata struct {
	Filename      string // Original filename
	ContentType   string
	ContentLength string
	DeleteToken   string
}

func MakeMetadata(filename, contentType string, contentLength int64, deleteToken string) Metadata {
	metadata := Metadata{
		Filename:      filename,
		ContentType:   contentType,
		ContentLength: strconv.FormatInt(contentLength, 10),
		DeleteToken:   deleteToken,
	}

	return metadata
}

func MetadataToStringMap(metadata Metadata) map[string]*string {
	m := make(map[string]*string)

	m["Filename"] = &metadata.Filename
	m["Content-Type"] = &metadata.ContentType
	m["Content-Length"] = &metadata.ContentLength
	m["Delete-Token"] = &metadata.DeleteToken

	return m
}

func StringMapToMetadata(m map[string]*string) Metadata {
	metadata := Metadata{
		Filename:      *m["Filename"],
		ContentType:   *m["Content-Type"],
		ContentLength: *m["Content-Length"],
		DeleteToken:   *m["Delete-Token"],
	}

	return metadata
}

type Storage interface {
	List(ctx context.Context) (filenames []string, metadata []Metadata, err error)
	Get(ctx context.Context, filename string) (reader io.ReadCloser, err error)
	GetWithMetadata(ctx context.Context, filename string) (reader io.ReadCloser, metadata Metadata, err error)
	GetOnlyMetadata(ctx context.Context, filename string) (metadata Metadata, err error)
	Put(ctx context.Context, filename string, reader io.Reader, metadata Metadata) error
	Delete(ctx context.Context, filename string) error
	FileNotExists(err error) bool
	Type() string
}
