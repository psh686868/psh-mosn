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

package filter

import (
	"github.com/psh686868/psh-mosn/pkg/log"
	"github.com/psh686868/psh-mosn/pkg/types"
)

var creatorFactory map[string]StreamFilterFactoryCreator

func init() {
	creatorFactory = make(map[string]StreamFilterFactoryCreator)
}

// Register filter and its creator
func Register(filterType string, creator StreamFilterFactoryCreator) {
	creatorFactory[filterType] = creator
}

// CreateStreamFilterChainFactory
// Create Stream Filter Chain Factory according to filterType
func CreateStreamFilterChainFactory(filterType string, config map[string]interface{}) types.StreamFilterChainFactory {

	if cf, ok := creatorFactory[filterType]; ok {
		sfcf, err := cf(config)

		if err != nil {
			log.StartLogger.Fatalln("create stream filter chain factory failed: ", err)
		}

		return sfcf
	}

	log.StartLogger.Fatalln("unsupport stream filter type: ", filterType)
	return nil
}
