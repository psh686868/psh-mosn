package util

import (
	"fmt"
	"sync/atomic"

	"github.com/psh686868/psh-mosn/pkg/config"
	"github.com/psh686868/psh-mosn/pkg/types"
	"github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// use different mesh port to avoid "port in used" error
var meshIndex uint32

func CurrentMeshAddr() string {
	var basic uint32 = 2044
	atomic.AddUint32(&meshIndex, 1)
	return fmt.Sprintf("127.0.0.1:%d", basic+meshIndex)
}

// mesh as a proxy , client and servre have same protocol
func CreateProxyMesh(addr string, hosts []string, proto types.Protocol) *config.MOSNConfig {
	clusterName := "proxyCluster"
	cmconfig := config.ClusterManagerConfig{
		Clusters: []config.ClusterConfig{
			newBasicCluster(clusterName, hosts),
		},
	}
	routers := []config.Router{
		newPrefixRouter(clusterName, "/"),
		newHeaderRouter(clusterName, ".*"),
	}
	chains := []config.FilterChain{
		newFilterChain("proxyVirtualHost", proto, proto, routers),
	}
	listener := newListener("proxyListener", addr, chains)
	return newMOSNConfig([]config.ListenerConfig{listener}, cmconfig)
}

// Mesh to Mesh
// clientaddr and serveraddr is mesh's addr
// appproto is client and server (not mesh) protocol
// meshproto is mesh's protocol
// hosts is server's addresses
func CreateMeshToMeshConfig(clientaddr string, serveraddr string, appproto types.Protocol, meshproto types.Protocol, hosts []string, tls bool) *config.MOSNConfig {
	downstreamCluster := "downstream"
	upstreamCluster := "upstream"
	downstreamRouters := []config.Router{
		newPrefixRouter(downstreamCluster, "/"),
		newHeaderRouter(downstreamCluster, ".*"),
	}
	clientChains := []config.FilterChain{
		newFilterChain("downstreamFilter", appproto, meshproto, downstreamRouters),
	}
	clientListener := newListener("downstreamListener", clientaddr, clientChains)
	upstreamRouters := []config.Router{
		newPrefixRouter(upstreamCluster, "/"),
		newHeaderRouter(upstreamCluster, ".*"),
	}
	// client mesh -> cluster need tls
	var meshClusterConfig config.ClusterConfig
	//  server mesh listener need tls
	var meshServerChain config.FilterChain
	if tls {
		tlsConf := config.TLSConfig{
			Status:     true,
			CACert:     cacert,
			CertChain:  certchain,
			PrivateKey: privatekey,
			EcdhCurves: "P256",
		}
		meshClusterConfig = newBasicTLSCluster(downstreamCluster, []string{serveraddr}, tlsConf)
		meshServerChain = newTLSFilterChain("upstreamFilter", meshproto, appproto, upstreamRouters, tlsConf)
	} else {
		meshClusterConfig = newBasicCluster(downstreamCluster, []string{serveraddr})
		meshServerChain = newFilterChain("upstreamFilter", meshproto, appproto, upstreamRouters)
	}
	cmconfig := config.ClusterManagerConfig{
		Clusters: []config.ClusterConfig{
			meshClusterConfig,
			newBasicCluster(upstreamCluster, hosts),
		},
	}
	serverChains := []config.FilterChain{meshServerChain}
	serverListener := newListener("upstreamListener", serveraddr, serverChains)
	return newMOSNConfig([]config.ListenerConfig{
		clientListener, serverListener,
	}, cmconfig)

}
