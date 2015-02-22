package muta

import "github.com/leeola/muta/logging"

// Add two simple shortcuts for Info
func Log(t []string, args ...interface{}) {
	logging.Info(t, args...)
}

func Logf(t []string, m string, args ...interface{}) {
	logging.Infof(t, m, args...)
}
