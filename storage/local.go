package storage

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
