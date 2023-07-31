package common

import (
	"context"
	"io/ioutil"
	"os"
	"strings"
	"syscall"

	log "github.com/sirupsen/logrus"
)

const (
	Host          = "host"
	Container     = "container"
	Unknown       = "unknown"
	cgroupFile    = "/proc/1/cgroup"
	dockerEnvFile = "/.dockerenv"
	kubeHostEnv   = "KUBERNETES_SERVICE_HOST"
)

func CheckNodeEnv(ctx context.Context) (string, error) {
	_, err := os.Stat(dockerEnvFile)
	if err == nil {
		return Container, nil
	}
	_, found := syscall.Getenv(kubeHostEnv)
	if found {
		return Container, nil
	}
	data, err := ioutil.ReadFile(cgroupFile)
	if err != nil {
		log.WithContext(ctx).WithError(err).Warnf("read cgroup file failed, file path: %s", cgroupFile)
		return Unknown, err
	}
	if strings.Contains(string(data), "docker") || strings.Contains(string(data), "kubepods") {
		return Container, nil
	}

	return Host, nil
}
