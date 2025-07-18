package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/exler/fileigloo/datetime"
)

type LocalStorage struct {
	Storage
	basedir string
}

func NewLocalStorage(basedir string) (*LocalStorage, error) {
	if basedir[len(basedir)-1:] != "/" {
		basedir += "/"
	}

	storage := &LocalStorage{
		basedir: basedir,
	}

	return storage, nil
}

func (s *LocalStorage) Type() string {
	return "local"
}

func (s *LocalStorage) List(ctx context.Context) (filenames []string, metadata []Metadata, err error) {
	err = filepath.WalkDir(s.basedir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".metadata" {
			return nil
		}

		filenames = append(filenames, d.Name())

		m, err := s.GetOnlyMetadata(ctx, d.Name())
		metadata = append(metadata, m)

		return err
	})
	return
}

func (s *LocalStorage) Get(ctx context.Context, filename string) (reader io.ReadCloser, err error) {
	path := filepath.Join(s.basedir, filename)
	reader, err = os.Open(path) //#nosec
	return
}

func (s *LocalStorage) GetWithMetadata(ctx context.Context, filename string) (reader io.ReadCloser, metadata Metadata, err error) {
	reader, err = s.Get(ctx, filename)
	if err != nil {
		return
	}

	var mReader io.ReadCloser
	mPath := fmt.Sprintf("%s.metadata", filename)
	mReader, err = s.Get(ctx, mPath)
	if err != nil {
		return
	}
	defer mReader.Close()

	err = json.NewDecoder(mReader).Decode(&metadata)
	return
}

func (s *LocalStorage) GetOnlyMetadata(ctx context.Context, filename string) (metadata Metadata, err error) {
	var mReader io.ReadCloser
	mPath := fmt.Sprintf("%s.metadata", filename)
	mReader, err = s.Get(ctx, mPath)
	if err != nil {
		return
	}
	defer mReader.Close()

	err = json.NewDecoder(mReader).Decode(&metadata)
	return
}

func (s *LocalStorage) Put(ctx context.Context, filename string, reader io.Reader, metadata Metadata) error {
	var f, mf io.WriteCloser
	var err error

	if err = os.MkdirAll(s.basedir, 0600); os.IsNotExist(err) {
		return err
	}

	path := filepath.Join(s.basedir, filename)
	f, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) // #nosec G304
	if err != nil {
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

	mf, err = os.OpenFile(metadataPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600) // #nosec G304
	if err != nil {
		return err
	}
	defer mf.Close()

	_, err = io.Copy(mf, metadataBuffer)
	return err
}

func (s *LocalStorage) Delete(ctx context.Context, filename string) error {
	path := filepath.Join(s.basedir, filename)
	metadataPath := fmt.Sprintf("%s.metadata", path)
	if err := os.Remove(path); err != nil {
		return err
	} else if err := os.Remove(metadataPath); err != nil {
		return err
	}

	return nil
}

func (s *LocalStorage) DeleteExpired(ctx context.Context) (deletedCount int, err error) {
	filenames, metadata, err := s.List(ctx)
	if err != nil {
		return 0, err
	}

	for i, filename := range filenames {
		if datetime.IsExpired(metadata[i].ExpiresAt) {
			if err := s.Delete(ctx, filename); err != nil {
				// Log error but continue with other files
				continue
			}
			deletedCount++
		}
	}

	return deletedCount, nil
}

func (s *LocalStorage) FileNotExists(err error) bool {
	if err == nil {
		return false
	}

	return os.IsNotExist(err)
}
