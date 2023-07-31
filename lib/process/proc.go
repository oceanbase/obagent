package process

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Proc represents a process. Can used to run a program as a command or as a background service.
// Many options can be set with ProcessConfig
type Proc struct {
	lock    sync.Mutex
	conf    ProcessConfig
	cmd     *exec.Cmd
	done    chan struct{}
	err     error
	running bool
	state   ProcState
}

// Start the process. replace placeholders in Args with params
func (p *Proc) Start(params map[string]string) (err error) {
	if p.running {
		return ProcAlreadyRunningErr.NewError()
	}
	cmdConf := p.conf
	replacer := replacer(params)
	if replacer != nil {
		cmdConf.Program = replacer.Replace(cmdConf.Program)
		cmdConf.Args = replaceArgs(cmdConf.Args, replacer)
	}

	cmd, err := createCmd(cmdConf)
	if err != nil {
		return
	}

	var stdout, stderr io.ReadCloser
	var stdoutChan, stderrChan chan []byte
	defer func() {
		if err != nil {
			if stdout != nil {
				_ = stdout.Close()
			}
			if stderr != nil {
				_ = stderr.Close()
			}
		}
	}()
	if p.conf.Stdout == "" {
		stdout, err = cmd.StdoutPipe()
		if err != nil {
			return PipeOutputErr.NewError().WithCause(err)
		}
	}
	if p.conf.Stderr == "" {
		stderr, err = cmd.StderrPipe()
		if err != nil {
			return PipeOutputErr.NewError().WithCause(err)
		}
	}

	err = cmd.Start()
	if err != nil {
		return FailToStartProcessErr.NewError().WithCause(err)
	}

	p.running = true
	p.cmd = cmd
	p.done = make(chan struct{})
	p.state = ProcState{
		Running: true,
		Success: false,
		Exited:  false,
		Pid:     cmd.Process.Pid,

		StartAt: time.Now(),
	}

	if stdout != nil {
		stdoutChan = make(chan []byte, 1)
		go func() {
			b, _ := ioutil.ReadAll(stdout)
			_ = stdout.Close()
			stdoutChan <- b
			close(stdoutChan)
		}()
	}
	if stderr != nil {
		stderrChan = make(chan []byte, 1)
		go func() {
			b, _ := ioutil.ReadAll(stderr)
			_ = stderr.Close()
			stderrChan <- b
			close(stderrChan)
		}()
	}

	go func() {
		err := cmd.Wait()
		endAt := time.Now()

		var procStdout, procStderr []byte

		if stdoutChan != nil {
			procStdout = <-stdoutChan
		}
		if stderrChan != nil {
			procStderr = <-stderrChan
		}
		p.lock.Lock()
		p.err = err
		p.running = false
		p.cmd = nil
		state := cmd.ProcessState
		prevState := p.state
		p.state = ProcState{
			Running:  false,
			Success:  state.Success(),
			Pid:      prevState.Pid,
			Exited:   true,
			ExitCode: state.ExitCode(),

			StartAt: prevState.StartAt,
			EndAt:   endAt,
		}
		p.state.Stdout = procStdout
		p.state.Stderr = procStderr
		p.lock.Unlock()
		close(p.done)
	}()
	return nil
}

func replacer(params map[string]string) *strings.Replacer {
	if len(params) == 0 {
		return nil
	}
	var pairs []string
	for k, v := range params {
		pairs = append(pairs, "${"+k+"}", v)
	}
	return strings.NewReplacer(pairs...)
}

// replaceArgs replace placeholders in args with params
func replaceArgs(args []string, replacer *strings.Replacer) []string {
	if replacer == nil {
		return args
	}
	newArgs := make([]string, len(args))
	for i, arg := range args {
		newArgs[i] = replacer.Replace(arg)
	}
	return newArgs
}

func (p *Proc) signal(s os.Signal) error {
	if p.cmd == nil {
		return ProcNotExistsErr.NewError()
	}
	process := p.cmd.Process
	if process == nil {
		return ProcNotExistsErr.NewError()
	}
	err := process.Signal(s)
	if err != nil {
		return SignalErr.NewError(s.String(), p.Pid())
	}
	return nil
}

// Stop the process. Send a TERM signal.
// If ProcessConfig.KillWait is set, it will wait at most ProcessConfig.KillWait before send a KILL signal
// If ProcessConfig.FinalWait is set, it will wait at most ProcessConfig.FinalWait before process exited.
// return FinalWaitTimeoutErr when FinalWait exceed.
func (p *Proc) Stop() error {
	err := p.signal(syscall.SIGTERM)
	if err != nil {
		return err
	}
	if p.conf.KillWait > 0 {
		select {
		case <-p.Done():
		case <-time.After(p.conf.KillWait):
			err := p.signal(syscall.SIGKILL)
			if err != nil {
				return err
			}
		}
	}

	if p.conf.FinalWait > 0 {
		select {
		case <-p.Done():
		case <-time.After(p.conf.FinalWait):
			return FinalWaitTimeoutErr.NewError(p.conf.FinalWait)
		}
	}
	return nil
}

// Kill the process. Send a KILL signal.
func (p *Proc) Kill() error {
	return p.signal(syscall.SIGKILL)
}

func (p *Proc) Pid() int {
	return p.State().Pid
}

// ProcState process running state
type ProcState struct {
	//Running whether the process is running
	Running bool
	//Success whether the process is exited with code 0
	Success bool
	//Exited whether the process is exited
	Exited bool

	//Pid of the process
	Pid int
	//ExitCode of the finished process
	ExitCode int

	//Stdout all output in stdout if ProcessConfig.Stdout config not set
	Stdout []byte
	//Stderr all output in stderr if ProcessConfig.Stderr config not set
	Stderr []byte

	//StarAt time of the process started
	StartAt time.Time
	//EndAt time of the process ended
	EndAt time.Time
}

// State return the current process state
func (p *Proc) State() ProcState {
	p.lock.Lock()
	ret := p.state
	p.lock.Unlock()
	return ret
}

// Wait for the process ended and returns process state
func (p *Proc) Wait() ProcState {
	<-p.Done()
	return p.State()
}

// Done return a chan to wait for process ending
func (p *Proc) Done() <-chan struct{} {
	p.lock.Lock()
	done := p.done
	p.lock.Unlock()
	return done
}

// NewProc creates a new Proc with ProcessConfig
func NewProc(conf ProcessConfig) *Proc {
	return &Proc{
		conf: conf,
		cmd:  nil,
		done: nil,
		err:  nil,
	}
}

func createCmd(conf ProcessConfig) (*exec.Cmd, error) {
	envs := buildEnv(conf)
	sysProcAttr, err := buildSysProcAttr(conf)
	if err != nil {
		return nil, err
	}
	program, args, err := buildRLimitArgs(conf.Program, conf.Args, conf.Rlimit)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(program, args...)
	cmd.Env = envs
	cmd.SysProcAttr = sysProcAttr
	cmd.Dir = conf.Cwd
	err = setStdio(cmd, conf)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func setStdio(cmd *exec.Cmd, conf ProcessConfig) error {
	cmd.Stdin = nil
	if conf.Stdout != "" {
		stdout, err := logFile(conf.Stdout)
		if err != nil {
			return PipeOutputErr.NewError().WithCause(err)
		}
		cmd.Stdout = stdout
	}
	if conf.Stderr != "" {
		if conf.Stderr == conf.Stdout {
			cmd.Stderr = cmd.Stdout
		} else {
			stderr, err := logFile(conf.Stderr)
			if err != nil {
				return PipeOutputErr.NewError().WithCause(err)
			}
			cmd.Stderr = stderr
		}
	}
	return nil
}

func logFile(path string) (io.Writer, error) {
	return os.OpenFile(path, os.O_RDWR|os.O_APPEND|os.O_CREATE|syscall.O_CLOEXEC, 0644)
}

func buildEnv(conf ProcessConfig) []string {
	var envs []string
	if conf.InheritEnv {
		envs = os.Environ()
	}
	for k, v := range conf.Envs {
		envs = append(envs, k+"="+v)
	}
	return envs
}

func buildSysProcAttr(conf ProcessConfig) (*syscall.SysProcAttr, error) {
	cr, err := buildCredential(conf.User, conf.Group)
	if err != nil {
		return nil, CreateCmdFailErr.NewError().WithCause(err)
	}
	return &syscall.SysProcAttr{
		Foreground: false,
		Setsid:     true,
		Credential: cr,
	}, nil
}

const (
	RLimitCpuTime            = "cpu_time"
	RLimitFileSize           = "file_size"
	RLimitDataSize           = "data_size"
	RLimitStackSize          = "stack_size"
	RLimitCoreFileSize       = "core_file_size"
	RLimitResidentSet        = "resident_set"
	RLimitMemorySize         = "memory_size"
	RLimitProcess            = "processes"
	RLimitOpenFiles          = "open_files"
	RLimitLockedMemory       = "locked_memory"
	RLimitAddressSpace       = "address_space"
	RLimitVirtualMemory      = "virtual_memory"
	RLimitFileLocks          = "file_locks"
	RLimitPendingSignals     = "pending_signals"
	RLimitMsgQueueSize       = "msgqueue_size"
	RLimitSchedulingPriority = "scheduling_priority"
	RLimitNicePriority       = "nice_priority"
	RLimitRealtimePriority   = "realtime_priority"
	RLimitPipeSize           = "pipe_size"
)

var rLimitArgMap = map[string]string{
	RLimitCpuTime:            "-t",
	RLimitFileSize:           "-f",
	RLimitDataSize:           "-d",
	RLimitStackSize:          "-s",
	RLimitCoreFileSize:       "-c",
	RLimitResidentSet:        "-m",
	RLimitMemorySize:         "-m",
	RLimitProcess:            "-u",
	RLimitOpenFiles:          "-n",
	RLimitLockedMemory:       "-l",
	RLimitAddressSpace:       "-v",
	RLimitVirtualMemory:      "-v",
	RLimitFileLocks:          "-x",
	RLimitPendingSignals:     "-i",
	RLimitMsgQueueSize:       "-q",
	RLimitSchedulingPriority: "-e",
	RLimitNicePriority:       "-e",
	RLimitRealtimePriority:   "-r",
	RLimitPipeSize:           "-p",
}

func buildRLimitArgs(program string, args []string, rLimit map[string]int64) (string, []string, error) {
	if len(rLimit) == 0 {
		return program, args, nil
	}

	// use shell ulimit to set rlimit
	newProgram := "sh"

	shellCmd := "ulimit -S"
	for k, v := range rLimit {
		arg, ok := rLimitArgMap[k]
		if !ok {
			return "", nil, InvalidRlimitErr.NewError(k)
		}
		shellCmd += " " + arg + " " + strconv.FormatInt(v, 10)
	}
	shellCmd += ";exec \"${@}\""
	newArgs := []string{"-c", shellCmd, "--", program}
	newArgs = append(newArgs, args...)
	return newProgram, newArgs, nil
}

func buildCredential(uName, gName string) (*syscall.Credential, error) {
	if uName == "" && gName == "" {
		return nil, nil
	}
	var cr = &syscall.Credential{}
	if uName != "" {
		u, err := user.Lookup(uName)
		if err != nil {
			return nil, UserGroupErr.NewError(uName).WithCause(err)
		}
		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			return nil, UserGroupErr.NewError(u.Uid).WithCause(err)

		}
		cr.Uid = uint32(uid)
	}
	if gName != "" {
		g, err := user.LookupGroup(gName)
		if err != nil {
			return nil, UserGroupErr.NewError(gName).WithCause(err)
		}
		gid, err := strconv.Atoi(g.Gid)
		if err != nil {
			return nil, UserGroupErr.NewError(g.Gid).WithCause(err)
		}
		cr.Gid = uint32(gid)
	}
	return cr, nil
}

// Run a utility function to run program and get it's result synchronously
func Run(program string, args ...string) (ProcState, error) {
	proc := NewProc(ProcessConfig{
		Program: program,
		Args:    args,
	})
	err := proc.Start(nil)
	if err != nil {
		return ProcState{}, err
	}
	_ = proc.Wait()
	return proc.State(), nil
}
