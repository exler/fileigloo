package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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
	Get(filename string) (reader io.ReadSeeker, metadata Metadata, err error)
	Put(filename string, reader io.Reader, metadata Metadata) error
	Delete(filename string) error
	Purge(days time.Duration) error
	FileExists(filename string) bool
}

type LocalStorage struct {
	Storage
	basedir string
}

func NewLocalStorage(basedir string) (*LocalStorage, error) {
	if basedir[len(basedir)-1:] != "/" {
		basedir += "/"
	}

	return &LocalStorage{basedir: basedir}, nil
}

func (s *LocalStorage) Get(filename string) (reader io.ReadSeeker, metadata Metadata, err error) {
	path := filepath.Join(s.basedir, filename)
	if reader, err = os.Open(path); err != nil {
		return
	}

	metadataPath := fmt.Sprintf("%s.metadata", path)
	if metadataReader, metadataError := os.Open(metadataPath); metadataError != nil {
		err = metadataError
		return
	} else if metadataError = json.NewDecoder(metadataReader).Decode(&metadata); err != nil {
		err = metadataError
		return
	}

	return
}

func (s *LocalStorage) Put(filename string, reader io.Reader, metadata Metadata) error {
	var f, mf io.WriteCloser
	var err error

	if err = os.MkdirAll(s.basedir, 0755); os.IsNotExist(err) {
		return err
	}

	path := filepath.Join(s.basedir, filename)
	if f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, reader); err != nil {
		return err
	}

	metadataPath := fmt.Sprintf("%s.metadata", path)
	metadataBuffer := &bytes.Buffer{}
	if err = json.NewEncoder(metadataBuffer).Encode(metadata); err != nil {
		return err
	}

	if mf, err = os.OpenFile(metadataPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600); err != nil {
		return err
	}
	defer mf.Close()

	if _, err = io.Copy(mf, metadataBuffer); err != nil {
		return err
	}

	return nil
}

func (s *LocalStorage) Delete(filename string) error {
	path := filepath.Join(s.basedir, filename)
	metadataPath := fmt.Sprintf("%s.metadata", path)
	if err := os.Remove(path); err != nil {
		return err
	} else if err := os.Remove(metadataPath); err != nil {
		return err
	}

	return nil
}

func (s *LocalStorage) Purge(days time.Duration) error {
	log.Println("Purging old files")

	err := filepath.Walk(s.basedir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.ModTime().Before(time.Now().Add(-1 * days)) {
			err = os.Remove(path)
			return err
		}

		return nil
	})

	return err
}

func (s *LocalStorage) FileExists(filename string) bool {
	path := filepath.Join(s.basedir, filename)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	return true
}
