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

package stream

import (
	"context"

	"github.com/psh686868/psh-mosn/pkg/types"
)

var streamFactories map[types.Protocol]ProtocolStreamFactory

func init() {
	streamFactories = make(map[types.Protocol]ProtocolStreamFactory)
}

func Register(prot types.Protocol, factory ProtocolStreamFactory) {
	streamFactories[prot] = factory
}

func CreateServerStreamConnection(context context.Context, prot types.Protocol, connection types.Connection,
	callbacks types.ServerStreamConnectionEventListener) types.ServerStreamConnection {

	if ssc, ok := streamFactories[prot]; ok {
		return ssc.CreateServerStream(context, connection, callbacks)
	}

	return nil
}
