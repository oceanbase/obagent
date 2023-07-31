package http

import (
	"context"
	"net"
	"net/http"
	"syscall"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
)

type Listener struct {
	tcpListener  *net.TCPListener
	unixListener *net.UnixListener
	mux          *http.ServeMux
	srv          *http.Server
}

func NewListener() *Listener {
	mux := http.NewServeMux()
	return &Listener{
		mux: mux,
		srv: &http.Server{Handler: mux},
	}
}

func (l *Listener) StartTCP(addr string) error {
	tcpListener, err := NewTcpListener(addr)
	if err != nil {
		return err
	}
	go func() {
		_ = l.srv.Serve(tcpListener)
		log.Info("http tcp server exited")
	}()
	l.tcpListener = tcpListener
	return nil
}

func NewTcpListener(addr string) (*net.TCPListener, error) {
	cfg := net.ListenConfig{
		Control: reusePort,
	}
	listener, err := cfg.Listen(context.Background(), "tcp", addr)
	if err != nil {
		return nil, err
	}
	return listener.(*net.TCPListener), nil
}

func reusePort(network, address string, c syscall.RawConn) error {
	var err2 error
	err := c.Control(func(fd uintptr) {
		err2 = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		if err2 != nil {
			return
		}
		err2 = syscall.SetsockoptInt(int(fd), syscall.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	})
	if err2 != nil {
		return err2
	}
	return err
}

func (l *Listener) StartSocket(path string) error {
	listener, err := NewSocketListener(path)
	if err != nil {
		return err
	}

	go func() {
		_ = l.srv.Serve(listener)
		log.Info("http socket server exited")
	}()
	l.unixListener = listener
	return nil
}

func NewSocketListener(path string) (*net.UnixListener, error) {
	addr, err := net.ResolveUnixAddr("unix", path)
	if err != nil {
		return nil, err
	}
	return net.ListenUnix("unix", addr)
}

func (l *Listener) AddHandler(path string, h http.Handler) {
	l.mux.Handle(path, http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Connection", "close")
		h.ServeHTTP(writer, request)
	}))
}

func (l *Listener) Close() {
	_ = l.srv.Close()
	if l.tcpListener != nil {
		_ = l.tcpListener.Close()
	}
	if l.unixListener != nil {
		_ = l.unixListener.Close()
	}
}
