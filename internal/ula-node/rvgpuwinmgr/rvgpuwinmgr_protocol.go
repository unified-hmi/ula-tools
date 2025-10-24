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

package rvgpuwinmgr

import (
	"encoding/json"
	"sync"
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-node"
	. "ula-tools/internal/ulog"
)

const UHMI_RVGPU_LAYOUT_SOCK string = "uhmi-rvgpu_layout_sock"
const VERSION string = "0.0.0"
const OPACITY float64 = 1.0

type rvgpuLayoutJson struct {
	Id             int     `json:"id"`
	RvgpuSurfaceID string  `json:"rvgpu_surface_id"`
	Width          int     `json:"width"`
	Height         int     `json:"height"`
	SrcX           int     `json:"src_x"`
	SrcY           int     `json:"src_y"`
	SrcW           int     `json:"src_w"`
	SrcH           int     `json:"src_h"`
	DstX           int     `json:"dst_x"`
	DstY           int     `json:"dst_y"`
	DstW           int     `json:"dst_w"`
	DstH           int     `json:"dst_h"`
	Opacity        float64 `json:"opacity"`
	Visibility     int     `json:"visibility"`
}

type safetyAreaJson struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type InitialLayoutProtocol struct {
	Version      string            `json:"version"`
	Command      string            `json:"command"`
	RvgpuLayouts []rvgpuLayoutJson `json:"surfaces"`
	SafetyAreas  []safetyAreaJson  `json:"safety_areas"`
}

type Key struct {
	RDisplayId int
	LayerID    int
}

var layerSurfacesMap = make(map[Key][]ula.PixelSurface)
var mapMutex = sync.RWMutex{}

func addOrUpdateLayerSurfaces(rDisplayID int, layerID int, surfaces []ula.PixelSurface) {
	mapMutex.Lock()
	defer mapMutex.Unlock()
	layerSurfacesMap[Key{rDisplayID, layerID}] = surfaces
}

func genRvgpuLayoutParams(player ula.PixelLayer, psurfaces []ula.PixelSurface) []rvgpuLayoutJson {
	var rvgpuLayouts []rvgpuLayoutJson
	for _, psurf := range psurfaces {
		viewSrcX, viewSrcY, viewSrcW, viewSrcH, viewDstX, viewDstY, viewDstW, viewDstH := calcSrcViewArea(player, psurf)

		rvgpuLayout := rvgpuLayoutJson{
			Id:             player.VID,
			RvgpuSurfaceID: psurf.AppName, /* appli name */
			Width:          psurf.PixelW,
			Height:         psurf.PixelH,
			SrcX:           viewSrcX,
			SrcY:           viewSrcY,
			SrcW:           viewSrcW,
			SrcH:           viewSrcH,
			DstX:           viewDstX,
			DstY:           viewDstY,
			DstW:           viewDstW,
			DstH:           viewDstH,
			Opacity:        OPACITY,
			Visibility:     psurf.Visibility,
		}
		rvgpuLayouts = append(rvgpuLayouts, rvgpuLayout)
	}
	return rvgpuLayouts
}

func calcSrcViewArea(player ula.PixelLayer, psurface ula.PixelSurface) (int, int, int, int, int, int, int, int) {
	clipX, clipY, clipWidth, clipHeight := 0, 0, 0, 0
	clipX2, clipY2 := 0, 0

	SurfaceSrcX := psurface.PsrcX
	SurfaceSrcY := psurface.PsrcY
	SurfaceSrcW := psurface.PsrcW
	SurfaceSrcH := psurface.PsrcH

	LayerSrcX := player.PsrcX
	LayerSrxW := player.PsrcW
	SurfaceDstX := psurface.PdstX
	SurfaceDstW := psurface.PdstW

	LayerSrcY := player.PsrcY
	LayerSrxH := player.PsrcH
	SurfaceDstY := psurface.PdstY
	SurfaceDstH := psurface.PdstH

	// Calculate x clipping
	if LayerSrcX+LayerSrxW <= SurfaceDstX || LayerSrcX >= SurfaceDstX+SurfaceDstW {
		clipX = 0
		clipWidth = 0
	} else if LayerSrcX < SurfaceDstX && LayerSrcX+LayerSrxW > SurfaceDstX {
		clipX = SurfaceSrcX
		clipX2 = SurfaceDstX
		if LayerSrcX+LayerSrxW > SurfaceDstX+SurfaceDstW {
			clipWidth = SurfaceDstW
		} else {
			clipWidth = LayerSrcX + LayerSrxW - SurfaceDstX
		}
	} else if LayerSrcX >= SurfaceDstX && LayerSrcX < SurfaceDstX+SurfaceDstW {
		clipX = LayerSrcX
		if LayerSrcX+LayerSrxW <= SurfaceDstX+SurfaceDstW {
			clipWidth = LayerSrxW
		} else {
			clipWidth = SurfaceDstX + SurfaceDstW - LayerSrcX
		}
	}
	if clipX+clipWidth > player.PixelW {
		clipWidth = player.PixelW - clipX
	}

	// Calculate y clipping
	if LayerSrcY+LayerSrxH <= SurfaceDstY || LayerSrcY >= SurfaceDstY+SurfaceDstH {
		clipY = 0
		clipHeight = 0
	} else if LayerSrcY < SurfaceDstY && LayerSrcY+LayerSrxH > SurfaceDstY {
		clipY = SurfaceSrcY
		clipY2 = SurfaceDstY
		if LayerSrcY+LayerSrxH <= SurfaceDstY+SurfaceDstH {
			clipHeight = SurfaceDstH
		} else {
			clipHeight = LayerSrcY + LayerSrxH - SurfaceDstY
		}
	} else if LayerSrcY >= SurfaceDstY && LayerSrcY < SurfaceDstY+SurfaceDstH {
		clipY = LayerSrcY
		if LayerSrcY+LayerSrxH <= SurfaceDstY+SurfaceDstH {
			clipHeight = LayerSrxH
		} else {
			clipHeight = SurfaceDstY + SurfaceDstH - LayerSrcY
		}
	}
	if clipY+clipHeight > player.PixelH {
		clipHeight = player.PixelH - clipY
	}

	// Scale the clipped region based on the surface source and destination sizes
	finalSrcX := int(float64(clipX)/float64(SurfaceDstW)*float64(SurfaceSrcW)) + SurfaceSrcX
	finalSrcWidth := int(float64(clipWidth) / float64(SurfaceDstW) * float64(SurfaceSrcW))

	finalSrcY := int(float64(clipY)/float64(SurfaceDstH)*float64(SurfaceSrcH)) + SurfaceSrcY
	finalSrcHeight := int(float64(clipHeight) / float64(SurfaceDstH) * float64(SurfaceSrcH))

	finalDstX := player.PdstX + clipX2
	finalDstY := player.PdstY + clipY2
	finalDstWidth := player.PdstW
	finalDstHeight := player.PdstH

	return finalSrcX, finalSrcY, finalSrcWidth, finalSrcHeight, finalDstX, finalDstY, finalDstWidth, finalDstHeight
}

func genInitialLayoutProtocolJson(req ulanode.LocalCommandReq, rId int) (string, error) {
	var rvgpuLayouts []rvgpuLayoutJson
	var safetyareas []safetyAreaJson

	for _, rdcomm := range req.RDComms {

		if rdcomm.Rdisplay.RDisplayId == rId {

			for _, player := range rdcomm.Players {

				rvgpuLayout := genRvgpuLayoutParams(player, player.Psurfaces)
				addOrUpdateLayerSurfaces(rId, player.VID, player.Psurfaces)

				rvgpuLayouts = append(rvgpuLayouts, rvgpuLayout...)
			}
			for _, r := range rdcomm.SafetyAreas {
				rvgpuLayout := safetyAreaJson{
					X:      r.PixelX,
					Y:      r.PixelY,
					Width:  r.PixelW,
					Height: r.PixelH,
				}
				safetyareas = append(safetyareas, rvgpuLayout)
			}
		}
	}

	rvgpuProto := InitialLayoutProtocol{
		Version:      VERSION,
		Command:      "initial_layout",
		RvgpuLayouts: rvgpuLayouts,
		SafetyAreas:  safetyareas,
	}

	jsonBytes, err := json.Marshal(rvgpuProto)
	if err != nil {
		ELog.Println("JSON Marshal error:", err)
	}

	msg := string(jsonBytes)

	return msg, nil
}
