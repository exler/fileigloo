package storage

import (
	"context"
	"io"
)

type Metadata struct {
	Filename      string // Original filename
	ContentType   string
	ContentLength string
	PasswordHash  string // Argon2id hash of password (empty if no password)
}

func MetadataToStringMap(metadata Metadata) map[string]*string {
	m := make(map[string]*string)

	m["Filename"] = &metadata.Filename
	m["Content-Type"] = &metadata.ContentType
	m["Content-Length"] = &metadata.ContentLength
	m["Password-Hash"] = &metadata.PasswordHash

	return m
}

func StringMapToMetadata(m map[string]*string) Metadata {
	metadata := Metadata{
		Filename:      *m["Filename"],
		ContentType:   *m["Content-Type"],
		ContentLength: *m["Content-Length"],
		PasswordHash:  *m["Password-Hash"],
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
