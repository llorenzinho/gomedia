package gomedia

import (
	"io"

	"github.com/llorenzinho/gomedia/database"
)

type mediaSaver interface {
	SaveMedia(r *io.Reader, meta MediaMeta) (*database.Media, error) // saves media and return its database id
}

type mediaDeleter interface {
	DeleteMedia(id uint) error // deletes media by its database id
}

type mediaURLGetter interface {
	GetMediaURL(id uint) (*string, error) // returns a presigned url to access the media
}

type mediaReaderGetter interface {
	GetMediaReader(id uint) (*Media, error) // returns a reader to download the media
}

type MediaStorer interface {
	HealthCheck() error
	mediaSaver
	mediaDeleter
	mediaURLGetter
	mediaReaderGetter
}
