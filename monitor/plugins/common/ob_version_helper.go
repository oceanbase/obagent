package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

func CompareVersion(v1, v2 string) (int64, error) {
	v1 = strings.SplitN(v1, "-", 2)[0]
	v2 = strings.SplitN(v2, "-", 2)[0]
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	len1 := len(parts1)
	len2 := len(parts2)
	end := len1
	if len2 > len1 {
		end = len2
	}
	var n1, n2 int
	var err error
	for i := 0; i < end; i++ {
		if i < len(parts1) {
			n1, err = strconv.Atoi(parts1[i])
			if err != nil {
				return 0, err
			}
		} else {
			n1 = 0
		}
		if i < len(parts2) {
			n2, err = strconv.Atoi(parts2[i])
			if err != nil {
				return 0, err
			}
		} else {
			n2 = 0
		}
		if n1 == n2 {
			continue
		}
		return int64(n1 - n2), nil
	}
	return 0, nil
}

func ParseVersionComment(versionComment string) (string, error) {
	if len(versionComment) == 0 {
		return versionComment, errors.New(fmt.Sprintf("version_comment is empty"))
	}
	return strings.Split(versionComment, " ")[1], nil
}
