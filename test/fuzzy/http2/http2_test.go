package http2

import (
	"crypto/tls"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/psh686868/psh-mosn/pkg/log"
	_ "github.com/psh686868/psh-mosn/pkg/stream/http"
	_ "github.com/psh686868/psh-mosn/pkg/stream/http2"
	_ "github.com/psh686868/psh-mosn/pkg/stream/sofarpc"
	"github.com/psh686868/psh-mosn/test/fuzzy"
	"github.com/psh686868/psh-mosn/test/util"
	"golang.org/x/net/http2"
)

var (
	caseIndex    uint32 = 0
	caseDuration time.Duration
)

type HTTP2Client struct {
	Client          *http.Client
	t               *testing.T
	url             string
	unexpectedCount uint32
	successCount    uint32
	failureCount    uint32
}

func NewHTTP2Client(t *testing.T, addr string) *HTTP2Client {
	tr := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
			return net.Dial(netw, addr)
		},
	}
	httpClient := &http.Client{Transport: tr}
	return &HTTP2Client{
		Client: httpClient,
		t:      t,
		url:    fmt.Sprintf("http://%s/", addr),
	}
}
func (c *HTTP2Client) SendRequest() {
	resp, err := c.Client.Get(c.url)
	if err != nil {
		c.t.Errorf("unexpected error: %v\n", err)
		c.unexpectedCount++
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		c.successCount++
	} else {
		c.failureCount++
	}
	ioutil.ReadAll(resp.Body)
}

type HTTP2Server struct {
	util.UpstreamServer
	t       *testing.T
	ID      string
	mutex   sync.Mutex
	started bool
}

func NewHTTP2Server(t *testing.T, id string, addr string) *HTTP2Server {
	server := util.NewUpstreamHTTP2(t, addr)
	return &HTTP2Server{
		UpstreamServer: server,
		t:              t,
		ID:             id,
		mutex:          sync.Mutex{},
	}
}

//over write
func (s *HTTP2Server) Close() {
	s.UpstreamServer.Close()
	s.mutex.Lock()
	s.started = false
	s.mutex.Unlock()
}
func (s *HTTP2Server) GoServe() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.started {
		return
	}
	s.started = true
	s.UpstreamServer.GoServe()
}

func (s *HTTP2Server) ReStart() {
	s.mutex.Lock()
	check := s.started
	s.mutex.Unlock()
	if check {
		return
	}
	log.StartLogger.Infof("[FUZZY TEST] server restart #%s", s.ID)
	server := util.NewUpstreamHTTP2(s.t, s.UpstreamServer.Addr())
	s.UpstreamServer = server
	s.GoServe()
}
func (s *HTTP2Server) GetID() string {
	return s.ID
}

func CreateServers(t *testing.T, serverList []string, stop chan struct{}) []fuzzy.Server {
	var servers []fuzzy.Server
	for i, s := range serverList {
		id := fmt.Sprintf("server#%d", i)
		server := NewHTTP2Server(t, id, s)
		server.GoServe()
		go func(server fuzzy.Server) {
			<-stop
			server.Close()
		}(server)
		servers = append(servers, server)
	}
	return servers
}

//main
func TestMain(m *testing.M) {
	util.MeshLogPath = "./logs/rpc.log"
	util.MeshLogLevel = "INFO"
	log.InitDefaultLogger(util.MeshLogPath, log.INFO)
	casetime := flag.Int64("casetime", 1, "-casetime=1(min)")
	flag.Parse()
	caseDuration = time.Duration(*casetime) * time.Minute
	log.StartLogger.Infof("each case at least run %v", caseDuration)
	m.Run()
}
