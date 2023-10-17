package storage

import (
	"encoding/csv"
	"os"
	"shortener/config"
	"shortener/internal/logger"
)

type StorageT struct {
	cache      map[any]any
	file       *os.File
	fileWriter *csv.Writer
}

type StorageI interface {
	Load(key any) (value any, ok bool)
	Store(key, value any)
	Range(f func(key, value any) bool)
	Close() error
}

func New() *StorageT {
	f, err := os.OpenFile(*config.FileStoragePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		logger.WithField("error", err).Errorln("Error while open file")
	}
	c := make(map[any]any)
	r := csv.NewReader(f)
	r.FieldsPerRecord = 2

	rs, err := r.ReadAll()
	if err != nil {
		logger.WithField("error", err).Errorln("Error while decode csv")
	}
	for _, cols := range rs {
		c[cols[0]] = cols[1]
	}
	return &StorageT{
		cache:      c,
		file:       f,
		fileWriter: csv.NewWriter(f),
	}
}

func (s *StorageT) Load(key any) (value any, ok bool) {
	v, ok := s.cache[key]
	return v, ok
}

func (s *StorageT) Store(key, value any) {
	s.cache[key] = value
	s.fileWriter.Write([]string{key.(string), value.(string)})
}

func (s *StorageT) Range(f func(key, value any) bool) {
	for k, v := range s.cache {
		if !f(k, v) {
			break
		}
	}
}

func (s *StorageT) Close() error {
	s.fileWriter.Flush()
	if err := s.fileWriter.Error(); err != nil {
		logger.WithField("error", err).Errorln("Error while encode csv")
	}
	return s.file.Close()
}

//type FileStorageRecordT struct {
//	UUID        string `json:"uuid"`
//	ShortURL    string `json:"short_url"`
//	OriginalURL string `json:"original_url"`
//}
//
//type FileStorageT struct {
//	sync.Mutex
//	records []FileStorageRecordT
//}
//
//func (f *FileStorageT) ListRecords() ([]FileStorageRecordT, error) {
//	f.Lock()
//	defer f.Unlock()
//	file, err := os.OpenFile(*config.FileStoragePath, os.O_CREATE|os.O_RDONLY, 0644)
//	if err != nil {
//		return nil, err
//	}
//	defer file.Close()
//	return f.records, json.NewDecoder(file).Decode(&f.records)
//}
//
//func (f *FileStorageT) Add(ShortURL, OriginalURL string) error {
//	f.Lock()
//	defer f.Unlock()
//	f.records = append(f.records, FileStorageRecordT{
//		UUID:        uuid.New().String(),
//		ShortURL:    ShortURL,
//		OriginalURL: OriginalURL,
//	})
//	file, err := os.OpenFile(*config.FileStoragePath, os.O_CREATE|os.O_WRONLY, 0644)
//	if err != nil {
//		return err
//	}
//	defer file.Close()
//	return json.NewEncoder(file).Encode(f.records)
//}
//
//func (f *FileStorageT) Get(ShortURL string) (FileStorageRecordT, error) {
//	f.Lock()
//	defer f.Unlock()
//	file, err := os.OpenFile(*config.FileStoragePath, os.O_CREATE|os.O_RDONLY, 0644)
//	if err != nil {
//		return FileStorageRecordT{}, err
//	}
//	defer file.Close()
//	err = json.NewDecoder(file).Decode(&f.records)
//	if err != nil {
//		return FileStorageRecordT{}, err
//	}
//	for _, record := range f.records {
//		if ShortURL == record.ShortURL {
//			return record, nil
//		}
//	}
//	return FileStorageRecordT{}, errors.New("not found")
//}
