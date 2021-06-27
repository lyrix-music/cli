package logging

import (
	"github.com/withmandala/go-log"
	"os"
)

var logger = log.New(os.Stdout)

func GetLogger() *log.Logger {
	return logger
}