//go:build !linux
// +build !linux

package nodeexporter

const (
	ntp      = "collector.ntp"
	textfile = "collector.textfile"
	timec    = "collector.time"
	uname    = "collector.uname"
)

var collectItems = []string{
	ntp,
}

var noCollectItems = []string{
	textfile,
	timec,
	uname,
}
