package configs

import (
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

// logrus loader into stdout
var (
	loadLoggerOnce sync.Once
	logger         *logrus.Logger
)

func LoadLogger() *logrus.Logger {
	loadLoggerOnce.Do(func() {
		logger = logrus.New()
		logger.SetOutput(os.Stdout)
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})

	})
	return logger
}

func GetLogger() *logrus.Logger {
	if logger == nil {
		LoadLogger()
	}
	return logger
}
