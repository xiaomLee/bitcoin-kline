package logger

import (
	"path"
	"time"

	"github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

var DefaultLoggerFormat = &logrus.TextFormatter{
	TimestampFormat: "2006-01-02 15:04:05",
	FieldMap: logrus.FieldMap{
		logrus.FieldKeyMsg: "message",
	},
}

var (
	_appName = ""
)

//配置业务日志系统
func ConfigLogger(curMode string, appName, dir string, logName string, keepDays int, rotateHours int) {
	if curMode != "online" {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	_appName = appName
	logrus.SetFormatter(DefaultLoggerFormat)
	configLocalFs(dir, logName, time.Hour*time.Duration(keepDays*24), time.Hour*time.Duration(rotateHours))
}

//配置本地文件系统并按周期分割
func configLocalFs(logPath string, logFileName string, maxAge time.Duration, rotationTime time.Duration) {
	baseLogPath := path.Join(logPath, logFileName)
	writer, _ := rotatelogs.New(
		baseLogPath+".%Y%m%d",
		rotatelogs.WithLinkName(baseLogPath),
		rotatelogs.WithMaxAge(maxAge),
		rotatelogs.WithRotationTime(rotationTime),
	)

	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, DefaultLoggerFormat)
	logrus.AddHook(lfHook)
}

// ---------------------------------------------------------------------------------------------------------------------
func Error(action string, mates interface{}, msg string) {
	logrus.WithFields(logrus.Fields{
		"app":    _appName,
		"action": action,
		"mates":  mates,
	}).Error(msg)
}

func Info(action string, mates interface{}, msg string) {
	logrus.WithFields(logrus.Fields{
		"app":    _appName,
		"action": action,
		"mates":  mates,
	}).Info(msg)
}

func Debug(action string, mates interface{}, msg string) {
	logrus.WithFields(logrus.Fields{
		"app":    _appName,
		"action": action,
		"mates":  mates,
	}).Debug(msg)
}

func Warning(action string, mates interface{}, msg string) {
	logrus.WithFields(logrus.Fields{
		"app":    _appName,
		"action": action,
		"mates":  mates,
	}).Warn(msg)
}
