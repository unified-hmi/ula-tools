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
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"time"
	"ula-tools/internal/ula"
	. "ula-tools/internal/ulog"
	"ula-tools/proto/grpc/dwm"
)

func getTargetAddr(vscrnDef *ula.VScrnDef) string {
	var ipAddr string
	var port int
	for _, node := range vscrnDef.Nodes {
		for _, frameworkNode := range vscrnDef.DistributedWindowSystem.FrameworkNode {
			if node.NodeId != frameworkNode.NodeId {
				continue
			}
			if node.NodeId == vscrnDef.DistributedWindowSystem.ULAClientManager.NodeId {
				ipAddr = node.Ip
				port = vscrnDef.DistributedWindowSystem.ULAClientManager.Port
				return ipAddr + ":" + strconv.Itoa(port)
			}
		}
	}
	return DEFAULT_GRPC_SERVER_ADDR
}

func getMyNodeId(vscrnDef *ula.VScrnDef) (int, error) {
	keyHostName, err := os.Hostname()
	if err != nil {
		WLog.Println("os.Hostname error : ", err)
		return -1, err
	}
	nodeId, err := vscrnDef.GetNodeIdByHostName(keyHostName)
	if err != nil {
		WLog.Println("GetNodeIdByHostName error : ", err)
		return -1, err
	}
	return nodeId, nil
}

func DwmClientInit() (*grpc.ClientConn, error) {
	vscrnDef, err := ula.ReadVScrnDef()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("ReadVScrnDef error : %s\n", err))
	}
	targetAddr := getTargetAddr(vscrnDef)
	targetIP, _, err := net.SplitHostPort(targetAddr)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("SplitHostPort error : %s\n", err))
	}
	ipAddrs, err := ula.GetIpv4AddrsOfAllInterfaces()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("GetIpv4AddrsOfAllInterfaces error : %s\n", err))
	}
	nodeId, err := getMyNodeId(vscrnDef)
	var candidateIPs []string
	if err == nil {
		ipAddr, err := vscrnDef.GetIpAddrByNodeIdAndIpCandidateList(ipAddrs, nodeId)
		if err != nil {
			WLog.Println("GetIpAddrByNodeIdAndIpCandidateList error : ", err)
			candidateIPs = ipAddrs
		} else {
			candidateIPs = []string{ipAddr}
		}
	} else {
		candidateIPs = ipAddrs
	}

	if targetIP != "127.0.0.1" {
		filteredIPs := []string{}
		for _, ip := range candidateIPs {
			ipParsed := net.ParseIP(ip)
			if ipParsed != nil && ipParsed.IsLoopback() {
				continue
			}
			filteredIPs = append(filteredIPs, ip)
		}
		candidateIPs = filteredIPs
	}
	var conn *grpc.ClientConn
	var dialErr error

	for _, ip := range candidateIPs {
		parsedIP := net.ParseIP(ip)
		if parsedIP == nil {
			WLog.Printf("Invalid IP: %s, skipping", ip)
			continue
		}
		ILog.Printf("Trying to connect from IP: %s to target: %s", parsedIP, targetAddr)
		dialer := &net.Dialer{
			LocalAddr: &net.TCPAddr{
				IP: parsedIP,
			},
			Timeout: 1 * time.Second,
		}

		conn, dialErr = grpc.DialContext(
			context.Background(),
			targetAddr,
			grpc.WithInsecure(),
			grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
				return dialer.DialContext(ctx, "tcp", addr)
			}),
		)

		if dialErr == nil {
			ILog.Printf("Successfully connected from IP: %s", parsedIP)
			return conn, nil
		} else {
			WLog.Printf("Connection from IP %s failed: %v", parsedIP, dialErr)
		}
	}
	return nil, dialErr
}

func DwmClientSetSystemLayout(client dwm.DwmServiceClient, ctx context.Context) error {
	resp, err := client.DwmSetSystemLayout(ctx, &dwm.Empty{})
	if err != nil {
		return err
	}
	ILog.Println("DwmClientSetSystemLayout DwmSetSystemLayout:", resp.GetStatus())
	return nil
}

func DwmClientSetLayoutCommand(client dwm.DwmServiceClient, ctx context.Context, layoutCommandFilePath string) error {
	f, err := os.Open(layoutCommandFilePath)
	if err != nil {
		return err
	}
	layoutCommandBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	commReq := &dwm.SetLayoutCommandRequest{
		LayoutCommand: string(layoutCommandBytes),
	}
	resp, err := client.DwmSetLayoutCommand(ctx, commReq)
	if err != nil {
		return err
	}
	ILog.Println("DwmSetLayoutCommand response:", resp.GetStatus())
	return nil
}
