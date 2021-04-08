package log

import (
	"os"

	"github.com/ledgerhq/bitcoin-keychain/config"

	prefixed "github.com/x-cray/logrus-prefixed-formatter"

	"github.com/spf13/viper"

	"github.com/sirupsen/logrus"
)

// Logger defines a set of methods for writing application logs. Derived from and
// inspired by logrus.Entry.
type Logger interface {
	Debug(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Error(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Fatal(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
	Info(args ...interface{})
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Panic(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})
	Print(args ...interface{})
	Printf(format string, args ...interface{})
	Println(args ...interface{})
	Warn(args ...interface{})
	Warnf(format string, args ...interface{})
	Warning(args ...interface{})
	Warningf(format string, args ...interface{})
	Warningln(args ...interface{})
	Warnln(args ...interface{})
}

var defaultLogger = newLogrusLogger(*config.LoadProvider("bitcoin_keychain"))

func newLogrusLogger(cfg viper.Viper) *logrus.Logger {
	var level logrus.Level

	switch cfg.GetString("loglevel") {
	case "debug":
		level = logrus.DebugLevel
	case "warning":
		level = logrus.WarnLevel
	case "info":
		level = logrus.InfoLevel
	default:
		level = logrus.DebugLevel
	}

	if cfg.GetBool("json_logs") {
		return &logrus.Logger{
			Out:   os.Stderr,
			Level: level,
			Formatter: &logrus.JSONFormatter{
				TimestampFormat: "2006-01-02 15:04:05.100",
				FieldMap: logrus.FieldMap{
					logrus.FieldKeyTime:  "@timestamp",
					logrus.FieldKeyLevel: "@level",
					logrus.FieldKeyMsg:   "@message",
					logrus.FieldKeyFunc:  "@caller",
				},
			},
		}
	}

	return &logrus.Logger{
		Out:   os.Stderr,
		Level: level,
		Formatter: &prefixed.TextFormatter{
			TimestampFormat:  "2006-01-02 15:04:05.100",
			FullTimestamp:    true,
			QuoteEmptyFields: true,
			SpacePadding:     50,
			ForceFormatting:  true,
			ForceColors:      true,
		},
	}
}

// Fields is a map string interface to define fields in the structured log
type Fields map[string]interface{}

// With allow us to define fields in out structured logs
func (f Fields) With(k string, v interface{}) Fields {
	f[k] = v
	return f
}

// WithFields allow us to define fields in out structured logs
func (f Fields) WithFields(f2 Fields) Fields {
	for k, v := range f2 {
		f[k] = v
	}

	return f
}

// WithFields allow us to define fields in out structured logs
func WithFields(fields Fields) Logger {
	return defaultLogger.WithFields(logrus.Fields(fields))
}

// Debug package-level convenience method.
func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

// Debugf package-level convenience method.
func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

// Debugln package-level convenience method.
func Debugln(args ...interface{}) {
	defaultLogger.Debugln(args...)
}

// Error package-level convenience method.
func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

// Errorf package-level convenience method.
func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

// Errorln package-level convenience method.
func Errorln(args ...interface{}) {
	defaultLogger.Errorln(args...)
}

// Fatal package-level convenience method.
func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

// Fatalf package-level convenience method.
func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

// Fatalln package-level convenience method.
func Fatalln(args ...interface{}) {
	defaultLogger.Fatalln(args...)
}

// Info package-level convenience method.
func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

// Infof package-level convenience method.
func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

// Infoln package-level convenience method.
func Infoln(args ...interface{}) {
	defaultLogger.Infoln(args...)
}

// Panic package-level convenience method.
func Panic(args ...interface{}) {
	defaultLogger.Panic(args...)
}

// Panicf package-level convenience method.
func Panicf(format string, args ...interface{}) {
	defaultLogger.Panicf(format, args...)
}

// Panicln package-level convenience method.
func Panicln(args ...interface{}) {
	defaultLogger.Panicln(args...)
}

// Print package-level convenience method.
func Print(args ...interface{}) {
	defaultLogger.Print(args...)
}

// Printf package-level convenience method.
func Printf(format string, args ...interface{}) {
	defaultLogger.Printf(format, args...)
}

// Println package-level convenience method.
func Println(args ...interface{}) {
	defaultLogger.Println(args...)
}

// Warn package-level convenience method.
func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

// Warnf package-level convenience method.
func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

// Warning package-level convenience method.
func Warning(args ...interface{}) {
	defaultLogger.Warning(args...)
}

// Warningf package-level convenience method.
func Warningf(format string, args ...interface{}) {
	defaultLogger.Warningf(format, args...)
}

// Warningln package-level convenience method.
func Warningln(args ...interface{}) {
	defaultLogger.Warningln(args...)
}

// Warnln package-level convenience method.
func Warnln(args ...interface{}) {
	defaultLogger.Warnln(args...)
}
