package storage

import (
	"io"
	"time"
)

type Metadata struct {
	Filename      string // original filename
	ContentType   string
	ContentLength int64
}

func MakeMetadata(filename string, contentType string, contentLength int64) Metadata {
	metadata := Metadata{
		Filename:      filename,
		ContentType:   contentType,
		ContentLength: contentLength,
	}

	return metadata
}

type Storage interface {
	Get(filename string) (reader io.ReadCloser, err error)
	GetWithMetadata(filename string) (reader io.ReadCloser, metadata Metadata, err error)
	Put(filename string, reader io.Reader, metadata Metadata) error
	Delete(filename string) error
	Purge(days time.Duration) error
	FileNotExists(err error) bool
}
