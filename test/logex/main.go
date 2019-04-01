package main

import (
	log "github.com/sirupsen/logrus"
	// Register the default logger with accpeted default settings
	// Best done in main package at bootstrap.
	_ "github.com/govinda-attal/kiss-lib/pkg/logrus/reglog"

	"github.com/govinda-attal/kiss-lib/test/logex/valuable"
)

func main() {
	// Based on configuration, you could override default settings
	// For example the level
	log.SetLevel(log.DebugLevel)

	log.WithFields(
		log.Fields{
			"go": "less code is more readable!",
		},
	).Infoln("this is info message")

	log.Debugln("this is debug message")

	// this logs something valuable.
	// if you notice there less instrumentation and we didn't pass custom logger instance!
	// less is more!
	valuable.Operation()
}
