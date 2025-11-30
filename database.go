package gomedia

import (
	"sync"
	"time"

	"github.com/phuslu/log"

	"gorm.io/gorm"
)

type MediaEntity struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	Filename  string `gorm:"index"`
	Size      int64
	BasePath  *string
	Check     bool
}

type mediaService struct {
	db *gorm.DB
	l  *log.Logger
}

type dbOption func(*mediaService)

func WithPoolMaxIdleConns(maxIdleConns int) dbOption {
	return func(s *mediaService) {
		sqlDB, _ := s.db.DB()
		sqlDB.SetMaxIdleConns(maxIdleConns)
	}
}

func WithPoolMaxOpenConns(maxOpenConns int) dbOption {
	return func(s *mediaService) {
		sqlDB, _ := s.db.DB()
		sqlDB.SetMaxOpenConns(maxOpenConns)
	}
}

func WithPoolMaxLifetime(maxLifetime time.Duration) dbOption {
	return func(s *mediaService) {
		sqlDB, _ := s.db.DB()
		sqlDB.SetConnMaxLifetime(maxLifetime)
	}
}

func newMediaService(dialect gorm.Dialector, opts ...dbOption) *mediaService {

	db, err := gorm.Open(dialect, &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
	svc := &mediaService{db: db, l: getLogger()}
	for _, opt := range opts {
		opt(svc)
	}
	return svc
}

func (s *mediaService) AutoMigrate() error {
	return s.db.AutoMigrate(&MediaEntity{})
}

func (s *mediaService) Ping() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}

func (s *mediaService) GetMedia(id uint) *MediaEntity {
	media := &MediaEntity{}
	result := s.db.Model(&MediaEntity{}).Where(&MediaEntity{ID: id, Check: true}).First(media)
	if result.Error != nil {
		s.l.Error().Err(result.Error).Msgf("Failed to get media with id %d", id)
		return nil
	}
	return media
}

func (s *mediaService) CreateMedia(media *MediaEntity) error {
	tx := s.db.Begin()
	tx.Create(media)
	tx.Commit()
	return nil
}

func (s *mediaService) DeleteMedias(id ...uint) []*MediaEntity {
	if len(id) == 0 {
		return nil
	}
	wg := sync.WaitGroup{}
	tx := s.db.Begin()
	ec := make(chan error, len(id))
	ms := make([]*MediaEntity, 0, len(id))
	for _, mediaID := range id {
		wg.Go(func() {
			var media MediaEntity
			result := tx.First(&media, mediaID)
			if result.Error != nil {
				s.l.Error().Err(result.Error).Msgf("Error while retrieving media with id: %d", mediaID)
				ec <- result.Error
				return
			}
			result = tx.Delete(&media)
			if result.Error != nil {
				s.l.Error().Err(result.Error).Msgf("Error while deleting media with id %d", mediaID)
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

func (s *mediaService) CheckMedia(id uint) error {
	result := s.db.Model(&MediaEntity{}).Where("id = ?", id).Update("check", true)
	return result.Error
}
