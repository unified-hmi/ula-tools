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

import "C"
import (
	"context"
	"os"
	"ula-tools/internal/ula-client/dwmapi"
	. "ula-tools/internal/ulog"
	"ula-tools/proto/grpc/dwm"
)

var dwmClient dwm.DwmServiceClient
var dwmCtx context.Context

func init() {
	ILog.SetOutput(os.Stderr)
	conn, err := dwmapi.DwmClientInit()
	if err != nil {
		ELog.Println(err)
		return
	}
	dwmClient = dwm.NewDwmServiceClient(conn)
	dwmCtx = context.Background()
}

//export dwm_set_system_layout
func dwm_set_system_layout() C.int {
	err := dwmapi.DwmClientSetSystemLayout(dwmClient, dwmCtx)
	if err != nil {
		ELog.Printf("Error calling SetSystemLayout: %v", err)
		return -1
	}
	return 0
}

//export dwm_set_layout_command
func dwm_set_layout_command(filePathChar *C.char) C.int {
	layoutCommandFilePath := C.GoString(filePathChar)
	err := dwmapi.DwmClientSetLayoutCommand(dwmClient, dwmCtx, layoutCommandFilePath)
	if err != nil {
		ELog.Printf("Error calling SetLayoutCommand: %v", err)
		return -1
	}
	return 0
}

func main() {}
