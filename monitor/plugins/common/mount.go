package common

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/lib/shell"
)

func GetMountPath(filePath string) string {
	_, err := os.Stat(filePath)
	if err != nil {
		log.Warnf("check filepath %s stat failed, err: %s", filePath, err)
		return ""
	}
	cmd := "df " + filePath
	command := shell.ShellImpl{}.NewCommand(cmd)
	result, err := command.ExecuteWithDebug()
	if err != nil {
		log.Warnf("get path's mount failed, filePath: %s", filePath)
		return ""
	}

	strs := strings.Split(result.Output, " ")
	mountedOn := strings.Replace(strs[len(strs)-1], "\n", "", -1)
	return mountedOn
}
