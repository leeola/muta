package logging

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var defaultLogger *Logger

func init() {
	defaultLogger = NewLogger(os.Stderr)
}

type LogLevel uint8

const (
	VERBOSE LogLevel = iota
	DEBUG
	INFO
	WARN
	ERROR
)
const DEFAULT LogLevel = INFO

func LevelFromString(s string) LogLevel {
	s = strings.ToUpper(s)
	switch s {
	case "VERBOSE":
		return VERBOSE
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARN":
		return WARN
	case "ERROR":
		return ERROR
	default:
		return DEFAULT
	}
}

func NewLogger(w io.Writer) *Logger {
	return &Logger{
		writer:   w,
		logLevel: DEFAULT,
		tags:     nil,
	}
}

// A super simple tagging logger, made to satisfy the design needs
// of Muta. Hopefully to be replaced in the future with a more mature
// logger of similar style.
type Logger struct {
	writer   io.Writer
	logLevel LogLevel
	tags     []*regexp.Regexp
}

func (l *Logger) matchTags(sts []string) bool {
	if l.tags == nil {
		return true
	}
	for _, st := range sts {
		for _, rt := range l.tags {
			if rt.MatchString(st) {
				return true
			}
		}
	}
	return false
}

func (l *Logger) log(lv LogLevel, t []string, args ...interface{}) {
	if l.logLevel > lv {
		return
	}
	if l.matchTags(t) == false {
		return
	}
	if len(t) > 0 {
		args[0] = fmt.Sprintf("[%s] %s", t[0], args[0])
	}
	fmt.Fprintln(l.writer, args...)
}

func (l *Logger) logf(lv LogLevel, t []string, s string, args ...interface{}) {
	if l.logLevel > lv {
		return
	}
	if l.matchTags(t) == false {
		return
	}
	if len(t) > 0 {
		s = fmt.Sprintf("[%s] %s\n", t[0], s)
	}
	fmt.Fprintf(l.writer, s, args...)
}

// Set the tags that this logger will log. All other tags are ignored
func (l *Logger) SetTags(tags ...string) error {
	if len(tags) == 0 {
		l.tags = nil
	} else {
		var rs []*regexp.Regexp
		for _, t := range tags {
			t = strings.Replace(t, "*", ".*", -1)
			t = fmt.Sprintf("^%s$", t)
			r, err := regexp.Compile(t)
			if err != nil {
				return err
			}
			rs = append(rs, r)
		}
		l.tags = rs
	}
	return nil
}

// Set the log level that this logger wil log
func (l *Logger) SetLevel(lv LogLevel) {
	l.logLevel = lv
}

func (l *Logger) Verbose(t []string, args ...interface{}) {
	l.log(VERBOSE, t, args...)
}

func (l *Logger) Verbosef(t []string, s string, args ...interface{}) {
	l.logf(VERBOSE, t, s, args...)
}

func (l *Logger) Debug(t []string, args ...interface{}) {
	l.log(DEBUG, t, args...)
}

func (l *Logger) Debugf(t []string, s string, args ...interface{}) {
	l.logf(DEBUG, t, s, args...)
}

func (l *Logger) Info(t []string, args ...interface{}) {
	l.log(INFO, t, args...)
}

func (l *Logger) Infof(t []string, s string, args ...interface{}) {
	l.logf(INFO, t, s, args...)
}

func (l *Logger) Warn(t []string, args ...interface{}) {
	l.log(WARN, t, args...)
}

func (l *Logger) Warnf(t []string, s string, args ...interface{}) {
	l.logf(WARN, t, s, args...)
}

func (l *Logger) Error(t []string, args ...interface{}) {
	l.log(ERROR, t, args...)
}

func (l *Logger) Errorf(t []string, s string, args ...interface{}) {
	l.log(ERROR, t, args...)
}

// Default loggers

func DefaultLogger() *Logger {
	return defaultLogger
}

func SetLevel(lv LogLevel) {
	defaultLogger.SetLevel(lv)
}
func SetTags(t ...string) {
	defaultLogger.SetTags(t...)
}
func Verbose(t []string, args ...interface{}) {
	defaultLogger.Verbose(t, args...)
}

func Verbosef(t []string, s string, args ...interface{}) {
	defaultLogger.Verbosef(t, s, args...)
}

func Debug(t []string, args ...interface{}) {
	defaultLogger.Debug(t, args...)
}

func Debugf(t []string, s string, args ...interface{}) {
	defaultLogger.Debugf(t, s, args...)
}

func Info(t []string, args ...interface{}) {
	defaultLogger.Info(t, args...)
}

func Infof(t []string, s string, args ...interface{}) {
	defaultLogger.Infof(t, s, args...)
}

func Warn(t []string, args ...interface{}) {
	defaultLogger.Warn(t, args...)
}

func Warnf(t []string, s string, args ...interface{}) {
	defaultLogger.Warnf(t, s, args...)
}

func Error(t []string, args ...interface{}) {
	defaultLogger.Error(t, args...)
}

func Errorf(t []string, s string, args ...interface{}) {
	defaultLogger.Errorf(t, s, args...)
}
