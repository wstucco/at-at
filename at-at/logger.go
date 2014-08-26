package at_at

import (
	"log"
	"os"
)

type ILogger interface {
	Print(...interface{})
	Printf(string, ...interface{})
}

var logger ILogger

func init() {
	logger = log.New(os.Stderr, "", log.LstdFlags)
}

func Logger() ILogger {
	return logger
}
