package errors

import (
	"testing"
)

func TestErrorCode_HasMessage(t *testing.T) {
	for _, e := range errorCodes {
		message := GetMessage(defaultLanguage, e, []interface{}{})
		if message == e.key {
			t.Errorf("ErrorCode %v(%v) has no i18n message defined", e.Code, e.key)
		}
	}
}
