package database

import (
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

type MediaService struct {
	db *gorm.DB
}

type Option func(*MediaService)

func WithPoolMaxIdleConns(maxIdleConns int) Option {
	return func(s *MediaService) {
		sqlDB, _ := s.db.DB()
		sqlDB.SetMaxIdleConns(maxIdleConns)
	}
}

func WithPoolMaxOpenConns(maxOpenConns int) Option {
	return func(s *MediaService) {
		sqlDB, _ := s.db.DB()
		sqlDB.SetMaxOpenConns(maxOpenConns)
	}
}

func WithPoolMaxLifetime(maxLifetime time.Duration) Option {
	return func(s *MediaService) {
		sqlDB, _ := s.db.DB()
		sqlDB.SetConnMaxLifetime(maxLifetime)
	}
}

func NewMediaService(dialect gorm.Dialector, opts ...Option) *MediaService {

	db, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	svc := &MediaService{db: db}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func (s *MediaService) AutoMigrate() error {
	return s.db.AutoMigrate(&Media{})
}

func (s *MediaService) Ping() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func (s *MediaService) GetMedia(id uint) *Media {
	var media Media
	result := s.db.First(&media, id)
	if result.Error != nil {
		log.Println("Error while retrieving media with id", id, ":", result.Error)
		return nil
	}
	return &media
}

func (s *MediaService) CreateMedia(media *Media) error {
	tx := s.db.Begin()
	tx.Create(media)
	tx.Commit()
	return nil
}

func (s *MediaService) DeleteMedias(id ...uint) []*Media {
	if len(id) == 0 {
		return nil
	}
	wg := sync.WaitGroup{}
	tx := s.db.Begin()
	ec := make(chan error, len(id))
	ms := make([]*Media, 0, len(id))
	for _, mediaID := range id {
		wg.Go(func() {
			var media Media
			result := tx.First(&media, mediaID)
			if result.Error != nil {
				log.Println("Error while retrieving media with id", mediaID, ":", result.Error)
				ec <- result.Error
				return
			}
			result = tx.Delete(&media)
			if result.Error != nil {
				log.Println("Error while deleting media with id", mediaID, ":", result.Error)
				ec <- result.Error
				return
			}
			ms = append(ms, &media)
		})
	}
	wg.Wait()
	close(ec)
	if len(ec) > 0 {
		tx.Rollback()
		return nil
	}
	tx.Commit()
	return ms
}

func (s *MediaService) CheckMedia(id uint) error {
	result := s.db.Model(&Media{}).Where("id = ?", id).Update("check", true)
	return result.Error
}
