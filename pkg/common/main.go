package common

import (
	"os"

	log "github.com/sirupsen/logrus"
)

func SetLogging() *log.Entry {
	cLogger := log.New()
	cLogger.SetFormatter(&log.JSONFormatter{})
	cLogger.SetOutput(os.Stdout)
	cLogger.SetLevel(log.InfoLevel)
	cLoggerEntry := cLogger.WithFields(log.Fields{
		"app": "gcp-status-exporter",
	})

	return cLoggerEntry

}
