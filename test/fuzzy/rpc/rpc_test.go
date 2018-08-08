package rpc

import (
	"flag"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/psh686868/psh-mosn/pkg/log"
	"github.com/psh686868/psh-mosn/pkg/protocol/sofarpc"
	_ "github.com/psh686868/psh-mosn/pkg/protocol/sofarpc/codec"
	_ "github.com/psh686868/psh-mosn/pkg/stream/http"
	_ "github.com/psh686868/psh-mosn/pkg/stream/http2"
	_ "github.com/psh686868/psh-mosn/pkg/stream/sofarpc"
	"github.com/psh686868/psh-mosn/test/fuzzy"
	"github.com/psh686868/psh-mosn/test/util"
)

var (
	caseIndex    uint32 = 0
	caseDuration time.Duration
)

// this client needs verify response's status code
// do not care stream id
type RPCStatusClient struct {
	*util.RPCClient
	addr            string
	t               *testing.T
	mutex           sync.Mutex
	streamId        uint32
	unexpectedCount uint32
	successCount    uint32
	failureCount    uint32
	started         bool
}

func NewRPCClient(t *testing.T, id string, proto string, addr string) *RPCStatusClient {
	client := util.NewRPCClient(t, id, proto)
	return &RPCStatusClient{
		RPCClient: client,
		addr:      addr,
		t:         t,
		mutex:     sync.Mutex{},
	}
}

// over write
func (c *RPCStatusClient) SendRequest() {
	c.mutex.Lock()
	check := c.started
	c.mutex.Unlock()
	if !check {
		return
	}
	ID := atomic.AddUint32(&c.streamId, 1)
	streamID := sofarpc.StreamIDConvert(ID)
	requestEncoder := c.Codec.NewStream(streamID, c)
	headers := util.BuildBoltV1Reuqest(ID)
	requestEncoder.AppendHeaders(headers, true)
}

func (c *RPCStatusClient) OnReceiveHeaders(headers map[string]string, endStream bool) {
	status, ok := headers[sofarpc.SofaPropertyHeader(sofarpc.HeaderRespStatus)]
	if !ok {
		c.t.Errorf("unexpected headers :%v\n", headers)
		c.unexpectedCount++
		return
	}
	code, err := strconv.Atoi(status)
	if err != nil {
		c.t.Errorf("unexpected status code: %s\n", status)
		c.unexpectedCount++
		return
	}
	if int16(code) == sofarpc.RESPONSE_STATUS_SUCCESS {
		c.successCount++
	} else {
		c.failureCount++
	}
}
func (c *RPCStatusClient) Connect() error {
	c.mutex.Lock()
	check := c.started
	c.mutex.Unlock()
	if check {
		return nil
	}
	if err := c.RPCClient.Connect(c.addr); err != nil {
		return err
	}
	c.mutex.Lock()
	c.started = true
	c.mutex.Unlock()
	return nil
}
func (c *RPCStatusClient) Close() {
	c.mutex.Lock()
	c.RPCClient.Close()
	c.started = false
	c.mutex.Unlock()
}
func (c *RPCStatusClient) RandomEvent(stop chan struct{}) {
	go func() {
		t := time.NewTicker(caseDuration / 5)
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				time.Sleep(util.RandomDuration(100*time.Millisecond, time.Second))
				switch rand.Intn(2) {
				case 0: //close
					c.Close()
					log.StartLogger.Infof("[FUZZY TEST] Close client #%s\n", c.ClientID)
				default: //
					log.StartLogger.Infof("[FUZZY TEST] Connect client #%s, error: %v\n", c.ClientID, c.Connect())
				}
			}
		}
	}()
}

type RPCServer struct {
	util.UpstreamServer
	t       *testing.T
	ID      string
	mutex   sync.Mutex
	started bool
}

func NewRPCServer(t *testing.T, id string, addr string) *RPCServer {
	server := util.NewUpstreamServer(t, addr, util.ServeBoltV1)
	return &RPCServer{
		UpstreamServer: server,
		t:              t,
		ID:             id,
		mutex:          sync.Mutex{},
	}
}

//over write
func (s *RPCServer) Close() {
	s.UpstreamServer.Close()
	s.mutex.Lock()
	s.started = false
	s.mutex.Unlock()
}
func (s *RPCServer) GoServe() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.started {
		return
	}
	s.started = true
	s.UpstreamServer.GoServe()
}

func (s *RPCServer) ReStart() {
	s.mutex.Lock()
	check := s.started
	s.mutex.Unlock()
	if check {
		return
	}
	log.StartLogger.Infof("[FUZZY TEST] server restart #%s", s.ID)
	server := util.NewUpstreamServer(s.t, s.UpstreamServer.Addr(), util.ServeBoltV1)
	s.UpstreamServer = server
	s.GoServe()
}
func (s *RPCServer) GetID() string {
	return s.ID
}

func CreateServers(t *testing.T, serverList []string, stop chan struct{}) []fuzzy.Server {
	var servers []fuzzy.Server
	for i, s := range serverList {
		id := fmt.Sprintf("server#%d", i)
		server := NewRPCServer(t, id, s)
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
