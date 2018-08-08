package main

import (
	"flag"
	"fmt"
	"net"
	"time"

	"github.com/psh686868/psh-mosn/pkg/log"
	"github.com/psh686868/psh-mosn/pkg/network"
	"github.com/psh686868/psh-mosn/pkg/protocol"
	"github.com/psh686868/psh-mosn/pkg/protocol/serialize"
	"github.com/psh686868/psh-mosn/pkg/protocol/sofarpc"
	_ "github.com/psh686868/psh-mosn/pkg/protocol/sofarpc/codec"
	"github.com/psh686868/psh-mosn/pkg/stream"
	_ "github.com/psh686868/psh-mosn/pkg/stream/sofarpc"
	"github.com/psh686868/psh-mosn/pkg/types"
)

type Client struct {
	Codec stream.CodecClient
	conn  types.ClientConnection
	Id    uint32
}

func NewClient(addr string) *Client {
	c := &Client{}
	stopChan := make(chan struct{})
	remoteAddr, _ := net.ResolveTCPAddr("tcp", addr)
	conn := network.NewClientConnection(nil, nil, remoteAddr, stopChan, log.DefaultLogger)
	if err := conn.Connect(true); err != nil {
		fmt.Println(err)
		return nil
	}
	c.Codec = stream.NewCodecClient(nil, protocol.SofaRPC, conn, nil)
	c.conn = conn
	return c
}

func (c *Client) OnReceiveData(data types.IoBuffer, endStream bool)  {}
func (c *Client) OnReceiveTrailers(trailers map[string]string)       {}
func (c *Client) OnDecodeError(err error, headers map[string]string) {}
func (c *Client) OnReceiveHeaders(headers map[string]string, endStream bool) {
	fmt.Printf("[RPC Client] Receive Data:")
	if streamID, ok := headers[sofarpc.SofaPropertyHeader(sofarpc.HeaderReqID)]; ok {
		fmt.Println(streamID)
	}
}

func (c *Client) Request() {
	c.Id++
	streamID := sofarpc.StreamIDConvert(c.Id)
	requestEncoder := c.Codec.NewStream(streamID, c)
	headers := buildBoltV1Request(c.Id)
	requestEncoder.AppendHeaders(headers, true)
}

func buildBoltV1Request(requestID uint32) *sofarpc.BoltRequestCommand {
	request := &sofarpc.BoltRequestCommand{
		Protocol: sofarpc.PROTOCOL_CODE_V1,
		CmdType:  sofarpc.REQUEST,
		CmdCode:  sofarpc.RPC_REQUEST,
		Version:  1,
		ReqID:    requestID,
		CodecPro: sofarpc.HESSIAN_SERIALIZE, //todo: read default codec from config
		Timeout:  -1,
	}

	headers := map[string]string{"service": "testSofa"} // used for sofa routing

	if headerBytes, err := serialize.Instance.Serialize(headers); err != nil {
		panic("serialize headers error")
	} else {
		request.HeaderMap = headerBytes
		request.HeaderLen = int16(len(headerBytes))
	}

	return request
}

func main() {
	log.InitDefaultLogger("", log.DEBUG)
	t := flag.Bool("t", false, "-t")
	flag.Parse()
	if client := NewClient("127.0.0.1:2045"); client != nil {
		for {
			client.Request()
			time.Sleep(200 * time.Millisecond)
			if !*t {
				time.Sleep(3 * time.Second)
				return
			}
		}
	}
}
