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

package main

import (
	_ "flag"
	_ "net/http/pprof"
	"os"
	"time"

	_ "github.com/psh686868/psh-mosn/pkg/filter/stream/healthcheck/sofarpc"
	_ "github.com/psh686868/psh-mosn/pkg/network"
	_ "github.com/psh686868/psh-mosn/pkg/network/buffer"
	_ "github.com/psh686868/psh-mosn/pkg/protocol"
	_ "github.com/psh686868/psh-mosn/pkg/protocol/sofarpc/codec"
	_ "github.com/psh686868/psh-mosn/pkg/upstream/healthcheck"
	_ "github.com/psh686868/psh-mosn/pkg/xds"
	_"github.com/psh686868/psh-mosn/pkg/filter/stream/healthcheck/sofarpc"
	"github.com/urfave/cli"
)



func main() {
	app := cli.NewApp()
	app.Name = "mosn"
	app.Version = "0.0.1"
	app.Compiled = time.Now()
	app.Copyright = "(c) 2018 Ant Financial"
	app.Usage = "MOSN is modular observable smart netstub."

	//commands
	app.Commands = []cli.Command{
		cmdStart,
		cmdStop,
		cmdReload,
	}

	//action
	app.Action = func(c *cli.Context) error {
		cli.ShowAppHelp(c)

		c.App.Setup()
		return nil
	}

	// ignore error so we don't exit non-zero and break gfmrun README example tests
	_ = app.Run(os.Args)
}
