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

package dwmapi

import (
	"context"
	"google.golang.org/grpc"
	"net"
	"os"
	"runtime"
	"strconv"
	"sync"
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-client/readclusterapp"
	"ula-tools/internal/ula-client/ulacommgen"
	"ula-tools/internal/ula-client/ulamulticonn"
	"ula-tools/internal/ula-client/ulavscreen"
	. "ula-tools/internal/ulog"
	"ula-tools/proto/grpc/dwm"
)

type clientWg struct {
	requestWg *sync.WaitGroup
	clientId  string
}

var (
	requestWgsMutex sync.Mutex
	requestWgs      map[string]*clientWg = make(map[string]*clientWg)
)

func logFunc() {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		WLog.Println("Could not get caller info")
		return
	}
	funcName := runtime.FuncForPC(pc).Name()
	DLog.Println("Function:", funcName)
}

type server struct {
	dwm.UnimplementedDwmServiceServer
}

func (s *server) DwmSetSystemLayout(ctx context.Context, req *dwm.Empty) (*dwm.Response, error) {
	logFunc()

	calayoutTree, err := readclusterapp.ReadCALayoutTreeFromCfg()
	if err != nil {
		return &dwm.Response{Status: "Failed to DwmSetSystemLayout"}, err
	}

	var layoutComm string
	layoutComm, err = ulacommgen.GenerateUlaCommInitialVscreen(calayoutTree)
	if err != nil {
		return &dwm.Response{Status: "Failed to DwmSetSystemLayout"}, err
	}

	err = ulamulticonn.UlaMulCon.SendLayoutCommand(layoutComm)
	if err != nil {
		return &dwm.Response{Status: "Failed to DwmSetSystemLayout"}, err
	}
	return &dwm.Response{Status: "System layout set successfully"}, nil
}

func (s *server) DwmSetLayoutCommand(ctx context.Context, req *dwm.SetLayoutCommandRequest) (*dwm.Response, error) {
	logFunc()
	layoutCommand := req.GetLayoutCommand()
	err := ulamulticonn.UlaMulCon.SendLayoutCommand(layoutCommand)
	if err != nil {
		ELog.Println(err)
		return &dwm.Response{Status: "Failed to DwmSetLayoutCommand"}, err
	}
	return &dwm.Response{Status: "Set layout command successfully"}, nil
}

func getServerAddr(vscrnDef *ula.VScrnDef) string {
	keyHostName, err := os.Hostname()
	if err != nil {
		WLog.Println("os.Hostname error : ", err)
		return DEFAULT_GRPC_SERVER_ADDR
	}
	nodeId, err := vscrnDef.GetNodeIdByHostName(keyHostName)
	if err != nil {
		WLog.Println("GetNodeIdByHostName error : ", err)
		return DEFAULT_GRPC_SERVER_ADDR
	}
	ipAddrs, err := ula.GetIpv4AddrsOfAllInterfaces()
	if err != nil {
		WLog.Println("GetIpv4AddrsOfAllInterfaces error : ", err)
		return DEFAULT_GRPC_SERVER_ADDR
	}
	listenIp, err := vscrnDef.GetIpAddrByNodeIdAndIpCandidateList(ipAddrs, nodeId)
	if err != nil {
		WLog.Println("GetIpAddrByNodeIdAndIpCandidateList error : ", err)
		return DEFAULT_GRPC_SERVER_ADDR
	}
	if vscrnDef.DistributedWindowSystem.ULAClientManager.Port == 0 {
		WLog.Println("ULAClientManager is not set port")
		return DEFAULT_GRPC_SERVER_ADDR
	}
	serverAddr := listenIp + ":" + strconv.Itoa(vscrnDef.DistributedWindowSystem.ULAClientManager.Port)
	return serverAddr
}

func DwmServerInit(vsdPath string) error {
	vscrnDef, err := ula.ReadVScrnDef(vsdPath)
	if err != nil {
		ELog.Printf("Failed to Read VirtualScreen: %s\n", err)
		return err
	}
	ulavscreen.VScreen, err = ulavscreen.NewVirtualScreen(vscrnDef)
	if err != nil {
		ELog.Printf("Failed to Create VirtualScreen: %s\n", err)
		return err
	}
	force := ula.GetEnvBool("ULA_FORCE", false)
	err = ulamulticonn.UlaConnectionInit(force)
	if err != nil {
		ELog.Println("Failed to Init Connection: %s\n", err)
		return err
	}

	serverAddr := getServerAddr(vscrnDef)

	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		ELog.Printf("Failed to listen: %v", err)
		return err
	}

	s := grpc.NewServer()
	dwm.RegisterDwmServiceServer(s, &server{})
	ILog.Println("Server listening on ", serverAddr)
	if err := s.Serve(listener); err != nil {
		ELog.Printf("Failed to serve: %v", err)
		return err
	}
	return nil
}
