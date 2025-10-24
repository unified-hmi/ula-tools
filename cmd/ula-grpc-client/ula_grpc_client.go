// SPDX-License-Identifier: Apache-2.0
/**
 * Copyright (c) 2024  Panasonic Automotive Systems, Co., Ltd.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
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
	"context"
	"flag"
	"fmt"
	"os"
	"time"
	"ula-tools/internal/ula-client/dwmapi"
	. "ula-tools/internal/ulog"
	"ula-tools/proto/grpc/dwm"
)

func printUsage() {
	usage := `
Usage: ula-grpc-client [options]
Options:
  -c      specify a dwm api command (default: DwmSetSystemLayout)
          DwmSetSystemLayout           no arguments
          DwmSetLayoutCommand          filePath
  -h      Show this message
`
	fmt.Println(usage)
}

func main() {
	var command string
	var showHelp bool
	flag.StringVar(&command, "c", "DwmSetSystemLayout", "Command to execute (e.g., DwmSetSystemLayout, etc.)")
	flag.BoolVar(&showHelp, "h", false, "Show this message")
	flag.Parse()
	args := flag.Args()

	if showHelp {
		printUsage()
		os.Exit(0)
	}

	ILog.SetOutput(os.Stderr)
	DLog.SetOutput(os.Stderr)

	conn, err := dwmapi.DwmClientInit()
	if err != nil {
		ELog.Printf("Error calling DwmClientInit: %v", err)
		return
	}
	defer conn.Close()
	client := dwm.NewDwmServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	switch command {
	case "DwmSetSystemLayout":
		err = dwmapi.DwmClientSetSystemLayout(client, ctx)
		if err != nil {
			ELog.Printf("Error calling SetSystemLayout: %v", err)
		}
	case "DwmSetLayoutCommand":
		if len(args) < 1 {
			ELog.Printf("DwmSetLayoutCommand requires an argument: filePath")
			os.Exit(1)
		}
		layoutCommandFilePath := args[0]
		err = dwmapi.DwmClientSetLayoutCommand(client, ctx, layoutCommandFilePath)
		if err != nil {
			ELog.Printf("Error calling SetLayoutCommand: %v", err)
		}
	default:
		err = dwmapi.DwmClientSetSystemLayout(client, ctx)
		if err != nil {
			ELog.Printf("Error calling SetSystemLayout: %v", err)
		}
	}
}
