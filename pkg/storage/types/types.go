package types

type Storage interface {
	Put() error
}

type StorageOpts struct {
	Endpoint  string
	AccessKey string
	SecretKey string
}
