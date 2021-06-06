package storage

import (
	"io"
	"strconv"
	"time"
)

type Metadata struct {
	Filename      string // original filename
	ContentType   string
	ContentLength string
}

func MakeMetadata(filename, contentType string, contentLength int64) Metadata {
	metadata := Metadata{
		Filename:      filename,
		ContentType:   contentType,
		ContentLength: strconv.FormatInt(contentLength, 10),
	}

	return metadata
}

func MetadataToStringMap(metadata Metadata) map[string]*string {
	m := make(map[string]*string)

	m["Filename"] = &metadata.Filename
	m["Content-Type"] = &metadata.ContentType
	m["Content-Length"] = &metadata.ContentLength

	return m
}

func StringMapToMetadata(m map[string]*string) Metadata {
	metadata := Metadata{
		Filename:      *m["Filename"],
		ContentType:   *m["Content-Type"],
		ContentLength: *m["Content-Length"],
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
	Type() string
}
