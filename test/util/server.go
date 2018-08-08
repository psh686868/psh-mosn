package util

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/http2"
)

type UpstreamServer interface {
	GoServe()
	Close()
	Addr() string
}

type ServeConn func(t *testing.T, conn net.Conn)
type upstreamServer struct {
	Listener net.Listener
	Serve    ServeConn
	Address  string
	conns    []net.Conn
	mu       sync.Mutex
	t        *testing.T
	started  bool
}

func NewUpstreamServer(t *testing.T, addr string, serve ServeConn) UpstreamServer {
	return &upstreamServer{
		Serve:   serve,
		Address: addr,
		conns:   []net.Conn{},
		mu:      sync.Mutex{},
		t:       t,
	}
}
func (s *upstreamServer) GoServe() {
	go s.serve()
}
func (s *upstreamServer) serve() {
	//try listen
	var err error
	for i := 0; i < 3; i++ {
		s.Listener, err = net.Listen("tcp", s.Address)
		if s.Listener != nil {
			break
		}
		time.Sleep(time.Second)
	}
	if s.Listener == nil {
		s.t.Fatalf("listen %s failed, error : %v\n", s.Address, err)
	}
	s.mu.Lock()
	s.started = true
	s.mu.Unlock()
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				s.t.Logf("Accept error: %v\n", err)
				continue
			}
			return
		}
		s.t.Logf("server %s Accept connection: %s\n", s.Listener.Addr().String(), conn.RemoteAddr().String())
		s.mu.Lock()
		s.conns = append(s.conns, conn)
		s.mu.Unlock()
		go s.Serve(s.t, conn)
	}
}

func (s *upstreamServer) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.started {
		return
	}
	s.t.Logf("server %s closed\n", s.Address)
	s.Listener.Close()
	for _, conn := range s.conns {
		conn.Close()
	}
	s.started = false
}
func (s *upstreamServer) Addr() string {
	return s.Address
}

//Server Implement
type HTTP2Server struct {
	t      *testing.T
	Server *http2.Server
}

func (s *HTTP2Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//s.t.Logf("[server] Receive request : %s\n", r.Header.Get("Requestid"))
	w.Header().Set("Content-Type", "text/plain")

	for k := range r.Header {
		w.Header().Set(k, r.Header.Get(k))
	}

	fmt.Fprintf(w, "\nRequestId:%s\n", r.Header.Get("Requestid"))

}
func (s *HTTP2Server) ServeConn(t *testing.T, conn net.Conn) {
	s.Server.ServeConn(conn, &http2.ServeConnOpts{Handler: s})
}

func NewUpstreamHTTP2(t *testing.T, addr string) UpstreamServer {
	s := &HTTP2Server{
		t:      t,
		Server: &http2.Server{IdleTimeout: 1 * time.Minute},
	}
	return NewUpstreamServer(t, addr, s.ServeConn)
}

//Http Server
type HTTPServer struct {
	t      *testing.T
	server *httptest.Server
}

func (s *HTTPServer) GoServe() {
	s.server.Start()
}
func (s *HTTPServer) Close() {
	s.server.Close()
}
func (s *HTTPServer) Addr() string {
	addr := strings.Split(s.server.URL, "http://")
	if len(addr) == 2 {
		return addr[1]
	}
	return ""
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//s.t.Log("HTTP server receive data")
	for k := range r.Header {
		w.Header().Set(k, r.Header.Get(k))
	}
	fmt.Fprintf(w, "\nRequestId:%s\n", r.Header.Get("Requestid"))
}

func NewHTTPServer(t *testing.T) UpstreamServer {
	s := &HTTPServer{t: t}
	s.server = httptest.NewUnstartedServer(s)
	return s
}
