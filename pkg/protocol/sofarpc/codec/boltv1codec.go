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

package codec

import (
	"context"
	"encoding/binary"
	"errors"
	"reflect"
	"time"

	"github.com/psh686868/psh-mosn/pkg/log"
	"github.com/psh686868/psh-mosn/pkg/network/buffer"
	"github.com/psh686868/psh-mosn/pkg/protocol/serialize"
	"github.com/psh686868/psh-mosn/pkg/protocol/sofarpc"
	"github.com/psh686868/psh-mosn/pkg/types"
	"fmt"
)

// BoltV1PropertyHeaders map the cmdkey and its data type
var (
	BoltV1PropertyHeaders = make(map[string]reflect.Kind, 11)
	defaultTmpBufferSize  = 1 << 6
)

func init() {
	BoltV1PropertyHeaders[sofarpc.HeaderProtocolCode] = reflect.Uint8
	BoltV1PropertyHeaders[sofarpc.HeaderCmdType] = reflect.Uint8
	BoltV1PropertyHeaders[sofarpc.HeaderCmdCode] = reflect.Int16
	BoltV1PropertyHeaders[sofarpc.HeaderVersion] = reflect.Uint8
	BoltV1PropertyHeaders[sofarpc.HeaderReqID] = reflect.Uint32
	BoltV1PropertyHeaders[sofarpc.HeaderCodec] = reflect.Uint8
	BoltV1PropertyHeaders[sofarpc.HeaderClassLen] = reflect.Int16
	BoltV1PropertyHeaders[sofarpc.HeaderHeaderLen] = reflect.Int16
	BoltV1PropertyHeaders[sofarpc.HeaderContentLen] = reflect.Int
	BoltV1PropertyHeaders[sofarpc.HeaderTimeout] = reflect.Int
	BoltV1PropertyHeaders[sofarpc.HeaderRespStatus] = reflect.Int16
	BoltV1PropertyHeaders[sofarpc.HeaderRespTimeMills] = reflect.Int64
}

// types.Encoder & types.Decoder
type boltV1Codec struct{}

func (c *boltV1Codec) EncodeHeaders(context context.Context, headers interface{}) (types.IoBuffer, error) {
	if headerMap, ok := headers.(map[string]string); ok {

		cmd := c.mapToCmd(headerMap)
		return c.encodeHeaders(context, cmd)
	}

	return c.encodeHeaders(context, headers)
}

func (c *boltV1Codec) encodeHeaders(context context.Context, headers interface{}) (types.IoBuffer, error) {
	switch headers.(type) {
	case *sofarpc.BoltRequestCommand:
		return c.encodeRequestCommand(context, headers.(*sofarpc.BoltRequestCommand))
	case *sofarpc.BoltResponseCommand:
		return c.encodeResponseCommand(context, headers.(*sofarpc.BoltResponseCommand))
	default:

		errMsg := sofarpc.InvalidCommandType
		err := errors.New(errMsg)
		log.ByContext(context).Errorf("boltV1" + errMsg)
		return nil, err
	}
}

func (c *boltV1Codec) EncodeData(context context.Context, data types.IoBuffer) types.IoBuffer {
	return data
}

func (c *boltV1Codec) EncodeTrailers(context context.Context, trailers map[string]string) types.IoBuffer {
	return nil
}

func (c *boltV1Codec) encodeRequestCommand(context context.Context, cmd *sofarpc.BoltRequestCommand) (types.IoBuffer, error) {
	result := c.doEncodeRequestCommand(context, cmd)
	return buffer.NewIoBufferBytes(result), nil
}

func (c *boltV1Codec) encodeResponseCommand(context context.Context, cmd *sofarpc.BoltResponseCommand) (types.IoBuffer, error) {
	result := c.doEncodeResponseCommand(context, cmd)
	return buffer.NewIoBufferBytes(result), nil
}

func (c *boltV1Codec) doEncodeRequestCommand(context context.Context, cmd *sofarpc.BoltRequestCommand) []byte {
	offset := 0
	// todo: reuse bytes @boqin
	data := make([]byte, 22, defaultTmpBufferSize)

	data[offset] = cmd.Protocol
	offset++

	data[offset] = cmd.CmdType
	offset++

	binary.BigEndian.PutUint16(data[offset:], uint16(cmd.CmdCode))
	offset += 2

	data[offset] = cmd.Version
	offset++

	binary.BigEndian.PutUint32(data[offset:], uint32(cmd.ReqID))
	offset += 4

	data[offset] = cmd.CodecPro
	offset++

	binary.BigEndian.PutUint32(data[offset:], uint32(cmd.Timeout))
	offset += 4

	binary.BigEndian.PutUint16(data[offset:], uint16(cmd.ClassLen))
	offset += 2

	binary.BigEndian.PutUint16(data[offset:], uint16(len(cmd.HeaderMap)))
	offset += 2

	binary.BigEndian.PutUint32(data[offset:], uint32(cmd.ContentLen))
	offset += 4

	if cmd.ClassLen > 0 {
		data = append(data, cmd.ClassName...)
	}

	if len(cmd.HeaderMap) > 0 {
		data = append(data, cmd.HeaderMap...)
	}

	return data
}

func (c *boltV1Codec) doEncodeResponseCommand(context context.Context, cmd *sofarpc.BoltResponseCommand) []byte {
	offset := 0
	// todo: reuse bytes @boqin
	data := make([]byte, 20, defaultTmpBufferSize)

	data[offset] = cmd.Protocol
	offset++

	data[offset] = cmd.CmdType
	offset++

	binary.BigEndian.PutUint16(data[offset:], uint16(cmd.CmdCode))
	offset += 2

	if cmd.CmdCode == sofarpc.HEARTBEAT {
		log.ByContext(context).Debugf("Build HeartBeat Response")
	}

	data[offset] = cmd.Version
	offset++

	binary.BigEndian.PutUint32(data[offset:], uint32(cmd.ReqID))
	offset += 4

	data[offset] = cmd.CodecPro
	offset++

	binary.BigEndian.PutUint16(data[offset:], uint16(cmd.ResponseStatus))
	offset += 2

	binary.BigEndian.PutUint16(data[offset:], uint16(cmd.ClassLen))
	offset += 2

	binary.BigEndian.PutUint16(data[offset:], uint16(len(cmd.HeaderMap)))
	offset += 2

	binary.BigEndian.PutUint32(data[offset:], uint32(cmd.ContentLen))
	offset += 4

	if cmd.ClassLen > 0 {
		data = append(data, cmd.ClassName...)
	}

	if len(cmd.HeaderMap) > 0 {
		data = append(data, cmd.HeaderMap...)
	}

	return data
}

func (c *boltV1Codec) mapToCmd(headers map[string]string) interface{} {
	if len(headers) < 10 {
		return nil
	}

	protocolCode := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderProtocolCode)
	cmdType := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderCmdType)
	cmdCode := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderCmdCode)
	version := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderVersion)
	requestID := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderReqID)
	codec := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderCodec)
	classLength := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderClassLen)
	headerLength := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderHeaderLen)
	contentLength := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderContentLen)

	//class
	className := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, "classname")
	class, _ := serialize.Instance.Serialize(className)

	//RPC Request
	if cmdCode == sofarpc.RPC_REQUEST {
		timeout := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderTimeout)

		//serialize header
		header, _ := serialize.Instance.Serialize(headers)

		request := sofarpc.BoltRequestCommand{
			protocolCode.(byte),
			cmdType.(byte),
			cmdCode.(int16),
			version.(byte),
			requestID.(uint32),
			codec.(byte),
			timeout.(int),
			classLength.(int16),
			//int16(len(class)),
			headerLength.(int16),
			//int16(len(header)),
			contentLength.(int),
			class,
			header,
			nil,
			nil,
			nil,
		}

		return &request
	} else if cmdCode == sofarpc.RPC_RESPONSE || cmdCode == sofarpc.HEARTBEAT {
		//todo : review
		responseStatus := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderRespStatus)
		responseTime := sofarpc.GetPropertyValue(BoltV1PropertyHeaders, headers, sofarpc.HeaderRespTimeMills)

		//serialize header
		header, _ := serialize.Instance.Serialize(headers)
		response := sofarpc.BoltResponseCommand{
			protocolCode.(byte),
			cmdType.(byte),
			cmdCode.(int16),
			version.(byte),
			requestID.(uint32),
			codec.(byte),
			responseStatus.(int16),
			classLength.(int16),
			headerLength.(int16),
			contentLength.(int),
			class,
			header,
			nil,
			nil,
			responseTime.(int64),
			nil,
		}

		return &response
	}

	return nil
}

func (c *boltV1Codec) Decode(context context.Context, data types.IoBuffer) (interface{},error) {
	readableBytes := data.Len()
	read := 0
	var cmd interface{}
	logger := log.ByContext(context)

	if readableBytes >= sofarpc.LESS_LEN_V1 {
		bytes := data.Bytes()
		dataType := bytes[1]

		//1. request
		if dataType == sofarpc.REQUEST || dataType == sofarpc.REQUEST_ONEWAY {
			if readableBytes >= sofarpc.REQUEST_HEADER_LEN_V1 {

				cmdCode := binary.BigEndian.Uint16(bytes[2:4])

				//if cmdCode == uint16(sofarpc.HEARTBEAT) {
				//	logger.Debugf("BoltV1 DECODE Request: Get Bolt HB Msg")
				//}
				ver2 := bytes[4]
				requestID := binary.BigEndian.Uint32(bytes[5:9])
				codec := bytes[9]
				timeout := binary.BigEndian.Uint32(bytes[10:14])
				classLen := binary.BigEndian.Uint16(bytes[14:16])
				headerLen := binary.BigEndian.Uint16(bytes[16:18])
				contentLen := binary.BigEndian.Uint32(bytes[18:22])

				read = sofarpc.REQUEST_HEADER_LEN_V1
				var class, header, content []byte

				if readableBytes >= read+int(classLen)+int(headerLen)+int(contentLen) {
					if classLen > 0 {
						class = bytes[read : read+int(classLen)]
						read += int(classLen)
					}
					if headerLen > 0 {
						header = bytes[read : read+int(headerLen)]
						read += int(headerLen)
					}
					if contentLen > 0 {
						content = bytes[read : read+int(contentLen)]
						read += int(contentLen)
					}

					data.Drain(read)

				} else { // not enough data

					logger.Debugf("BoltV1 DECODE Request, no enough data for fully decode")
					return cmd, nil
				}

				request := sofarpc.BoltRequestCommand{

					sofarpc.PROTOCOL_CODE_V1,
					dataType,
					int16(cmdCode),
					ver2,
					requestID,
					codec,
					int(timeout),
					int16(classLen),
					int16(headerLen),
					int(contentLen),
					class,
					header,
					content,
					nil,
					nil,
				}
				logger.Debugf("BoltV1 DECODE REQUEST, Protocol = %d, CmdType = %d, CmdCode = %d, ReqID = %d",
					request.Protocol, request.CmdType, request.CmdCode, request.ReqID)
				cmd = &request
			}
		} else if dataType == sofarpc.RESPONSE{
			//2. response
			if readableBytes >= sofarpc.RESPONSE_HEADER_LEN_V1 {

				cmdCode := binary.BigEndian.Uint16(bytes[2:4])
				ver2 := bytes[4]
				requestID := binary.BigEndian.Uint32(bytes[5:9])
				codec := bytes[9]
				status := binary.BigEndian.Uint16(bytes[10:12])
				classLen := binary.BigEndian.Uint16(bytes[12:14])
				headerLen := binary.BigEndian.Uint16(bytes[14:16])
				contentLen := binary.BigEndian.Uint32(bytes[16:20])

				read = sofarpc.RESPONSE_HEADER_LEN_V1
				var class, header, content []byte

				if readableBytes >= read+int(classLen)+int(headerLen)+int(contentLen) {
					if classLen > 0 {
						class = bytes[read : read+int(classLen)]
						read += int(classLen)
					}
					if headerLen > 0 {
						header = bytes[read : read+int(headerLen)]
						read += int(headerLen)
					}
					if contentLen > 0 {
						content = bytes[read : read+int(contentLen)]
						read += int(contentLen)
					}

					data.Drain(read)
				} else {
					// not enough data
					logger.Debugf("BoltV1 DECODE RESPONSE: no enough data for fully decode")

					return cmd, nil
				}

				response := sofarpc.BoltResponseCommand{
					sofarpc.PROTOCOL_CODE_V1,
					dataType,
					int16(cmdCode),
					ver2,
					requestID,
					codec,
					int16(status),
					int16(classLen),
					int16(headerLen),
					int(contentLen),
					class,
					header,
					content,
					nil,
					time.Now().UnixNano() / int64(time.Millisecond),
					nil,
				}

				if cmdCode == uint16(sofarpc.HEARTBEAT) {
					//logger.Debugf("BoltV1 DECODE RESPONSE: Get Bolt HB Msg")
				}
				logger.Debugf("BoltV1 DECODE RESPONSE,RespStatus = %d, Protocol = %d, CmdType = %d, CmdCode = %d, ReqID = %d",
					response.ResponseStatus, response.Protocol, response.CmdType, response.CmdCode, response.ReqID)
				cmd = &response
			}
		} else {
			// 3. unknown type error
			return nil, fmt.Errorf("Decode Error, type = %s, value = %d", sofarpc.UnKnownReqtype, dataType)
		}
	}

	return cmd,nil
}
