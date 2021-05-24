package server

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

type Storage interface {
	Get(filename string) (reader io.ReadSeekCloser, contentLength int64, err error)
	Put(filename string, reader io.Reader) error
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

func (s *LocalStorage) Get(filename string) (reader io.ReadSeekCloser, contentLength int64, err error) {
	path := filepath.Join(s.basedir, filename)

	if reader, err = os.Open(path); err != nil {
		return
	}

	var fi os.FileInfo
	if fi, err = os.Lstat(path); err != nil {
		return
	}

	contentLength = fi.Size()
	return
}

func (s *LocalStorage) Put(filename string, reader io.Reader) error {
	var f io.WriteCloser
	var err error

	if err = os.MkdirAll(s.basedir, os.ModeDir); os.IsNotExist(err) {
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

	return nil
}

func (s *LocalStorage) Delete(filename string) error {
	path := filepath.Join(s.basedir, filename)
	err := os.Remove(path)
	return err
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
