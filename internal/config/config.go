package config

import "os"

const defaultListenAddress = "127.0.0.1:8780"

type Config struct {
	ListenAddress string
	StorageRoot   string
}

func LoadFromEnv() Config {
	listenAddress := os.Getenv("GC_PHOTOS_LISTEN")
	if listenAddress == "" {
		listenAddress = defaultListenAddress
	}

	return Config{
		ListenAddress: listenAddress,
		StorageRoot:   os.Getenv("GC_PHOTOS_STORAGE_ROOT"),
	}
}
