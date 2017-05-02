// Package log implements the logger of all the application, the approach for
// this logger is based on a global configuration (debug, writer and default
// fields). When you setup the logger it will set those settings on the global
// logger. Apart from this you can create new loggers with new fields and
// reuse that logger using `WithFields` or `WithField` instead of using the global one (`log.Logger`)
package log
