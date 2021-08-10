package cmd

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type typeOfLog int

const (
	Uncategorized typeOfLog = iota
	Info
	Warn
	Fatal
)

func logger(message string, logType typeOfLog) {
	customFormatter := new(log.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	log.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
	switch logType {
	case Uncategorized:
		fmt.Println(message)
	case Info:
		if Verbose {
			log.Info(message)
		}
	case Warn:
		log.Warn(message)
	case Fatal:
		log.Fatal(message)
	default:
		fmt.Println(message)
	}
}
