package gomedia

type ErrUnsupportedMediaProvider struct{}

func (e ErrUnsupportedMediaProvider) Error() string {
	return "unsupported media provider"
}
