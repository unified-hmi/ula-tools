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

package iviwinmgr

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"reflect"
	"sync"
	"time"
	"ula-tools/internal/ula-node"
	. "ula-tools/internal/ulog"
)

type iviWinMgr struct {
	conn     net.Conn
	sendChan chan string
	waitChan chan []byte
	recvChan chan []byte
}

func retryConnectTarget(sockChan chan net.Conn, stopChan chan struct{}) {
	for {
		conn, err := net.Dial("unix", UHMI_IVI_WM_SOCK)
		if err == nil {
			sockChan <- conn
			break
		}

		time.Sleep(10 * time.Millisecond)

		select {
		case <-stopChan:
			return
		default:
		}
	}
}

func connectTarget() net.Conn {
	var conn net.Conn

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	sockChan := make(chan net.Conn, 1)
	stopChan := make(chan struct{})

	defer func() {
		close(sockChan)
		close(stopChan)
	}()

	go retryConnectTarget(sockChan, stopChan)

	select {
	case <-ctx.Done():
		ELog.Println("Dial cannot connect to uhmi-ivi-wm")
		return nil
	case conn = <-sockChan:
		ILog.Println("Dial connected to uhmi-ivi-wm")
	}

	DLog.Printf("connect OK\n")
	return conn
}

func connectTargetOnce() net.Conn {
	var conn net.Conn

	conn, err := net.Dial("unix", UHMI_IVI_WM_SOCK)
	if err != nil {
		ELog.Println("Dial cannot connect to uhmi-ivi-wm")
		return nil
	}

	DLog.Printf("connect to uhmi-ivi-wm OK\n")
	return conn
}

func handleConnectTarget(iviwinmgr *iviWinMgr, isretry bool, wg *sync.WaitGroup) {

	if isretry == true {
		iviwinmgr.conn = connectTarget()
	} else {
		iviwinmgr.conn = connectTargetOnce()
	}

	if iviwinmgr.conn == nil {
		wg.Done()
		return
	}
	iviwinmgr.waitChan = make(chan []byte, 1)
	iviwinmgr.sendChan = make(chan string, 1)
	iviwinmgr.recvChan = make(chan []byte, 1)

	defer func() {
		iviwinmgr.conn.Close()
		iviwinmgr.conn = nil
		close(iviwinmgr.waitChan)
		close(iviwinmgr.sendChan)
		close(iviwinmgr.recvChan)
	}()

	go connReadLoop(iviwinmgr.conn, iviwinmgr.recvChan)
	wg.Done()

	for {
		select {
		case recvMsg := <-iviwinmgr.recvChan:
			if recvMsg != nil {
				iviwinmgr.waitChan <- recvMsg
			} else {
				iviwinmgr.waitChan <- nil
				return
			}
		}
	}
}

func (plugin IviPlugin) Start(reqChan chan ulanode.LocalCommandReq, respChan chan ulanode.LocalCommandReq) {

	var wg sync.WaitGroup
	var iviwinmgr iviWinMgr
	wg.Add(1)
	isRetry := true
	go handleConnectTarget(&iviwinmgr, isRetry, &wg)
	wg.Wait()

	for {
		select {
		case wVDsp := <-reqChan:
			if iviwinmgr.conn == nil {
				wg.Add(1)
				isRetry = false
				go handleConnectTarget(&iviwinmgr, isRetry, &wg)
				wg.Wait()
			}

			ret := sendUhmiIviWmJson(&iviwinmgr, wVDsp)
			lcr := ulanode.LocalCommandReq{}
			lcr.Ret = ret
			respChan <- lcr
			break
		}
	}

}

func sendMagicCode(iviwinmgr *iviWinMgr) error {

	n, err := iviwinmgr.conn.Write(MAGIC_CODE)
	if err != nil || n == 0 {
		return errors.New(fmt.Sprintf("Write error: %s \n", err))
	}

	select {
	case result := <-iviwinmgr.waitChan:
		if result != nil {
			if reflect.DeepEqual(result, MAGIC_CODE) == false {
				return errors.New(fmt.Sprintf("Read magic code false: %s\n", result))
			}
			return nil
		} else {
			return errors.New(fmt.Sprintf("Read error: %s \n", err))
		}
	}
}

func sendUhmiIviWmJson(iviwinmgr *iviWinMgr, req ulanode.LocalCommandReq) int {

	DLog.Println("sendUhmiIviWmJson's reqCommand:", req.Command)

	msg := ""
	var err error

	switch req.Command {
	case "initial_vscreen":
		msg, err = genInitialScreenProtocolJson(req)
	case "local_comm":
		return 0
	default:
		ELog.Println("Error req.Command")
		return -1
	}
	if err != nil {
		ELog.Println("Error ProtocolJson")
		return -1
	}

	if iviwinmgr.conn == nil {
		ELog.Printf("Error Not connected to uhmi-ivi-wm")
		return -1
	}

	err = sendMagicCode(iviwinmgr)
	if err != nil {
		ELog.Printf("Error SendRecv MagicCode: %s", err)
		return -1
	}

	DLog.Println("sendCommand", req)
	err = sendCommand(iviwinmgr, msg)
	if err != nil {
		ELog.Printf("Error SendCommand: %s", err)
		return -1
	}

	return 0

}

func sendCommand(iviwinmgr *iviWinMgr, command string) error {

	msgLen := uint32(len(command))
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, msgLen)

	n, err := iviwinmgr.conn.Write(size)
	if err != nil || n == 0 {
		ELog.Printf("Write DATA Size error: %s \n", err)
	}

	n, err = iviwinmgr.conn.Write([]byte(command))
	if err != nil || n == 0 {
		ELog.Printf("Write error: %s \n", err)
	}

	select {
	case result := <-iviwinmgr.waitChan:
		if result != nil {
			ret := binary.BigEndian.Uint32(result)
			DLog.Printf("Read uhmi-ivi-wm ret: %x", ret)
			return nil
		} else {
			return errors.New(fmt.Sprintf("Read uhmi-ivi-wm error \n"))
		}
	}
}

func connReadLoop(conn net.Conn, rcvChan chan []byte) {

	cbio := bufio.NewReader(conn)
	for {
		recvBuf := make([]byte, 4)
		_, err := io.ReadFull(cbio, recvBuf)
		if err != nil {
			rcvChan <- nil
			return
		}
		rcvChan <- recvBuf
	}
}
