package storage

import (
	"context"
	"io"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type S3Storage struct {
	Storage
	s3      *s3.S3
	session *session.Session
	bucket  string
}

func newAWSSession(accessKey, secretKey, sessionToken, endpointUrl, region string) *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		Endpoint:    aws.String(endpointUrl),
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKey, secretKey, sessionToken),
	}))
}

func (s *S3Storage) Type() string {
	return "s3"
}

func NewS3Storage(accessKey, secretKey, sessionToken, endpointUrl, region, bucket string) (*S3Storage, error) {
	session := newAWSSession(accessKey, secretKey, sessionToken, endpointUrl, region)

	return &S3Storage{
		s3:      s3.New(session),
		session: session,
		bucket:  bucket,
	}, nil
}

func (s *S3Storage) List(ctx context.Context) (filenames []string, metadata []Metadata, err error) {
	r := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
	}

	response, err := s.s3.ListObjectsV2(r)
	if err != nil {
		return
	}

	contents := response.Contents
	for _, obj := range contents {
		filenames = append(filenames, *obj.Key)
		metadata = append(metadata, Metadata{ContentLength: strconv.Itoa(int(*obj.Size)), Filename: "N/A"})
	}
	return
}

func (s *S3Storage) Get(ctx context.Context, filename string) (reader io.ReadCloser, err error) {
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

func (s *S3Storage) GetWithMetadata(ctx context.Context, filename string) (reader io.ReadCloser, metadata Metadata, err error) {
	r := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}

	response, err := s.s3.GetObject(r)
	if err != nil {
		return
	}
	reader = response.Body
	metadata = StringMapToMetadata(response.Metadata)
	return
}

func (s *S3Storage) GetOnlyMetadata(ctx context.Context, filename string) (metadata Metadata, err error) {
	r := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}

	response, err := s.s3.HeadObject(r)
	if err != nil {
		return
	}
	metadata = StringMapToMetadata(response.Metadata)
	return
}

func (s *S3Storage) Put(ctx context.Context, filename string, reader io.Reader, metadata Metadata) error {
	uploader := s3manager.NewUploader(s.session, func(u *s3manager.Uploader) {
		u.LeavePartsOnError = false
	})
	mMap := MetadataToStringMap(metadata)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket:   aws.String(s.bucket),
		Key:      aws.String(filename),
		Body:     reader,
		Metadata: mMap,
	})

	return err
}

func (s *S3Storage) Delete(ctx context.Context, filename string) error {
	r := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(filename),
	}
	_, err := s.s3.DeleteObject(r)

	return err
}

func (s *S3Storage) DeleteExpired(ctx context.Context) (deletedCount int, err error) {
	filenames, metadata, err := s.List(ctx)
	if err != nil {
		return 0, err
	}

	for i, filename := range filenames {
		if IsMetadataExpired(metadata[i]) {
			if err := s.Delete(ctx, filename); err != nil {
				// Log error but continue with other files
				continue
			}
			deletedCount++
		}
	}

	return deletedCount, nil
}

func (s *S3Storage) FileNotExists(err error) bool {
	if err == nil {
		return false
	}

	if awsError, ok := err.(awserr.Error); ok {
		if awsError.Code() == s3.ErrCodeNoSuchKey {
			return true
		}
	}

	return false
}
