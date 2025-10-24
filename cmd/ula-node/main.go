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
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"reflect"
	"strconv"
	_ "strings"
	"sync"
	_ "time"
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-node"
	"ula-tools/internal/ula-node/iviwinmgr"
	"ula-tools/internal/ula-node/rvgpuwinmgr"
	. "ula-tools/internal/ulog"
)

var MAGIC_CODE []byte = []byte{0x55, 0x4C, 0x41, 0x30} // 'ULA0' ascii code

var Mutex struct {
	sync.Mutex
}

func readConnection(conn net.Conn) ([]byte, uint32, error) {

	cbio := bufio.NewReader(conn)

	magicBuf := make([]byte, 4)

	_, err := io.ReadFull(cbio, magicBuf[:4])
	if err != nil {
		if err == io.EOF {
			return nil, 0, err
		} else {
			DLog.Printf("Magic Size Read Failed: %s\n", err)
			return nil, 0, errors.New("zero byte read(maybe connection check)")
		}
	}
	if reflect.DeepEqual(magicBuf, MAGIC_CODE) == false {
		return nil, 0, errors.New("Magic Code Read Fail")
	}

	szBuf := make([]byte, 4)
	_, err = io.ReadFull(cbio, szBuf[:4])
	if err != nil {
		return nil, 0, errors.New("Command Size Read Fail")
	}

	recvSize := binary.BigEndian.Uint32(szBuf[:4])
	if recvSize == 0 {
		return nil, recvSize, errors.New("zero byte read(meaningless)")
	}

	recvBuf := make([]byte, recvSize)
	_, err = io.ReadFull(cbio, recvBuf[:recvSize])
	if err != nil {
		return recvBuf, recvSize, errors.New(fmt.Sprintf("Command Read Fail: %s \n", err))
	}
	return recvBuf, recvSize, nil
}

func readCommandLoop(conn net.Conn, jsonChan chan map[string]interface{}, listenerId int, retChansMap map[int]interface{}) {

	defer conn.Close()

	for {
		recvBuf, recvSize, err := readConnection(conn)
		if err != nil {
			if err == io.EOF {
				DLog.Printf("Ula-node zero byte read(maybe Client closed the connection)\n")
			} else {
				ELog.Printf("Ula-node command Read Fail: %s \n", err)
			}
			break
		}
		mJson := make(map[string]interface{})
		err = json.Unmarshal([]byte(string(recvBuf[:recvSize])), &mJson)
		if err != nil {
			ELog.Printf("Unmarshal json command error: %s \n", err)
			break
		}

		mJson["listener_id"] = listenerId

		jsonChan <- mJson

		Mutex.Lock()
		retChan := retChansMap[listenerId].(chan map[string]interface{})
		Mutex.Unlock()
		select {
		case retJson := <-retChan:
			retBuf, _ := json.Marshal(retJson)
			msgLen := uint32(len(retBuf))
			size := make([]byte, 4)
			binary.BigEndian.PutUint32(size, msgLen)
			n, err := conn.Write([]byte(size))
			if err != nil || n == 0 {
				ELog.Printf("Write DATA Size error: %s \n", err)
			}

			n, err = conn.Write(retBuf)
			if err != nil || n == 0 {
				ELog.Printf("Write error: %s \n", err)
			}
			break
		}

	}
	Mutex.Lock()
	delete(retChansMap, listenerId)
	Mutex.Unlock()
}

func processCommandLoop(
	nodeId int,
	reqChan chan ulanode.LocalCommandReq,
	respChan chan ulanode.LocalCommandReq,
	jsonChan chan map[string]interface{},
	retChansMap map[int]interface{},
	plugin ulanode.LocalCommandGenerator,
) {
	spscrns := new(ula.NodePixelScreens)
	for {
		var mJson map[string]interface{}
		select {
		case mJson = <-jsonChan:
			break
		}

		ret := 0
		listenerId := mJson["listener_id"].(int)
		jsonBytes, err := json.Marshal(mJson)
		if err != nil {
			ret = -1
			commResponseResult(ret, listenerId, retChansMap)
			continue
		}

		acdata := new(ula.ApplyCommandData)
		err = json.Unmarshal(jsonBytes, acdata)
		if err != nil {
			ret = -1
			commResponseResult(ret, listenerId, retChansMap)
			continue
		}

		reqs, err := plugin.GenerateLocalCommandReq(acdata, spscrns)
		if err != nil {
			ret = -1
			commResponseResult(ret, listenerId, retChansMap)
			continue
		}

		ret = submitCommand(reqs, reqChan, respChan)

		spscrns = acdata.NPScreens

		commResponseResult(ret, listenerId, retChansMap)
	}
}

func submitCommand(
	reqs []*ulanode.LocalCommandReq,
	reqChan chan ulanode.LocalCommandReq,
	respChan chan ulanode.LocalCommandReq,
) int {

	ret := 0
	for _, req := range reqs {
		reqChan <- *req
		select {
		case lcr := <-respChan:
			ret = lcr.Ret
			break
		}
	}
	return ret
}

func commResponseResult(
	result int,
	listenerId int,
	retChansMap map[int]interface{},
) {
	Mutex.Lock()
	respChan := retChansMap[listenerId].(chan map[string]interface{})
	Mutex.Unlock()
	retJson := map[string]interface{}{
		"type":   "result",
		"result": result,
	}
	respChan <- retJson
}

func mainLoop(
	listener net.Listener,
	nodeId int,
	reqChan chan ulanode.LocalCommandReq,
	respChan chan ulanode.LocalCommandReq,
	plugin ulanode.LocalCommandGenerator) {

	jsonChan := make(chan map[string]interface{}, 1)
	retChansMap := make(map[int]interface{})
	go processCommandLoop(nodeId, reqChan, respChan, jsonChan, retChansMap, plugin)

	listenerId := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			ELog.Printf("Accept error: %s", err)
			continue
		}
		retChan := make(chan map[string]interface{}, 1)
		Mutex.Lock()
		retChansMap[listenerId] = retChan
		Mutex.Unlock()
		go readCommandLoop(conn, jsonChan, listenerId, retChansMap)
		listenerId += 1
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "%s [option] | listenIp listenPort nodeId\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "[option]\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = printUsage

	var (
		verbose      bool
		debug        bool
		vScrnDefFile string
		keyNodeId    int
		keyHostName  string
	)

	flag.BoolVar(&verbose, "v", true, "verbose info log")
	flag.BoolVar(&debug, "d", false, "verbose debug log")
	flag.StringVar(&vScrnDefFile, "f", "", "virtual-screen-def.json file Path")
	flag.IntVar(&keyNodeId, "N", -1, "search ula-node param by node_id from VScrnDef file")
	flag.StringVar(&keyHostName, "H", "", "search ula-node param by hostname from VScrnDef file")

	flag.Parse()

	if verbose == true {
		ILog.SetOutput(os.Stderr)
	}

	if debug == true {
		DLog.SetOutput(os.Stderr)
	}

	DLog.Printf("ARG0:%s, ARG1:%s, ARG2:%s", flag.Arg(0), flag.Arg(1), flag.Arg(2))

	vscrnDef, err := ula.ReadVScrnDef(vScrnDefFile)
	if err != nil {
		ELog.Println("ReadVScrnDef error : ", err)
		return
	}

	var (
		listenIp   string
		listenPort int
		nodeId     int
	)

	if len(flag.Args()) == 0 {

		if keyNodeId != -1 && keyHostName != "" {
			ELog.Println("error multiple search keys")
			return
		}

		if keyHostName == "" {
			keyHostName, err = os.Hostname()
			if err != nil {
				ELog.Println("os.Hostname error : ", err)
				return
			}
		}

		/* search ulaNode Params */
		if keyNodeId != -1 {
			nodeId = keyNodeId
		} else {
			nodeId, err = vscrnDef.GetNodeIdByHostName(keyHostName)
			if err != nil {
				ELog.Println("GetNodeIdByHostName error : ", err)
				return
			}
		}

		/* search my Ip Addr */
		ipAddrs, err := ula.GetIpv4AddrsOfAllInterfaces()
		if err != nil {
			ELog.Println("GetIpv4AddrsOfAllInterfaces error : ", err)
			return
		}
		listenIp, err = vscrnDef.GetIpAddrByNodeIdAndIpCandidateList(ipAddrs, nodeId)
		if err != nil {
			ELog.Println("GetIpAddrByNodeIdAndIpCandidateList error : ", err)
			return
		}

		listenPort, err = vscrnDef.GetPort(nodeId)
		if err != nil {
			ELog.Println("GetPort error : ", err)
			return
		}
	} else if len(flag.Args()) == 3 {
		listenIp = flag.Arg(0)
		listenPort, _ = strconv.Atoi(flag.Arg(1))
		nodeId, _ = strconv.Atoi(flag.Arg(2))
	} else {
		printUsage()
		return
	}

	prefix := "ula-node-" + strconv.Itoa(nodeId)
	SetLogPrefix(prefix)
	DLog.Println(listenIp, ":", listenPort)

	listenAddr := listenIp + ":" + strconv.Itoa(listenPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		ELog.Printf("Listen error: %s", err)
		return
	}

	defer listener.Close()
	reqChan := make(chan ulanode.LocalCommandReq, 5)
	respChan := make(chan ulanode.LocalCommandReq, 5)

	var plugin ulanode.LocalCommandGenerator
	plugin = iviwinmgr.IviPlugin{}

	if rvgpuwinmgr.IsRvgpuCompositor(vscrnDef, nodeId) {
		plugin = rvgpuwinmgr.RvgpuPlugin{}
	}
	go plugin.Start(reqChan, respChan)

	mainLoop(listener, nodeId, reqChan, respChan, plugin)
}

//export StartUlanode
func StartUlanode(vScrnDefFile string, keyNodeId int, keyHostName string, keyIpAddr string) {

	vscrnDef, err := ula.ReadVScrnDef(vScrnDefFile)
	if err != nil {
		ELog.Println("ReadVScrnDef fail", err)
		return
	}
	var (
		listenIp   string
		listenPort int
		nodeId     int
	)

	listenIp = keyIpAddr
	nodeId, err = vscrnDef.GetNodeIdByHostName(keyHostName)
	if err != nil {
		ELog.Println("getNodeIdByHostName error : ", err)
		return
	}

	listenPort, err = vscrnDef.GetPort(nodeId)
	if err != nil {
		ELog.Println("getPort error : ", err)
		return
	}

	listenAddr := listenIp + ":" + strconv.Itoa(listenPort)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		ELog.Printf("Listen error: %s", err)
		return
	}

	defer listener.Close()
	reqChan := make(chan ulanode.LocalCommandReq, 5)
	respChan := make(chan ulanode.LocalCommandReq, 5)

	var plugin ulanode.LocalCommandGenerator
	plugin = iviwinmgr.IviPlugin{}

	if rvgpuwinmgr.IsRvgpuCompositor(vscrnDef, nodeId) {
		plugin = rvgpuwinmgr.RvgpuPlugin{}
	}
	go plugin.Start(reqChan, respChan)

	mainLoop(listener, nodeId, reqChan, respChan, plugin)
}
