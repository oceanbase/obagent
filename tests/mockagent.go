package main

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/oceanbase/obagent/executor/agent"
	"github.com/oceanbase/obagent/lib/http"
)

func main() {
	if len(os.Args) < 5 {
		_, _ = fmt.Fprintf(os.Stderr, "not enough args")
		os.Exit(1)
	}
	name := os.Args[1]
	runDir := os.Args[2]
	os.Args[0] = name
	n, err := strconv.Atoi(os.Args[3])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	startWait := time.Duration(n) * time.Second
	n, err = strconv.Atoi(os.Args[3])
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
	stopWait := time.Duration(n) * time.Second

	log.Infof("mock agent starting name='%s' runDir='%s' startWait=%d stopWait=%d", name, runDir, startWait, stopWait)

	state := http.Starting
	startAt := time.Now().UnixNano()
	listener := http.NewListener()
	socket := agent.SocketPath(runDir, name, os.Getpid())
	listener.AddHandler("/api/v1/status", http.NewFuncHandler(func() http.Status {
		return http.Status{
			State:   http.State(atomic.LoadInt32((*int32)(&state))),
			Pid:     os.Getpid(),
			StartAt: startAt,
		}
	}))
	err = listener.StartSocket(socket)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(2)
	}
	if startWait < 0 {
		time.Sleep(-startWait)
		os.Exit(3)
	}
	time.Sleep(startWait)
	atomic.StoreInt32((*int32)(&state), int32(http.Running))
	log.Info("mock agent started")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-ch:
		log.Infof("signal '%s' received. exiting...", sig.String())
		atomic.StoreInt32((*int32)(&state), int32(http.Stopping))
		time.Sleep(stopWait)
		atomic.StoreInt32((*int32)(&state), int32(http.Stopped))
		listener.Close()
		log.Info("mock agent exited")
	}
}
