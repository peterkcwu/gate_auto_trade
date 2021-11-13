package main

import (
	rotatelogs "github.com/lestrrat/go-file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

func InitLogger(path string, reserveDay int) error {
	if !filepath.IsAbs(path) {
		path = filepath.Join(filepath.Dir(os.Args[0]), path)
	}
	writer, err := rotatelogs.New(
		path+".%Y%m%d",
		rotatelogs.WithLinkName(path),
		rotatelogs.WithMaxAge(time.Duration(reserveDay)*24*time.Hour),
		rotatelogs.WithRotationTime(24*time.Hour),
	)

	if err != nil {
		return err
	}
	logrus.AddHook(lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
	}, &logrus.TextFormatter{}))

	logrus.SetOutput(ioutil.Discard)
	logrus.SetLevel(logrus.DebugLevel)

	return nil
}
