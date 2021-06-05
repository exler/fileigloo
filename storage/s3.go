package storage

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
