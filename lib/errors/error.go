package errors

import (
	"fmt"
	"path"
	"runtime"
	"strconv"
	"strings"
)

type Error struct {
	code   Code
	values []interface{}
	cause  error
	stack  string
}

type FormatFunc func(err *Error) string

var defaultFormatFunc = func(err *Error) string {
	code := err.code
	return fmt.Sprintf("Module=%s, kind=%s, code=%s; ",
		code.Module, code.Kind.Name, code.Name) + code.Format(err.values...)
}

var GlobalFormatFunc = defaultFormatFunc

func (err *Error) Error() string {
	ret := defaultFormatFunc(err)
	if err.stack != "" {
		ret += "\n" + string(err.stack)
	}
	if err.cause != nil {
		ret += "\ncause:\n"
		if causeErr, ok := err.cause.(*Error); ok {
			ret += causeErr.Error()
		} else {
			ret += err.cause.Error()
		}
	}
	return ret
}

func (err *Error) Kind() Kind {
	return err.code.Kind
}

func (err *Error) CodeName() string {
	return err.code.Name
}

func (err *Error) Module() string {
	return err.code.Module
}

func (err *Error) Code() Code {
	return err.code
}

func (err *Error) HttpCode() int {
	return err.code.Kind.HttpCode
}

func (err *Error) Message() string {
	return GlobalFormatFunc(err)
}

func (err *Error) WithCause(cause error) *Error {
	err.cause = cause
	return err
}

func (err *Error) WithStack() *Error {
	lines := make([]string, 0, 8)
	lines = append(lines, "stack:\n")
	i := 1
	for {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		fn := runtime.FuncForPC(pc)
		fnName := ""
		if fn != nil {
			fnName = fn.Name()
		}

		lines = append(lines, fnName+" ("+path.Base(file)+":"+strconv.Itoa(line)+")\n")
		i++
	}
	err.stack = strings.Join(lines, "")
	return err
}

func (err *Error) Values() []interface{} {
	return err.values
}

func (err *Error) Cause() error {
	return err.cause
}

func (err *Error) Stack() string {
	return err.stack
}

func HttpCode(err error, def int) int {
	if e, ok := err.(*Error); ok {
		return e.HttpCode()
	}
	return def
}
