/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mosn

import (
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/psh686868/psh-mosn/pkg/api/v2"
	"github.com/psh686868/psh-mosn/pkg/config"
	"github.com/psh686868/psh-mosn/pkg/filter"
	"github.com/psh686868/psh-mosn/pkg/log"
	"github.com/psh686868/psh-mosn/pkg/protocol"
	"github.com/psh686868/psh-mosn/pkg/server"
	"github.com/psh686868/psh-mosn/pkg/server/config/proxy"
	"github.com/psh686868/psh-mosn/pkg/types"
	"github.com/psh686868/psh-mosn/pkg/upstream/cluster"
	"github.com/psh686868/psh-mosn/pkg/xds"
	"github.com/psh686868/psh-mosn/pkg/network"
)

// Mosn class which wrapper server
type Mosn struct {
	servers []server.Server
}

// NewMosn
// Create server from mosn config
func NewMosn(c *config.MOSNConfig) *Mosn {
	m := &Mosn{} // create Mosn

	// config server from config file
	mode := c.Mode()

	// 只从文件读取
	if mode == config.Xds {
		log.StartLogger.Fatalln("现在只做从文件读取文件")
	}
	if c.ClusterManager.Clusters == nil || len(c.ClusterManager.Clusters) == 0 {
		log.StartLogger.Fatalln("客户端失败")
	}

	countClues := len(c.ClusterManager.Clusters)

	if countClues == 0 {
		log.StartLogger.Fatalln("请配置 clusters")
	}

	// get inherit fds 慢慢消化
	inheritListeners := getInheritListeners()

	//get inherit fds
	//inheritListeners := getInheritListeners()

	// load filter
	cmf := &clusterManagerFilter{}

	//parse cluster all in one  并且进行心跳检查
	clusters, clusterMap := config.ParseClusterConfig(c.ClusterManager.Clusters)

	for _, serverConfig := range c.Servers {
		// 1. erver config prepare
		// server config
		sc := config.ParseServerConfig(&serverConfig)

		//create cluster manager

		// init default log
		server.InitDefaultLogger(sc)

		cm := cluster.NewClusterManager(nil, clusters, clusterMap, c.ClusterManager.AutoDiscovery, c.ClusterManager.RegistryUseHealthCheck)

		// init  server
		srv := server.NewServer(sc, cmf, cm)

		//add listener

		if serverConfig.Listeners == nil || len(serverConfig.Listeners) == 0 {
			log.StartLogger.Fatalln("no listenner found")
		}

		for _, listenerConfig := range serverConfig.Listeners {

			// parse ListenerConfig
			lc := config.ParseListenerConfig(&listenerConfig, inheritListeners)

			nfcf, downstreamProtocol := getNetworkFilter(&lc.FilterChains[0])

			// Note: as we use fasthttp and net/http2.0, the IO we created in mosn should be disabled
			// in the future, if we realize these two protocol by-self, this this hack method should be removed
			if downstreamProtocol == string(protocol.HTTP2) || downstreamProtocol == string(protocol.HTTP1) {
				lc.DisableConnIo = true
			}

			// network filters
			if lc.HandOffRestoredDestinationConnections {
				srv.AddListener(lc, nil, nil)
				continue
			}

			//stream filters
			sfcf := getStreamFilters(listenerConfig.StreamFilters)
			config.SetGlobalStreamFilter(sfcf)
			srv.AddListener(lc, nfcf, sfcf)
		}
		m.servers = append(m.servers, srv)
	}

	//parse service registry info
	config.ParseServiceRegistry(c.ServiceRegistry)

	//close legacy listeners
	for _, ln := range inheritListeners {
		if !ln.Remain {
			log.StartLogger.Println("close useless legacy listener:", ln.Addr)
			ln.InheritListener.Close()
		}
	}

	// set TransferTimeout
	network.TransferTimeout = server.GracefulTimeout
	// transfer old mosn connections
	go network.TransferServer(m.servers[0].Handler())

	return m
}

// Start mosn's server
func (m *Mosn) Start() {
	for _, srv := range m.servers {
		go srv.Start()
	}
}

// Close mosn's server
func (m *Mosn) Close() {
	for _, srv := range m.servers {
		srv.Close()
	}
}

// Start mosn project
// stap1. NewMosn
// step2. Start Mosn
func Start(c *config.MOSNConfig, serviceCluster string, serviceNode string) {
	log.StartLogger.Infof("start by config : %+v", c)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		// pprof server
		http.ListenAndServe("0.0.0.0:9090", nil)
	}()

	Mosn := NewMosn(c)
	Mosn.Start()
	////get xds config
	xdsClient := xds.Client{}
	xdsClient.Start(c, serviceCluster, serviceNode)
	//
	////todo: daemon running
	wg.Wait()
	xdsClient.Stop()
}

// getNetworkFilter
// Used to parse proxy from config
func getNetworkFilter(c *v2.FilterChain) (types.NetworkFilterChainFactory,string) {

	if len(c.Filters) != 1 || c.Filters[0].Name != v2.DEFAULT_NETWORK_FILTER {
		log.StartLogger.Fatalln("Currently, only Proxy Network Filter Needed!")
	}

	v2proxy := config.ParseProxyFilterJSON(&c.Filters[0])

	return &proxy.GenericProxyFilterConfigFactory{
		Proxy: v2proxy,
	},v2proxy.DownstreamProtocol
}

func getStreamFilters(configs []config.FilterConfig) []types.StreamFilterChainFactory {
	var factories []types.StreamFilterChainFactory

	for _, c := range configs {
		factories = append(factories, filter.CreateStreamFilterChainFactory(c.Type, c.Config))
	}
	return factories
}

type clusterManagerFilter struct {
	cccb types.ClusterConfigFactoryCb
	chcb types.ClusterHostFactoryCb
}

func (cmf *clusterManagerFilter) OnCreated(cccb types.ClusterConfigFactoryCb, chcb types.ClusterHostFactoryCb) {
	cmf.cccb = cccb
	cmf.chcb = chcb
}

func getInheritListeners() []*v2.ListenerConfig {
	if os.Getenv("_MOSN_GRACEFUL_RESTART") == "true" {
		count, _ := strconv.Atoi(os.Getenv("_MOSN_INHERIT_FD"))
		listeners := make([]*v2.ListenerConfig, count)

		log.StartLogger.Infof("received %d inherit fds", count)

		for idx := 0; idx < count; idx++ {
			//because passed listeners fd's index starts from 3
			fd := uintptr(3 + idx)
			file := os.NewFile(fd, "")
			fileListener, err := net.FileListener(file)
			if err != nil {
				log.StartLogger.Errorf("recover listener from fd %d failed: %s", fd, err)
				continue
			}
			if listener, ok := fileListener.(*net.TCPListener); ok {
				listeners[idx] = &v2.ListenerConfig{Addr: listener.Addr(), InheritListener: listener}
			} else {
				log.StartLogger.Errorf("listener recovered from fd %d is not a tcp listener", fd)
			}
		}
		return listeners
	}
	return nil
}
