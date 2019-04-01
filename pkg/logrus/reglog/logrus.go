// Packge reglog is purely to facilitate registration of default logger
package reglog

import (
	"github.com/sirupsen/logrus"
)

func init() {
	fmter := new(logrus.JSONFormatter)
	// fmter.PrettyPrint = true
	fmter.FieldMap = logrus.FieldMap{
		logrus.FieldKeyTime:  "@timestamp",
		logrus.FieldKeyLevel: "@level",
		logrus.FieldKeyMsg:   "@message",
		logrus.FieldKeyFunc:  "@caller",
		logrus.FieldKeyFile:  "@file",
	}
	logrus.SetFormatter(fmter)
	logrus.SetReportCaller(true)
}
