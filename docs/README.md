# MOSN Project

[![Build Status](https://travis-ci.org/psh686868/psh-mosn.svg?branch=master)](https://travis-ci.org/psh686868/psh-mosn)
[![codecov](https://codecov.io/gh/psh686868/psh-mosn/branch/master/graph/badge.svg)](https://codecov.io/gh/psh686868/psh-mosn)
[![Go Report Card](https://goreportcard.com/badge/github.com/psh686868/psh-mosn)](https://goreportcard.com/report/github.com/psh686868/psh-mosn)
![license](https://img.shields.io/badge/license-Apache--2.0-green.svg)

MOSN, the short name of Modular Observable Smart Network, is a powerful proxy acting as Service Mesh's data plane like Envoy but written in golang.
MOSN supports Envoy and Istio's APIs and can be integrated with Istio, and we use MOSN instead of Envoy in SOFAMesh.
The initial version of MOSN was jointly contributed by Ant Financial and UC Business Unit of Alibaba, and we look forward to the community to participate in the
follow-up development and build an open source boutique project together.

## Core Competence

+ Integrated with Istio
    + Integrated with Istio version 1.0 and V4 API and can be run based on full dynamic resource configuration
+ Packet Forwarding
    + Self-contained network server
    + Support TCP proxy
    + Support TProxy mode
+ Multi-protocol support
    + HTTP/1.1，HTTP/2.0
    + SOFARPC
    + Dubbo
    + HSF(ongoing)
+ Routing support
    + Routing in form of virtual host
    + Routing with headers/url/prefix
    + Routing with host metadata in form of subset
    + Routing retry
+ Cluster Management & Load Balance support
    + Connection pool
    + Circuit Breaker
    + Health Checker
    + Random/RR LoadBalance
    + Subset LoadBalance with host's metadata
+ Observable
    + Network layer data
    + Protocol data
+ TLS support
    + HTTP/1.x on TLS
    + HTTP/2 on TLS
    + SOFARPC on TLS
+ Process management
    + Smooth reload
    + Smooth upgrade
+ Scalable ability
    + Support self-defined private protocol
    + Scalable Network/IO ，stream layer

## Architecture design
* [Reference](docs/design/README.md)

## Quic Start
* [Details](docs/develop/quickstart.md)
   
## Docs
* [More here](http://www.sofastack.tech/sofa-mesh/docs/mosn-README)

## Community
* [Issues](https://github.com/psh686868/psh-mosn/issues)

## Contribution
+ [How to contribute the code](docs/CONTRIBUTING.md)
+ MOSN is still in its infancy with many capabilities need to be completed, and many bugs to be fixed.
  So we welcome everyone to participate in and commit code in following but not limited aspect：
   + Completing core routing function
   + Outlier detection
   + Tracing support
   + HTTP/1.x, HTTP/2.0 performance optimization
   + Flow control
   
## Thanks
Thanks to Google, IBM, Lyft for creating the Envoy and Istio system, so that MOSN has a very good reference and we can
quickly land our own ideas.
