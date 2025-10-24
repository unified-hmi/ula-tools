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

package ulamulticonn

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"strconv"
	"sync"
	"syscall"
	"time"
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-client/ulavscreen"
	. "ula-tools/internal/ulog"
)

var UlaMulCon *UlaMultiConnector

var MAGIC_CODE []byte = []byte{0x55, 0x4C, 0x41, 0x30} // 'ULA0' ascii code

var Mutex struct {
	sync.Mutex
}

type TargetNodeAddr struct {
	NodeId     int
	TargetAddr string
}

type UlaMultiConnector struct {
	targetNodeAddrs []TargetNodeAddr
	sendChans       []chan string
	respChans       []chan UlaCommandResponse
	force           bool
}

type UlaCommandResponse struct {
	Type   string
	Result int
}

type DistribNode struct {
	NodeId int
	Ip     string
	Port   int
}

func newUlaMultiConn(force bool, vsdPath ...string) (*UlaMultiConnector, error) {

	dNodes, err := getDistribNodes(vsdPath...)
	if err != nil {
		return nil, err
	}

	targetNum := 0
	var targets []TargetNodeAddr
	for _, d := range dNodes {
		targetAddr := d.Ip + ":" + strconv.Itoa(d.Port)
		targets = append(targets, TargetNodeAddr{
			NodeId:     d.NodeId,
			TargetAddr: targetAddr,
		})
		targetNum++
	}

	if targetNum == 0 {
		return nil, errors.New("targetNode is not set correctly. Please check your virtual-screen-def.json file")
	}

	if UlaMulCon != nil && reflect.DeepEqual(UlaMulCon.targetNodeAddrs, targets) {
		WLog.Println("UlaMulCon has already initialized")
		return nil, nil
	}

	sendChans := make([]chan string, len(targets))
	respChans := make([]chan UlaCommandResponse, len(targets))
	for i := range targets {
		sendChans[i] = nil
		respChans[i] = nil
	}

	var ulaMulCon *UlaMultiConnector
	ulaMulCon = &UlaMultiConnector{
		targetNodeAddrs: targets,
		sendChans:       sendChans,
		respChans:       respChans,
		force:           force,
	}
	return ulaMulCon, nil
}

func retryConnectTarget(sockChan chan net.Conn, addr string) {
	for {
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			sockChan <- conn
			break
		}
		DLog.Printf("Dial failed connect : %s retry\n", err)
		time.Sleep(10 * time.Millisecond)
	}
}

func connectTarget(addr string, timeout time.Duration) (net.Conn, error) {
	conn, err := net.Dial("tcp", addr)
	if err == nil {
		ILog.Println("Dial connected to ", addr)
		return conn, nil
	}

	if timeout > 0 {
		ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
		defer cancel()
		sockChan := make(chan net.Conn, 1)
		go retryConnectTarget(sockChan, addr)

		select {
		case <-ctx.Done():
			return nil, errors.New("Dial cannot connect to master")
		case conn = <-sockChan:
			ILog.Println("Retry Dial connected to ", addr)
			return conn, nil
		}
	}

	return nil, errors.New("Dial cannot connect to master")
}

func sendCommand(conn net.Conn, command string) ([]byte, uint32, error) {

	n, err := conn.Write(MAGIC_CODE)
	if err != nil || n == 0 {
		ELog.Printf("Write MAGIC_CODE error: %s \n", err)
		return nil, 0, err
	}

	msgLen := uint32(len(command))
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, msgLen)

	DLog.Println("Write JSON size:", msgLen)
	n, err = conn.Write([]byte(size))
	if err != nil || n == 0 {
		ELog.Printf("Write DATA Size error: %s \n", err)
		return nil, 0, err
	}

	n, err = conn.Write([]byte(command))
	if err != nil || n == 0 {
		ELog.Printf("Write error: %s \n", err)
		return nil, 0, err
	}

	n, err = conn.Read(size)
	retSize := binary.BigEndian.Uint32(size)

	if err != nil || n == 0 {
		ELog.Printf("Read error: %s \n", err)
		return nil, 0, err
	}

	buf := make([]byte, 0)
	tmp := make([]byte, retSize)
	count := uint32(0)
	for {
		n, err = conn.Read(tmp)
		if err != nil {
			return nil, 0, err
		}
		count += uint32(n)
		buf = append(buf, tmp[:n]...)
		if count >= retSize {
			return buf, retSize, nil
		}
	}
	return buf, retSize, nil
}

func handleConnectTarget(ums *UlaMultiConnector, chanId int, targetNodeAddr TargetNodeAddr, sendChan chan string, respChan chan UlaCommandResponse, wg *sync.WaitGroup) {

	var err error
	conn, err := connectTarget(targetNodeAddr.TargetAddr, 0)
	if err != nil {
		WLog.Println("Failed connect target: ", targetNodeAddr.TargetAddr, " err: ", err)
		wg.Done()
		return
	}
	defer conn.Close()

	Mutex.Lock()
	ums.sendChans[chanId] = sendChan
	ums.respChans[chanId] = respChan
	Mutex.Unlock()

	wg.Done()
	for {
		select {
		case command := <-sendChan:
			jsonCommand, err := ulavscreen.ApplyAndGenCommand(command, targetNodeAddr.NodeId)
			if err != nil {
				ELog.Printf("Apply and Generate command Fail: %s \n", err)
				continue
			}
			respBuf, respSize, err := sendCommand(conn, jsonCommand)
			var ucr UlaCommandResponse
			if err != nil {
				if err == io.EOF || errors.Is(err, syscall.EPIPE) {
					WLog.Println("Connection closed, Retrying connect to ", targetNodeAddr.TargetAddr)
					conn.Close()
					conn, err = connectTarget(targetNodeAddr.TargetAddr, 1)
					if err != nil {
						WLog.Println("Reconnection failed for ", targetNodeAddr.TargetAddr)
						Mutex.Lock()
						ums.sendChans[chanId] = nil
						ums.respChans[chanId] = nil
						Mutex.Unlock()
						return
					}
					ILog.Println("Successfully reconnected to ", targetNodeAddr.TargetAddr)
					continue
				}
				ELog.Printf("Send command Fail: %s \n", err)
			} else {
				err = json.Unmarshal([]byte(string(respBuf[:respSize])), &ucr)
				if err != nil {
					ELog.Printf("Unmarshal json command error: %s \n", err)
				}
			}

			respChan <- ucr
		}
	}
}

func (ums *UlaMultiConnector) countConnection() int {
	connectNum := 0
	for chanId := range ums.targetNodeAddrs {
		if ums.sendChans[chanId] == nil {
			continue
		}
		if ums.respChans[chanId] == nil {
			continue
		}
		connectNum++
	}

	return connectNum
}

func (ums *UlaMultiConnector) handleConnectTargets() {

	var wg sync.WaitGroup
	for chanId, targetNodeAddr := range ums.targetNodeAddrs {
		if ums.sendChans[chanId] == nil || ums.respChans[chanId] == nil {
			wg.Add(1)
			sendChan := make(chan string, 1)
			respChan := make(chan UlaCommandResponse, 1)
			go handleConnectTarget(ums, chanId, targetNodeAddr, sendChan, respChan, &wg)
		} else {
			WLog.Println("targetNodeAddr ", targetNodeAddr, " has already connected to ula-node")
		}
	}
	wg.Wait()
}

func waitResponse(waitTime time.Duration, respChan chan UlaCommandResponse, targetAddr string, wg *sync.WaitGroup, resp **UlaCommandResponse) {

	t := time.NewTicker(waitTime * time.Second)
	defer t.Stop()
	defer wg.Done()

	select {
	case ucr := <-respChan:
		*resp = &ucr
		break
	case <-t.C:
		timeoutResp := UlaCommandResponse{
			Type:   "result",
			Result: -1,
		}
		*resp = &timeoutResp
		ELog.Printf("Command response watchdog was timeout. target: %s", targetAddr)
		break
	}
}

func (ums *UlaMultiConnector) sendCommand(command string) UlaCommandResponse {
	Mutex.Lock()
	var wg sync.WaitGroup
	resps := make([]*UlaCommandResponse, len(ums.sendChans))
	for chanId, sendChan := range ums.sendChans {
		if sendChan != nil && ums.respChans[chanId] != nil {
			wg.Add(1)
			sendChan <- command
			go waitResponse(1, ums.respChans[chanId], ums.targetNodeAddrs[chanId].TargetAddr, &wg, &resps[chanId])
		}
	}
	wg.Wait()

	ret := mergeResponses(resps)
	Mutex.Unlock()

	return ret
}

func (ums *UlaMultiConnector) SendLayoutCommand(command string) error {
	connectNum := ums.countConnection()
	if connectNum < len(ums.targetNodeAddrs) {
		ums.handleConnectTargets()
		connectNum = ums.countConnection()
		if connectNum == 0 {
			return errors.New("All targets cannot connect master")
		}

		if !ums.force {
			if connectNum < len(ums.targetNodeAddrs) {
				return errors.New(fmt.Sprintf("Some targets cannot connect master (%d < %d)", connectNum, len(ums.targetNodeAddrs)))
			}
		}
	}

	ucr := ums.sendCommand(command)
	if ucr.Type == "result" {
		ret := ucr.Result
		if ret != 0 {
			return errors.New("SendLayoutCommand Failed")
		}
	} else {
		return errors.New("result format type miss matched")
	}

	return nil
}

func mergeResponses(resps []*UlaCommandResponse) UlaCommandResponse {
	var ret UlaCommandResponse
	for _, resp := range resps {
		if resp != nil {
			ret.Type = resp.Type
			switch ret.Type {
			case "result":
				ret.Result = ret.Result | resp.Result
				break
			}
		}
	}
	return ret

}

func getDistribNodes(vsdPath ...string) ([]DistribNode, error) {
	var dNodes []DistribNode
	vscrnDef, err := ula.ReadVScrnDef(vsdPath...)
	if err != nil {
		return dNodes, err
	}
	dNodes = make([]DistribNode, 0)
	for _, node := range vscrnDef.Nodes {
		for _, frameworkNode := range vscrnDef.DistributedWindowSystem.FrameworkNode {
			if node.NodeId != frameworkNode.NodeId {
				continue
			}

			var tmp DistribNode
			tmp.NodeId = node.NodeId
			tmp.Ip = node.Ip
			tmp.Port = frameworkNode.Ula.Port
			dNodes = append(dNodes, tmp)
		}

	}
	return dNodes, nil
}

func UlaConnectionInit(force bool, vsdPath ...string) error {
	var err error
	UlaMulCon, err = newUlaMultiConn(force, vsdPath...)
	if err != nil {
		return err
	}

	UlaMulCon.handleConnectTargets()
	return nil
}
