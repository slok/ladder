package log

import (
	"io"
	"os"
	"sync"

	"github.com/Sirupsen/logrus"
)

var (
	// Logger is the app logger
	Logger *Log
	lMutex = &sync.Mutex{}
)

// Values of the logger
var (
	lOut       io.Writer
	lDebug     bool
	lFormatter logrus.Formatter
	lFields    = Fields{}
)

// Set default logger
func init() {
	Setup(os.Stderr, Fields{}, false, false)
	Logger.Info("Finished setting up default logger")
}

// Setup configures defaulot values of the application logger, this values will be used to create new loggers and global loggers
func Setup(out io.Writer, fields Fields, json bool, debug bool) {
	lMutex.Lock()
	defer lMutex.Unlock()

	// Set package globals
	lOut = out
	lDebug = debug
	lFields = fields

	if json {
		lFormatter = new(logrus.JSONFormatter)
	} else {
		lFormatter = new(logrus.TextFormatter)
	}

	Logger = New()
	Logger.Info("Logger set up")
}

// New creates a logger based on the global logger settings
func New() *Log {
	lvl := logrus.InfoLevel
	if lDebug {
		lvl = logrus.DebugLevel
	}

	l := &Log{
		Logger: logrus.Logger{
			Out:       lOut,
			Formatter: lFormatter,
			Hooks:     make(logrus.LevelHooks),
			Level:     lvl,
		},
		fields: lFields,
	}
	return l

}

// Fields is a custom value for logger fields
type Fields map[string]interface{}

// Log is the abstraction layer type used for logging on application
type Log struct {
	logrus.Logger

	fields Fields
}

// WithFields Returns a logger based on the global one with one field
func WithFields(fields Fields) *Log {
	lg := New()

	// Custom fields
	lg.fields = fields

	// Add default fields
	lg.WithFields(lFields)
	return lg
}

// WithField returns  logger based on the global on with multiple fields
func WithField(key string, value interface{}) *Log {
	f := Fields{key: value}
	return WithFields(f)
}

// WithFields Returns a logger based on the global one with one field
func (l *Log) WithFields(fields map[string]interface{}) {
	for k, v := range fields {
		l.fields[k] = v
	}
}

// WithField returns  logger based on the global on with multiple fields
func (l *Log) WithField(key string, value interface{}) {
	l.fields[key] = value
}

// Debugf prints debug log line
func (l *Log) Debugf(format string, args ...interface{}) {
	l.Logger.WithFields(logrus.Fields(l.fields)).Debugf(format, args...)
}

// Infof prints info log line
func (l *Log) Infof(format string, args ...interface{}) {
	l.Logger.WithFields(logrus.Fields(l.fields)).Infof(format, args...)
}

// Printf prints log line
func (l *Log) Printf(format string, args ...interface{}) {
	l.Logger.WithFields(logrus.Fields(l.fields)).Printf(format, args...)
}

// Warnf prints warning log line
func (l *Log) Warnf(format string, args ...interface{}) {
	l.Logger.WithFields(logrus.Fields(l.fields)).Warnf(format, args...)
}

// Warningf prints warning log line
func (l *Log) Warningf(format string, args ...interface{}) {
	l.Logger.WithFields(logrus.Fields(l.fields)).Warningf(format, args...)
}

// Errorf prints error log line
func (l *Log) Errorf(format string, args ...interface{}) {
	l.Logger.WithFields(logrus.Fields(l.fields)).Errorf(format, args...)
}

// Fatalf prints fatal log line
func (l *Log) Fatalf(format string, args ...interface{}) {
	l.Logger.WithFields(logrus.Fields(l.fields)).Fatalf(format, args...)
}

// Panicf prints panic log line
func (l *Log) Panicf(format string, args ...interface{}) {
	l.Logger.WithFields(logrus.Fields(l.fields)).Panicf(format, args...)
}
