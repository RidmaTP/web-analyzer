package configs

import (
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var (
	loadEnvOnce sync.Once
	version     string
	port        string
)

// load env in config pkg idempotently with sync.once
func LoadEnv() error {
	var loadErr error
	loadEnvOnce.Do(func() {
		err := godotenv.Load("configs/.env")
		if err != nil {
			loadErr = err
		}

		version = os.Getenv("APP_VERSION")
		port = os.Getenv("PORT")
	})
	if loadErr != nil {
		return loadErr
	}
	return nil
}

func GetAppVersion() string {
	if version == ""{
		LoadEnv()
	}
	return version
}

func GetPort() string {
	if port == "" {
		LoadEnv()
	}
	return port
}
