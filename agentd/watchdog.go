package agentd

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/process"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"

	"github.com/oceanbase/obagent/agentd/api"
	"github.com/oceanbase/obagent/config"
	"github.com/oceanbase/obagent/executor/agent"
	"github.com/oceanbase/obagent/lib/command"
	http2 "github.com/oceanbase/obagent/lib/http"
	path2 "github.com/oceanbase/obagent/lib/path"
)

// Agentd supervisor process for agents
// start, stop sub services, view status of self and sub services
type Agentd struct {
	config   Config
	services map[string]*Service
	listener *http2.Listener
	state    *http2.StateHolder
}

// NewAgentd create a new Agentd via config
func NewAgentd(config Config) *Agentd {
	listener := http2.NewListener()
	services := make(map[string]*Service)
	for name, svcConf := range config.Services {
		svc := NewService(name, svcConf)
		services[name] = svc
	}
	ret := &Agentd{
		config:   config,
		services: services,
		listener: listener,
		state:    http2.NewStateHolder(http2.Stopped),
	}
	ret.initRoutes()
	return ret
}

var startAt = time.Now().UnixNano()

func (w *Agentd) initRoutes() {
	rootPath := "/api/v1"
	statusHandler := http2.NewHandler(command.WrapFunc(w.Status))
	startServiceHandler := http2.NewHandler(command.WrapFunc(func(param api.StartStopAgentParam) error {
		return w.StartService(param.Service)
	}))
	stopServiceHandler := http2.NewHandler(command.WrapFunc(func(param api.StartStopAgentParam) error {
		return w.StopService(param.Service)
	}))
	w.listener.AddHandler(path.Join(rootPath, "/status"), statusHandler)
	w.listener.AddHandler(path.Join(rootPath, "/startService"), startServiceHandler)
	w.listener.AddHandler(path.Join(rootPath, "/stopService"), stopServiceHandler)

	w.listener.AddHandler("/debug/pprof/", http.HandlerFunc(pprof.Index))
	w.listener.AddHandler("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	w.listener.AddHandler("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	w.listener.AddHandler("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	w.listener.AddHandler("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
}

// Start agentd and sub services
func (w *Agentd) Start() error {
	w.cleanup()
	if w.state.Get() == http2.Running {
		return nil
	}
	log.Info("starting agentd")
	var err error
	err = w.writePid()
	if err != nil {
		return err
	}
	w.state.Set(http2.Starting)
	err = os.MkdirAll(w.config.RunDir, 0755)
	if err != nil {
		return err
	}

	socketPath := w.socketPath()
	log.Infof("starting socket listener on '%s'", socketPath)
	err = w.listener.StartSocket(socketPath)
	if err != nil {
		log.Errorf("start socket listener '%s' failed: %v", socketPath, err)
		_ = w.Stop()
		return err
	}
	w.state.Set(http2.Running)

	for name, svc := range w.services {
		log.Infof("starting service '%s'", name)
		err = svc.Start()
		if err != nil {
			log.Errorf("start service '%s' failed: %s", name, err)
		}
	}
	log.Info("agentd started")
	return nil
}

// Stop agentd and sub services
func (w *Agentd) Stop() error {
	if w.state.Get() == http2.Stopped {
		return nil
	}
	log.Info("stopping agentd")
	if !w.state.Cas(http2.Running, http2.Stopping) {
		return AgentdNotRunningErr.NewError(w.state.Get())
	}

	for name, svc := range w.services {
		state := svc.State().State
		log.Infof("stopping service '%s'. current state: %s", name, state)
		if state == http2.Stopped {
			log.Infof("service '%s' already stopped. State: %s", name, state)
			continue
		}
		err := svc.Stop()
		if err != nil {
			log.Errorf("stop service '%s' got error: %v", name, err)
		}
	}

	w.state.Set(http2.Stopped)
	log.Info("agentd stopped")

	w.listener.Close()

	err := w.removePid()
	if err != nil {
		log.Warn("remove pid file failed: ", err)
	}
	w.cleanup()
	return nil
}

// ListenSignal capture SIGTERM and SIGINT, do a normal Stop
func (w *Agentd) ListenSignal() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
	select {
	case sig := <-ch:
		log.Infof("signal '%s' received. exiting...", sig.String())
		_ = w.Stop()
	}
}

// Status returns agentd and sub services status
func (w *Agentd) Status() api.Status {
	svcStates := make(map[string]api.ServiceStatus)
	ready := w.state.Get() == http2.Running

	for name, svc := range w.services {
		state := svc.State()
		svcStates[name] = state
		if state.State != http2.Running {
			ready = false
		}
	}
	dangling := w.danglingServices()
	socket := w.socketPath()
	if !isSocket(socket) {
		socket = ""
	}
	return api.Status{
		State:    w.state.Get(),
		Ready:    ready,
		Pid:      os.Getpid(),
		Socket:   socket,
		Services: svcStates,
		Dangling: dangling,
		Version:  config.AgentVersion,
		StartAt:  startAt,
	}
}

func (w *Agentd) StartService(name string) error {
	service, ok := w.services[name]
	if !ok {
		return ServiceNotFoundErr.NewError(name)
	}
	return service.Start()
}

func (w *Agentd) StopService(name string) error {
	service, ok := w.services[name]
	if !ok {
		return ServiceNotFoundErr.NewError(name)
	}
	return service.Stop()
}

func (w *Agentd) socketPath() string {
	return agent.SocketPath(w.config.RunDir, path2.Agentd, os.Getpid())
}

func (w *Agentd) pidPath() string {
	return agent.PidPath(w.config.RunDir, path2.Agentd)
}

func (w *Agentd) writePid() error {
	f, err := os.OpenFile(w.pidPath(), os.O_RDWR|os.O_CREATE|os.O_EXCL|os.O_SYNC|syscall.O_CLOEXEC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%d", os.Getpid())
	if err != nil {
		return err
	}
	return nil
}

func (w *Agentd) removePid() error {
	return removePid(w.pidPath(), os.Getpid())
}

func (w *Agentd) cleanupPid(program, pidPath string) error {
	if _, err := os.Stat(pidPath); err != nil && os.IsNotExist(err) {
		return nil
	}
	pid, err := agent.ReadPid(pidPath)
	if err != nil {
		return err
	}
	_, err = process.NewProcess(int32(pid))
	if err != nil {
		return removePid(pidPath, pid)
	}
	return nil
}

func (w *Agentd) cleanupPidPattern(program, pattern string) {
	pidPaths, err := filepath.Glob(filepath.Join(w.config.RunDir, pattern))
	if err != nil {
		return
	}
	for _, pidPath := range pidPaths {
		err = w.cleanupPid(program, pidPath)
		if err != nil {
			log.WithError(err).Errorf("cleanup pid file %s got error", pidPath)
		}
	}
}

func (w *Agentd) cleanupSocketPattern(pattern string) {
	sockPaths, err := filepath.Glob(filepath.Join(w.config.RunDir, pattern))
	if err != nil {
		return
	}
	for _, sockPath := range sockPaths {
		err = w.cleanupSocket(sockPath)
		if err != nil {
			log.WithError(err).Errorf("cleanup socket file %s got error", sockPath)
		}
	}
}

func (w *Agentd) cleanupSocket(sockPath string) error {
	if !isSocket(sockPath) {
		return nil
	}
	if http2.CanConnect("unix", sockPath, time.Second) {
		return nil
	}
	return os.Remove(sockPath)
}

func (w *Agentd) cleanupDangling() {
	if !w.config.CleanupDangling {
		return
	}
	n := len(w.danglingServices())
	if n == 0 {
		return
	}
	log.Infof("cleaning up dangling services, %d to cleanup", n)
	for _, dangling := range w.danglingServices() {
		log.Infof("cleaning up dangling service: %s %d", dangling.Name, dangling.Pid)
		proc, err := process.NewProcess(int32(dangling.Pid))
		if err != nil {
			log.Warnf("get process of dangling service: %s %d got error %v", dangling.Name, dangling.Pid, err)
			continue
		}
		err = proc.SendSignal(unix.SIGTERM)
		if err != nil {
			log.Warnf("terminate process of dangling service: %s %d got error %v", dangling.Name, dangling.Pid, err)
			continue
		}
	}
}

func (w *Agentd) cleanup() {
	var err error
	w.cleanupDangling()
	w.cleanupPidPattern(path2.Agentd, path2.Agentd+".pid")
	w.cleanupPidPattern(path2.Agentd, path2.Agentd+".*.pid")
	w.cleanupSocketPattern(path2.Agentd + ".*.sock")

	pidPath := filepath.Join(w.config.RunDir, path2.Agentd+".pid")
	err = w.cleanupPid(path2.Agentd, pidPath)
	if err != nil {
		log.WithError(err).Errorf("cleanup pid file %s got error", pidPath)
	}

	if err != nil {
		log.WithError(err).Errorf("cleanup pid file %s got error", pidPath)
	}

	for name, conf := range w.config.Services {
		w.cleanupPidPattern(conf.Program, fmt.Sprintf("%s.pid", name))
		w.cleanupPidPattern(conf.Program, fmt.Sprintf("%s.*.pid", name))
		w.cleanupSocketPattern(fmt.Sprintf("%s.*.sock", name))
	}
}

func (w *Agentd) danglingServices() []api.DanglingService {
	var ret []api.DanglingService
	for name := range w.config.Services {
		pidPath := agent.PidPath(w.config.RunDir, name)
		pid, err := agent.ReadPid(pidPath)
		if err == nil {
			if w.isDangling(name, pid) {
				socket := agent.SocketPath(w.config.RunDir, name, pid)
				if !isSocket(socket) {
					socket = ""
				}
				ret = append(ret, api.DanglingService{
					Name:    name,
					Pid:     pid,
					PidFile: pidPath,
					Socket:  socket,
				})
			}
		}
		pidPaths, err := filepath.Glob(filepath.Join(w.config.RunDir, fmt.Sprintf("%s.*.pid", name)))
		if err == nil {
			for _, pidPath = range pidPaths {
				pid, err = agent.ReadPid(pidPath)
				if err != nil {
					continue
				}
				if w.isDangling(name, pid) {
					socket := agent.SocketPath(w.config.RunDir, name, pid)
					if !isSocket(socket) {
						socket = ""
					}
					ret = append(ret, api.DanglingService{
						Name:    name,
						Pid:     pid,
						PidFile: pidPath,
						Socket:  socket,
					})
				}
			}
		}
	}
	return ret
}

func (w *Agentd) isDangling(program string, pid int) bool {
	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		return false
	}
	name, err := proc.Name()
	if err != nil {
		return false
	}
	if name != program {
		return false
	}
	ppid, err := proc.Ppid()
	if err != nil {
		return false
	}
	return ppid == 1 || ppid == 0
}
