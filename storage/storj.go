package storage

import (
	"context"
	"errors"
	"io"
	"time"

	"storj.io/common/fpath"
	"storj.io/uplink"
)

type StorjStorage struct {
	Storage
	project    *uplink.Project
	bucket     *uplink.Bucket
	purgeOlder time.Duration
}

func (s *StorjStorage) Type() string {
	return "storj"
}

func NewStorjStorage(access, bucket string, purgeOlder int) (*StorjStorage, error) {
	var instance StorjStorage
	var err error
	tempCtx := context.TODO()

	ctx := fpath.WithTempData(tempCtx, "", true)
	if err != nil {
		return nil, err
	}

	parsedAccess, err := uplink.ParseAccess(access)
	if err != nil {
		return nil, err
	}

	instance.project, err = uplink.OpenProject(ctx, parsedAccess)
	if err != nil {
		return nil, err
	}

	instance.bucket, err = instance.project.EnsureBucket(ctx, bucket)
	if err != nil {
		instance.project.Close()
		return nil, err
	}

	instance.purgeOlder = time.Duration(purgeOlder) * time.Hour

	return &instance, nil

}

func (s *StorjStorage) Get(ctx context.Context, filename string) (reader io.ReadCloser, err error) {
	download, err := s.project.DownloadObject(fpath.WithTempData(ctx, "", true), s.bucket.Name, filename, nil)
	if err != nil {
		return nil, err
	}

	reader = download
	return
}

func (s *StorjStorage) GetWithMetadata(ctx context.Context, filename string) (reader io.ReadCloser, metadata Metadata, err error) {
	download, err := s.project.DownloadObject(fpath.WithTempData(ctx, "", true), s.bucket.Name, filename, nil)
	if err != nil {
		return nil, Metadata{}, err
	}

	reader = download
	metadata = MakeMetadata(download.Info().Custom["Filename"], download.Info().Custom["Content-Type"], download.Info().System.ContentLength)
	return
}

func (s *StorjStorage) Put(ctx context.Context, filename string, reader io.Reader, metadata Metadata) error {
	var uploadOptions *uplink.UploadOptions
	if s.purgeOlder.Hours() > 0 {
		uploadOptions = &uplink.UploadOptions{Expires: time.Now().Add(s.purgeOlder)}
	}

	writer, err := s.project.UploadObject(fpath.WithTempData(ctx, "", true), s.bucket.Name, filename, uploadOptions)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, reader)
	if err != nil {
		_ = writer.Abort()
		return err
	}

	err = writer.SetCustomMetadata(ctx, uplink.CustomMetadata{"Filename": metadata.Filename, "Content-Type": metadata.ContentType})
	if err != nil {
		_ = writer.Abort()
		return err
	}

	err = writer.Commit()
	return err
}

func (s *StorjStorage) Delete(ctx context.Context, filename string) error {
	_, err := s.project.DeleteObject(fpath.WithTempData(ctx, "", true), s.bucket.Name, filename)
	return err
}

func (s *StorjStorage) Purge(ctx context.Context, days time.Duration) error {
	return nil
}

func (s *StorjStorage) FileNotExists(err error) bool {
	return errors.Is(err, uplink.ErrObjectNotFound)
}
