package main

import (
	"log"
)

func log_Fatalf(format string, v ...interface{}) {
	log.Fatalf(format, v)
}
func log_Errorf(format string, v ...interface{}) {
	log.Printf(format, v)
}
func log_Infof(format string, v ...interface{}) {
	log.Printf(format, v)
}
func log_Debugf(format string, v ...interface{}) {
	if debugEnabled {
		log.Printf(format, v)
	}
}