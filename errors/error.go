package errors

import (
	"fmt"

	"golang.org/x/text/language"
)

// OcpAgentError defines OB-Agent specific errors.
// It implements error interface.
type OcpAgentError struct {
	ErrorCode ErrorCode     // error code
	Args      []interface{} // args for error message formatting
}

func (e OcpAgentError) Message(lang language.Tag) string {
	return GetMessage(lang, e.ErrorCode, e.Args)
}

func (e OcpAgentError) DefaultMessage() string {
	return e.Message(defaultLanguage)
}

func (e OcpAgentError) Error() string {
	return fmt.Sprintf("OcpAgentError: code = %d, message = %s", e.ErrorCode.Code, e.DefaultMessage())
}

func Occur(errorCode ErrorCode, args ...interface{}) *OcpAgentError {
	return &OcpAgentError{
		ErrorCode: errorCode,
		Args:      args,
	}
}
