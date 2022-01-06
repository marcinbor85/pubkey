package log

import (
	"log"
	"os"
)

var (
	debugLogger   *log.Logger
	infoLogger    *log.Logger
	warningLogger *log.Logger
	errorLogger   *log.Logger
)

type LogLevel int

const (
	Off LogLevel = -1
	Error   	 = 0
	Warning 	 = 10
	Info    	 = 20
	Debug   	 = 30
)

var currentLogLevel LogLevel = Info

func Init(level LogLevel) {
	file := os.Stdout
	flags := log.Ldate | log.Ltime | log.Lmsgprefix
	debugLogger = log.New(file, "DEBUG: ", flags)
	infoLogger = log.New(file, "INFO: ", flags)
	warningLogger = log.New(file, "WARNING: ", flags)
	errorLogger = log.New(file, "ERROR: ", flags)

	currentLogLevel = level
}

func D(fmt string, args ...interface{}) {
	if currentLogLevel < Debug {
		return
	}
	debugLogger.Printf(fmt, args...)
}

func I(fmt string, args ...interface{}) {
	if currentLogLevel < Info {
		return
	}
	infoLogger.Printf(fmt, args...)
}

func W(fmt string, args ...interface{}) {
	if currentLogLevel < Warning {
		return
	}
	warningLogger.Printf(fmt, args...)
}

func E(fmt string, args ...interface{}) {
	if currentLogLevel < Error {
		return
	}
	errorLogger.Printf(fmt, args...)
}
