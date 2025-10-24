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
	"encoding/json"
	"ula-tools/internal/ula-node"
	. "ula-tools/internal/ulog"
)

const UHMI_IVI_WM_SOCK string = "/tmp/uhmi-ivi-wm_sock"

var MAGIC_CODE []byte = []byte{0x55, 0x4C, 0x41, 0x30} // 'ULA0' ascii code

const VERSION string = "1.0.0"
const OPACITY float64 = 1.0
const VISIBILITY int = 1

type IviSurfaceJson struct {
	Id         int     `json:"id"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	SrcX       int     `json:"src_x"`
	SrcY       int     `json:"src_y"`
	SrcW       int     `json:"src_w"`
	SrcH       int     `json:"src_h"`
	DstX       int     `json:"dst_x"`
	DstY       int     `json:"dst_y"`
	DstW       int     `json:"dst_w"`
	DstH       int     `json:"dst_h"`
	Opacity    float64 `json:"opacity"`
	Visibility int     `json:"visibility"`
}

type IviLayerJson struct {
	Id         int              `json:"id"`
	Width      int              `json:"width"`
	Height     int              `json:"height"`
	SrcX       int              `json:"src_x"`
	SrcY       int              `json:"src_y"`
	SrcW       int              `json:"src_w"`
	SrcH       int              `json:"src_h"`
	DstX       int              `json:"dst_x"`
	DstY       int              `json:"dst_y"`
	DstW       int              `json:"dst_w"`
	DstH       int              `json:"dst_h"`
	Opacity    float64          `json:"opacity"`
	Visibility int              `json:"visibility"`
	Surface    []IviSurfaceJson `json:"surfaces"`
}

type IviRDisplay struct {
	RDisplayId  int            `json:"id"`
	InsertOrder string         `json:"insert_order"`
	ReferenceId int            `json:"referenceID"`
	Layers      []IviLayerJson `json:"layers"`
}

type InitialScreenProtocol struct {
	Version  string        `json:"version"`
	Command  string        `json:"command"`
	RDisplay []IviRDisplay `json:"screens"`
}

func genInitialScreenProtocolJson(req ulanode.LocalCommandReq) (string, error) {
	var iviRDisp []IviRDisplay

	for _, rdcomm := range req.RDComms {

		var ivilayers []IviLayerJson

		for _, player := range rdcomm.Players {

			var ivisurfs []IviSurfaceJson

			for _, psurf := range player.Psurfaces {

				iviSurf := IviSurfaceJson{
					Id:         psurf.VID,
					Width:      psurf.PixelW,
					Height:     psurf.PixelH,
					SrcX:       psurf.PsrcX,
					SrcY:       psurf.PsrcY,
					SrcW:       psurf.PsrcW,
					SrcH:       psurf.PsrcH,
					DstX:       psurf.PdstX,
					DstY:       psurf.PdstY,
					DstW:       psurf.PdstW,
					DstH:       psurf.PdstH,
					Opacity:    OPACITY,
					Visibility: psurf.Visibility,
				}

				ivisurfs = append(ivisurfs, iviSurf)
			}

			iviLayer := IviLayerJson{
				Id:         player.VID,
				Width:      player.PixelW,
				Height:     player.PixelH,
				SrcX:       player.PsrcX,
				SrcY:       player.PsrcY,
				SrcW:       player.PsrcW,
				SrcH:       player.PsrcH,
				DstX:       player.PdstX,
				DstY:       player.PdstY,
				DstW:       player.PdstW,
				DstH:       player.PdstH,
				Opacity:    OPACITY,
				Visibility: player.Visibility,
				Surface:    ivisurfs,
			}

			ivilayers = append(ivilayers, iviLayer)
		}

		rdisp := IviRDisplay{
			RDisplayId: rdcomm.Rdisplay.RDisplayId,
			Layers:     ivilayers,
		}

		iviRDisp = append(iviRDisp, rdisp)
	}

	iviProto := InitialScreenProtocol{
		Version:  VERSION,
		Command:  "initial_screen",
		RDisplay: iviRDisp,
	}

	jsonBytes, err := json.Marshal(iviProto)
	if err != nil {
		ELog.Println("JSON Marshal error:", err)
	}

	msg := string(jsonBytes)

	return msg, nil
}
