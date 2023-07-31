package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorCode_CodeDistinct(t *testing.T) {
	m := map[int]ErrorCode{}
	for _, e := range errorCodes {
		if e2, ok := m[e.Code]; ok {
			assert.False(t, ok,
				"conflict code %v, both used by %v and %v", e.Code, e.key, e2.key)
		}
		m[e.Code] = e
	}
}

func TestErrorCode_KeyDistinct(t *testing.T) {
	m := map[string]ErrorCode{}
	for _, e := range errorCodes {
		if e2, ok := m[e.key]; ok {
			assert.False(t, ok,
				"conflict code %v, both used by %v and %v", e.key, e.Code, e2.Code)
		}
		m[e.key] = e
	}
}
