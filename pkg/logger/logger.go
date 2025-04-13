package logger

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Log *logrus.Logger

func InitLogger() {
	Log = logrus.New()

	// Output to stdout instead of the default stderr
	Log.Out = os.Stdout

	// Set JSON formatter for structured logging
	Log.SetFormatter(&logrus.JSONFormatter{})

	// Log level can be changed depending on environment
	Log.SetLevel(logrus.InfoLevel)
}
