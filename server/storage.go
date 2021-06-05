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

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
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

type LocalStorage struct {
	Storage
	scheduler     *gocron.Scheduler
	basedir       string
	purgeInterval int
	purgeOlder    time.Duration
}

func NewLocalStorage(basedir string, purgeInterval, purgeOlder int) (*LocalStorage, error) {
	if basedir[len(basedir)-1:] != "/" {
		basedir += "/"
	}

	storage := &LocalStorage{
		basedir:       basedir,
		scheduler:     gocron.NewScheduler(time.UTC),
		purgeInterval: purgeInterval,
		purgeOlder:    time.Duration(purgeOlder) * time.Hour,
	}

	if purgeInterval != 0 {
		storage.scheduler.Every(storage.purgeInterval).Hours().Do(storage.Purge, storage.purgeOlder)
	}

	storage.scheduler.StartAsync()

	return storage, nil
}

func (s *LocalStorage) Get(filename string) (reader io.ReadCloser, err error) {
	path := filepath.Join(s.basedir, filename)
	reader, err = os.Open(path)
	return
}

func (s *LocalStorage) GetWithMetadata(filename string) (reader io.ReadCloser, metadata Metadata, err error) {
	reader, err = s.Get(filename)
	if err != nil {
		return
	}

	var mReader *io.File
	mPath := fmt.Sprintf("%s.metadata", filename)
	mReader, err = s.Get(mPath)
	if err != nil {
		return
	}

	err = json.NewDecoder(mReader).Decode(&metadata)
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

	_, err = io.Copy(mf, metadataBuffer)
	return err
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
	log.Println("Local storage: purging old files...")

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

func (s *LocalStorage) FileNotExists(err error) bool {
	if err == nil {
		return false
	}

	return os.IsNotExist(err)
}

type S3Storage struct {
	Storage
	bucket     string
	s3         *s3.S3
	session    *session.Session
	purgeOlder time.Duration
}

func newAWSSession(accessKey, secretKey, sessionToken, region string) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, sessionToken),
	}))
}

func NewS3Storage(accessKey, secretKey, sessionToken, region, bucket string) (*S3Storage, error) {
	sess := newAWSSession(accessKey, secretKey, sessionToken, region)

	return &S3Storage{
		bucket:  bucket,
		s3:      s3.New(sess),
		session: sess,
	}, nil
}

func (s *S3Storage) Get(filename string) (reader io.ReadCloser, err error) {
	r := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}

	response, err := s.s3.GetObject(r)
	if err != nil {
		return
	}
	reader = response.Body
	return
}

func (s *S3Storage) GetWithMetadata(filename string) (reader io.ReadCloser, metadata Metadata, err error) {
	reader, err = s.Get(filename)
	if err != nil {
		return
	}

	var mReader *io.File
	mPath := fmt.Sprintf("%s.metadata", filename)
	mReader, err = s.Get(mPath)
	if err != nil {
		return
	}

	err = json.NewDecoder(mReader).Decode(&metadata)
	return
}

func (s *S3Storage) Put(filename string, reader io.Reader, metadata Metadata) error {
	uploader := s3manager.NewUploader(s.session, func(u *s3manager.Uploader) {
		u.LeavePartsOnError = false
	})

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:  aws.String(s.bucket),
		Key:     aws.String(filename),
		Body:    reader,
		Expires: aws.Time(time.Now().Add(s.purgeOlder)),
	})
	if err != nil {
		return err
	}

	mPath := fmt.Sprintf("%s.metadata", filename)
	mBuffer := &bytes.Buffer{}
	if err = json.NewEncoder(mBuffer).Encode(metadata); err != nil {
		return err
	}

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:  aws.String(s.bucket),
		Key:     aws.String(mPath),
		Body:    mBuffer,
		Expires: aws.Time(time.Now().Add(s.purgeOlder)),
	})

	return err
}

func (s *S3Storage) Delete(filename string) error {
	r := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}
	_, err := s.s3.DeleteObject(r)
	if err != nil {
		return err
	}

	mPath := fmt.Sprintf("%s.metadata", filename)
	r = &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(mPath),
	}
	_, err = s.s3.DeleteObject(r)
	if err != nil {
		return err
	}

	return nil
}

func (s *S3Storage) Purge(days time.Duration) error {
	return nil
}

func (s *S3Storage) FileNotExists(err error) bool {
	if err == nil {
		return false
	}

	if awsError, ok := err.(awserr.Error); ok {
		switch awsError.Code() {
		case s3.ErrCodeNoSuchKey:
			return true
		}
	}

	return false
}
