package log_tailer

import (
	"os"
	"time"
)

type logFileInfo struct {
	logSourceType   string
	logAnalyzerType string
	fileName        string
	fileDesc        *os.File
	fileOffset      int64
	offsetLineLogAt time.Time
	isRenamed       bool
}

// RecoveryInfo Persistence information to record the query location when closed
type RecoveryInfo struct {
	FileName   string    `json:"fileName" yaml:"fileName"`
	FileId     uint64    `json:"fileId" yaml:"fileId"`
	DevId      uint64    `json:"devId" yaml:"devId"`
	FileOffset int64     `json:"fileOffset" yaml:"fileOffset"`
	TimePoint  time.Time `json:"timePoint" yaml:"timePoint"`
}

func (r RecoveryInfo) GetFileId() uint64 {
	return r.FileId
}

func (r RecoveryInfo) GetDevId() uint64 {
	return r.DevId
}
