package errors

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorOccur(t *testing.T) {
	e := ErrUnexpected
	args := "unexpected"
	err := Occur(e, args)
	message := err.Error()
	assert.Contains(t, message, strconv.Itoa(e.Code))
	assert.Contains(t, message, args)
}
