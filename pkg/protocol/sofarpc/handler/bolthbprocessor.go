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

package handler

import (
	"context"

	"github.com/psh686868/psh-mosn/pkg/log"
	"github.com/psh686868/psh-mosn/pkg/protocol"
	"github.com/psh686868/psh-mosn/pkg/protocol/sofarpc"
	"github.com/psh686868/psh-mosn/pkg/types"
)

type boltHbProcessor struct {
}

func (b *boltHbProcessor) Process(context context.Context, msg interface{}, filter interface{}) {
	logger := log.ByContext(context)

	if cmd, ok := msg.(*sofarpc.BoltRequestCommand); ok {
		deserializeRequestAllFields(context, cmd)
		reqID := protocol.StreamIDConv(cmd.ReqID)

		//Heartbeat message only has request header
		if filter, ok := filter.(types.DecodeFilter); ok {
			if cmd.RequestHeader != nil {
				status := filter.OnDecodeHeader(reqID, cmd.RequestHeader)
				//		logger.Debugf("Process Heartbeat Request Msg")

				if status == types.StopIteration {
					return
				}
			}
		}
	} else if cmd, ok := msg.(*sofarpc.BoltResponseCommand); ok {
		deserializeResponseAllFields(context, cmd)
		reqID := protocol.StreamIDConv(cmd.ReqID)
		//logger := log.ByContext(context)

		//Heartbeat message only has request header
		if filter, ok := filter.(types.DecodeFilter); ok {
			if cmd.ResponseHeader != nil {
				status := filter.OnDecodeHeader(reqID, cmd.ResponseHeader)
				//	logger.Debugf("Process Heartbeat Response Msg")

				if status == types.StopIteration {
					return
				}
			}
		}
	} else {
		logger.Errorf("decode heart beat error\n")
	}
}
