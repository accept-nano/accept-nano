package main

import (
	"fmt"
	"path"

	"github.com/cenkalti/log"
)

type logFormatter struct{}

func (f logFormatter) Format(rec *log.Record) string {
	return fmt.Sprintf("%s [%s] %-8s %s:%d %s", fmt.Sprint(rec.Time)[:19], rec.LoggerName, rec.Level, path.Base(rec.Filename), rec.Line, rec.Message)
}

func init() {
	formatter := &logFormatter{}
	log.DefaultHandler.SetFormatter(formatter)
}
