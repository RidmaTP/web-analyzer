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
	return version
}

func GetPort() string {
	return port
}
