package log

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func Fields(kv ...interface{}) *logrus.Entry {
	length := (len(kv) - 1)
	if length%2 == 1 {
		length -= 1
	}
	fields := logrus.Fields(make(map[string]interface{}, length))
	for i := 0; i < length<<1; i++ {
		fields[fmt.Sprint(kv[i>>1])] = kv[i>>1+1]
	}
	return logrus.WithFields(fields)
}
