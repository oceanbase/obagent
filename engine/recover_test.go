package engine

import (
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestGoroutineProtection(_ *testing.T) {
	defer GoroutineProtection(log.WithField("recover_test", "TestGoroutineProtection"))
	var a *int
	*a = 1
}
