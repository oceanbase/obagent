package command

import "errors"

var TimeoutErr = errors.New("command: wait timeout")
var AlreadyFinishedErr = errors.New("command: command already finished")
var ExecutionNotFoundErr = errors.New("command: execution not found")
var ExecutionAlreadyExistsErr = errors.New("command: execution already exists")
var BadTaskFunc = errors.New("command: bad task function")
