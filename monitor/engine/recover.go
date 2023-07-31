package engine

import (
	"github.com/sirupsen/logrus"
)

// GoroutineProtection goroutine recovery mechanism for panic
func GoroutineProtection(entry *logrus.Entry) {
	if err := recover(); err != nil {
		entry.Error("goroutine protection error : ", err)
	}
}
