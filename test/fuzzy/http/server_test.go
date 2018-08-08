package http

import (
	"testing"
	"time"

	"github.com/psh686868/psh-mosn/pkg/log"
	"github.com/psh686868/psh-mosn/pkg/protocol"
	"github.com/psh686868/psh-mosn/pkg/types"
	"github.com/psh686868/psh-mosn/test/fuzzy"
)

func runClient(t *testing.T, meshAddr string, stop chan struct{}) {
	client := NewHTTPClient(t, meshAddr)
	fuzzy.FuzzyClient(stop, client)
	<-time.After(caseDuration)
	close(stop)
	time.Sleep(5 * time.Second)
	if client.unexpectedCount != 0 {
		t.Errorf("case%d client have unexpected request: %d\n", caseIndex, client.failureCount)
	}
	if client.successCount == 0 || client.failureCount == 0 {
		t.Errorf("case%d client suucess count: %d, failure count: %d\n", caseIndex, client.successCount, client.failureCount)
	}
	log.StartLogger.Infof("[FUZZY TEST] client suucess count: %d, failure count: %d\n", client.successCount, client.failureCount)
}

func TestServerCloseProxy(t *testing.T) {
	caseIndex++
	log.StartLogger.Infof("[FUZZY TEST] HTTP Server Close In ProxyMode  %d", caseIndex)
	serverList := []string{
		"127.0.0.1:8080",
		"127.0.0.1:8081",
		"127.0.0.1:8082",
	}
	stop := make(chan struct{})
	meshAddr := fuzzy.CreateMeshProxy(t, stop, serverList, protocol.HTTP1)
	servers := CreateServers(t, serverList, stop)
	fuzzy.FuzzyServer(stop, servers, caseDuration/5)
	runClient(t, meshAddr, stop)
}

func runServerCloseMeshToMesh(t *testing.T, proto types.Protocol) {
	serverList := []string{
		"127.0.0.1:8080",
		"127.0.0.1:8081",
		"127.0.0.1:8082",
	}
	stop := make(chan struct{})
	meshAddr := fuzzy.CreateMeshCluster(t, stop, serverList, protocol.HTTP1, proto)
	servers := CreateServers(t, serverList, stop)
	fuzzy.FuzzyServer(stop, servers, caseDuration/5)
	runClient(t, meshAddr, stop)
}

func TestServerCloseToHTTP1(t *testing.T) {
	caseIndex++
	log.StartLogger.Infof("[FUZZY TEST] HTTP Server Close HTTP1 %d", caseIndex)
	runServerCloseMeshToMesh(t, protocol.HTTP1)
}
func TestServerCloseToHTTP2(t *testing.T) {
	caseIndex++
	log.StartLogger.Infof("[FUZZY TEST] HTTP Server Close HTTP2 %d", caseIndex)
	runServerCloseMeshToMesh(t, protocol.HTTP2)
}
