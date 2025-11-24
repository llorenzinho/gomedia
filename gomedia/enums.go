package gomedia

type mediaProvider int

const (
	MediaProviderS3 mediaProvider = iota
	MediaProviderMinio
)
