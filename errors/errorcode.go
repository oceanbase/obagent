package errors

import "net/http"

type ErrorKind = int

const (
	badRequest      ErrorKind = http.StatusBadRequest
	illegalArgument ErrorKind = http.StatusBadRequest
	notFound        ErrorKind = http.StatusNotFound
	unexpected      ErrorKind = http.StatusInternalServerError
	notImplemented  ErrorKind = http.StatusNotImplemented
)

type ErrorCode struct {
	Code int
	Kind ErrorKind
	key  string
}

var errorCodes []ErrorCode

func NewErrorCode(code int, kind ErrorKind, key string) ErrorCode {
	errorCode := ErrorCode{
		Code: code,
		Kind: kind,
		key:  key,
	}
	errorCodes = append(errorCodes, errorCode)
	return errorCode
}

var (
	// general error codes, range: 1000 ~ 1999
	ErrBadRequest      = NewErrorCode(1000, badRequest, "err.bad.request")
	ErrIllegalArgument = NewErrorCode(1001, illegalArgument, "err.illegal.argument")
	ErrUnexpected      = NewErrorCode(1002, unexpected, "err.unexpected")

	// shell execute error codes
	ErrExecuteCommand = NewErrorCode(1500, unexpected, "err.execute.command")

	// file error codes - file, range: 2000 ~ 2099
	ErrDownloadFile    = NewErrorCode(2000, unexpected, "err.download.file")
	ErrInvalidChecksum = NewErrorCode(2001, unexpected, "err.invalid.checksum")
	ErrWriteFile       = NewErrorCode(2002, unexpected, "err.write.file")
	ErrFindFile        = NewErrorCode(2003, unexpected, "err.find.file")
	ErrCheckFileExists = NewErrorCode(2004, unexpected, "err.check.file.exists")

	// file error codes - directory, range: 2100 ~ 2199
	ErrCreateDirectory = NewErrorCode(2100, unexpected, "err.create.directory")
	ErrRemoveDirectory = NewErrorCode(2101, unexpected, "err.remove.directory")
	ErrChownDirectory  = NewErrorCode(2102, unexpected, "err.chown.directory")

	// file error codes - symlink, range: 2200 ~ 2299
	ErrCreateSymlink = NewErrorCode(2200, unexpected, "err.create.symlink")
	ErrProcessCGroup = NewErrorCode(2201, unexpected, "err.process.cgroup")

	// task error codes, range: 2300 ~ 2399
	ErrTaskNotFound = NewErrorCode(2300, notFound, "err.task.not.found")

	// software package error codes, range: 3000 ~ 3999
	ErrQueryPackage     = NewErrorCode(3000, unexpected, "err.query.package")
	ErrInstallPackage   = NewErrorCode(3001, unexpected, "err.install.package")
	ErrUninstallPackage = NewErrorCode(3002, unexpected, "err.uninstall.package")
	ErrExtractPackage   = NewErrorCode(3003, unexpected, "err.extract.package")

	// system management error codes, range: 4000 ~ 4999
	ErrCheckProcessExists = NewErrorCode(4000, unexpected, "err.check.process.exists")
	ErrGetProcessInfo     = NewErrorCode(4001, unexpected, "err.get.process.info")
	ErrStopProcess        = NewErrorCode(4002, unexpected, "err.stop.process")
	ErrProcessProcInfo    = NewErrorCode(4003, unexpected, "err.get.process.proc")

	ErrGetDiskUsage      = NewErrorCode(4100, unexpected, "err.system.disk.get.usage")
	ErrBatchGetDiskInfos = NewErrorCode(4101, unexpected, "err.system.disk.batch.get.disk.infos")

	// ob operation error codes, range: 10000 ~ 10999
	ErrObInstallPreCheck       = NewErrorCode(10000, unexpected, "err.ob.install.pre-check")
	ErrObIoBench               = NewErrorCode(10001, unexpected, "err.ob.io.bench")
	ErrStartObServerProcess    = NewErrorCode(10002, unexpected, "err.observer.start")
	ErrCheckObServerAccessible = NewErrorCode(10003, unexpected, "err.check.observer.accessible")
	ErrBootstrap               = NewErrorCode(10004, unexpected, "err.ob.bootstrap")
	ErrCleanObDataFiles        = NewErrorCode(10005, unexpected, "err.clean.ob.data.files")
	ErrCleanObAllFiles         = NewErrorCode(10006, unexpected, "err.clean.ob.all.files")
	ErrRunUpgradeScript        = NewErrorCode(10007, unexpected, "err.run.upgrade.script")

	// agent admin error codes, range: 13000 ~ 13999
	ErrAgentdRunning       = NewErrorCode(13000, unexpected, "err.agent.agentd.already.running")
	ErrAgentdNotRunning    = NewErrorCode(13001, unexpected, "err.agent.agentd.not.running")
	ErrAgentdExitedQuickly = NewErrorCode(13002, unexpected, "err.agent.agentd.exited.quickly")

	// monagent pipeline manager error codes, range: 14000 ~ 14999
	ErrMonPipelineStart     = NewErrorCode(14000, unexpected, "err.monagent.pipeline.already.start")
	ErrMonPipelineStartFail = NewErrorCode(14001, unexpected, "err.monagent.pipeline.start.failed")
	ErrRemoveMonPipeline    = NewErrorCode(14002, unexpected, "err.monagent.remove.pipeline")
)
