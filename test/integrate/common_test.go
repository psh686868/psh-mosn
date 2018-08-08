package integrate

import (
	"testing"
	"time"

	"github.com/psh686868/psh-mosn/pkg/protocol"
	_ "github.com/psh686868/psh-mosn/pkg/protocol/sofarpc/codec"
	"github.com/psh686868/psh-mosn/test/util"
)

// Notice can't use APP(HTTPX) to MESH(SofaRPC),
// because SofaRPC is a group of protocols,such as boltV1, boltV2.
func TestCommon(t *testing.T) {
	appaddr := "127.0.0.1:8080"
	testCases := []*testCase{
		newTestCase(t, protocol.HTTP1, protocol.HTTP1, util.NewHTTPServer(t)),
		newTestCase(t, protocol.HTTP1, protocol.HTTP2, util.NewHTTPServer(t)),
		newTestCase(t, protocol.HTTP2, protocol.HTTP1, util.NewUpstreamHTTP2(t, appaddr)),
		newTestCase(t, protocol.HTTP2, protocol.HTTP2, util.NewUpstreamHTTP2(t, appaddr)),
		newTestCase(t, protocol.SofaRPC, protocol.HTTP1, util.NewRPCServer(t, appaddr, util.Bolt1)),
		newTestCase(t, protocol.SofaRPC, protocol.HTTP2, util.NewRPCServer(t, appaddr, util.Bolt1)),
		newTestCase(t, protocol.SofaRPC, protocol.SofaRPC, util.NewRPCServer(t, appaddr, util.Bolt1)),
		//TODO:
		//newTestCase(t, protocol.SofaRPC, protocol.HTTP1, util.NewRPCServer(t, appaddr, util.Bolt2)),
		//newTestCase(t, protocol.SofaRPC, protocol.HTTP2, util.NewRPCServer(t, appaddr, util.Bolt2)),
		//newTestCase(t, protocol.SofaRPC, protocol.SofaRPC, util.NewRPCServer(t, appaddr, util.Bolt2)),
	}
	for i, tc := range testCases {
		t.Logf("start case #%d\n", i)
		tc.Start(false)
		go tc.RunCase(1)
		select {
		case err := <-tc.C:
			if err != nil {
				t.Errorf("[ERROR MESSAGE] #%d %v to mesh %v test failed, error: %v\n", i, tc.AppProtocol, tc.MeshProtocol, err)
			}
		case <-time.After(15 * time.Second):
			t.Errorf("[ERROR MESSAGE] #%d %v to mesh %v hang\n", i, tc.AppProtocol, tc.MeshProtocol)
		}
		close(tc.stop)
		time.Sleep(time.Second)
	}
}

func TestTLS(t *testing.T) {
	appaddr := "127.0.0.1:8080"
	testCases := []*testCase{
		newTestCase(t, protocol.HTTP1, protocol.HTTP1, util.NewHTTPServer(t)),
		newTestCase(t, protocol.HTTP1, protocol.HTTP2, util.NewHTTPServer(t)),
		newTestCase(t, protocol.HTTP2, protocol.HTTP1, util.NewUpstreamHTTP2(t, appaddr)),
		newTestCase(t, protocol.HTTP2, protocol.HTTP2, util.NewUpstreamHTTP2(t, appaddr)),
		newTestCase(t, protocol.SofaRPC, protocol.HTTP1, util.NewRPCServer(t, appaddr, util.Bolt1)),
		newTestCase(t, protocol.SofaRPC, protocol.HTTP2, util.NewRPCServer(t, appaddr, util.Bolt1)),
		newTestCase(t, protocol.SofaRPC, protocol.SofaRPC, util.NewRPCServer(t, appaddr, util.Bolt1)),
		//TODO:
		//newTestCase(t, protocol.SofaRPC, protocol.HTTP1, util.NewRPCServer(t, appaddr, util.Bolt2)),
		//newTestCase(t, protocol.SofaRPC, protocol.HTTP2, util.NewRPCServer(t, appaddr, util.Bolt2)),
		//newTestCase(t, protocol.SofaRPC, protocol.SofaRPC, util.NewRPCServer(t, appaddr, util.Bolt2)),

	}
	for i, tc := range testCases {
		t.Logf("start case #%d\n", i)
		tc.Start(true)
		go tc.RunCase(1)
		select {
		case err := <-tc.C:
			if err != nil {
				t.Errorf("[ERROR MESSAGE] #%d %v to mesh %v tls test failed, error: %v\n", i, tc.AppProtocol, tc.MeshProtocol, err)
			}
		case <-time.After(15 * time.Second):
			t.Errorf("[ERROR MESSAGE] #%d %v to mesh %v hang\n", i, tc.AppProtocol, tc.MeshProtocol)
		}
		close(tc.stop)
		time.Sleep(time.Second)
	}

}
