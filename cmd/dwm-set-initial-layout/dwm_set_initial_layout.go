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
	"flag"
	"os"
	"ula-tools/internal/ula-client/dwmapi"
	. "ula-tools/internal/ulog"
)

func setEnv(env string, value string) {
	err := os.Setenv(env, value)
	if err != nil {
		ELog.Println("Error setting environment variable:", err)
		return
	}
}

func main() {
	var (
		dwmpath string
		vsdpath string
	)

	flag.StringVar(&dwmpath, "d", "/var/local/uhmi-app/dwm", "path for dwm directory")
	flag.StringVar(&vsdpath, "v", "/etc/uhmi-framework/virtual-screen-def.json", "path for virtual-screen-def.json file")
	flag.Parse()

	setEnv("DWMPATH", dwmpath)
	setEnv("VSDPATH", vsdpath)

	dwmapi.DwmInit()
	dwmapi.DwmSetSystemLayout()
}
