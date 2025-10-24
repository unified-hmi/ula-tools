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

package ulavscreen

import (
	"errors"
	"ula-tools/internal/ula"
)

type GeometoryConverter interface {
	DoConvert() error
	GetNodePixelScreens() (*ula.NodePixelScreens, error)
}

type WorkV2R struct {
	vdisplay     ula.VirtualDisplay
	rdisplay     ula.RealDisplay
	vlayers      []ula.VirtualLayer
	vsafetyareas []ula.VirtualSafetyArea

	/* Geometry is Coordinated by RDisplay  */
	players      []ula.PixelLayer
	psafetyareas []ula.PixelSafetyArea
}

type Vscreen2RdisplayConverter struct {
	nodeId     int
	workV2RMap map[int]WorkV2R
}

func NewVscreen2RdisplayConverter(
	vscrn *VirtualScreen,
	nodeId int) (*Vscreen2RdisplayConverter, error) {

	wdisplay, err := generateWorkV2R(vscrn, nodeId, &vscrn.VScrnDef)
	if err != nil {
		return nil, errors.New("NewVscreen2RdisplayConverter error")
	}

	vs2rd := &Vscreen2RdisplayConverter{
		nodeId:     nodeId,
		workV2RMap: wdisplay,
	}

	return vs2rd, nil
}

func (vs2rd *Vscreen2RdisplayConverter) DoConvert() error {

	convVDisplay2RDisplayCoordinate(vs2rd.workV2RMap)

	return nil
}

func convVSurface2PSurface(vsurf *ula.VirtualSurface) *ula.PixelSurface {

	psurf := ula.NewEmptyPixelSurface()

	psurf.AppName = vsurf.AppName

	psurf.ParentVID = vsurf.ParentVID
	psurf.VID = vsurf.VID

	psurf.PixelW = vsurf.PixelW
	psurf.PixelH = vsurf.PixelH

	psurf.PsrcX = vsurf.PsrcX
	psurf.PsrcY = vsurf.PsrcY
	psurf.PsrcW = vsurf.PsrcW
	psurf.PsrcH = vsurf.PsrcH

	psurf.PdstX = vsurf.VdstX
	psurf.PdstY = vsurf.VdstY
	psurf.PdstW = vsurf.VdstW
	psurf.PdstH = vsurf.VdstH

	psurf.Visibility = vsurf.Visibility
	return psurf
}

func convVLayer2PLayer(vlayer *ula.VirtualLayer) *ula.PixelLayer {

	player := ula.NewEmptyPixelLayer()

	player.AppName = vlayer.AppName

	player.VID = vlayer.VID

	player.PixelW = vlayer.VirtualW
	player.PixelH = vlayer.VirtualH

	player.PsrcX = vlayer.VsrcX
	player.PsrcY = vlayer.VsrcY
	player.PsrcW = vlayer.VsrcW
	player.PsrcH = vlayer.VsrcH

	player.PdstX = vlayer.VdstX
	player.PdstY = vlayer.VdstY
	player.PdstW = vlayer.VdstW
	player.PdstH = vlayer.VdstH

	player.Visibility = vlayer.Visibility

	player.Psurfaces = make([]ula.PixelSurface, 0)

	for _, sVsurf := range vlayer.Vsurfaces {
		copiedPSurf := convVSurface2PSurface(&sVsurf)
		player.Psurfaces = append(player.Psurfaces, *copiedPSurf)
	}

	return player
}

func convVLayers2PLayers(sVlayers []ula.VirtualLayer) (dPlayers []ula.PixelLayer) {
	if sVlayers == nil {
		return nil
	}

	dPlayers = make([]ula.PixelLayer, 0)

	for _, sVlayer := range sVlayers {
		dPlayer := convVLayer2PLayer(&sVlayer)
		dPlayers = append(dPlayers, *dPlayer)
	}

	return dPlayers
}

func convVSafetyArea2PSafetyArea(vSafetyArea *ula.VirtualSafetyArea) *ula.PixelSafetyArea {

	pSafetyArea := ula.NewEmptyPixelSafetyArea()

	pSafetyArea.PixelX = vSafetyArea.VirtualX
	pSafetyArea.PixelY = vSafetyArea.VirtualY
	pSafetyArea.PixelW = vSafetyArea.VirtualW
	pSafetyArea.PixelH = vSafetyArea.VirtualH

	return pSafetyArea
}

func convVSafetyAreas2PSafetyAreas(sVSafetyAreas []ula.VirtualSafetyArea) (dPSafetyAreas []ula.PixelSafetyArea) {
	if sVSafetyAreas == nil {
		return nil
	}

	dPSafetyAreas = make([]ula.PixelSafetyArea, 0)

	for _, sVSafetyArea := range sVSafetyAreas {
		dPSafetyArea := convVSafetyArea2PSafetyArea(&sVSafetyArea)
		dPSafetyAreas = append(dPSafetyAreas, *dPSafetyArea)
	}

	return dPSafetyAreas
}

func isNeedForWorkV2R(vlayer *ula.VirtualLayer, arg interface{}) bool {
	vDisplayId := arg.(int)

	if vlayer.Coord == ula.COORD_GLOBAL {
		return true
	}

	if vlayer.Coord == ula.COORD_VDISPLAY &&
		vlayer.VDisplayId == vDisplayId {
		return true
	}

	return false
}

func generateWorkV2R(
	vscreen *VirtualScreen,
	nodeId int,
	vscrnDef *ula.VScrnDef) (map[int]WorkV2R, error) {

	workV2RMap := make(map[int]WorkV2R)

	for vdspid, vdisplay := range vscreen.VirtualDisplays {

		isMe := vscrnDef.IsVDisplayInNode(nodeId, vdspid)
		if !isMe {
			continue
		}

		rdisplay := vscreen.RealDisplays[vdspid]

		wvdisp := WorkV2R{
			vdisplay:     *vdisplay.Dup(),
			rdisplay:     *rdisplay.Dup(),
			vlayers:      ula.DupVirtualLayerSliceIfNeed(vscreen.VdispVlayers[vdspid], isNeedForWorkV2R, vdspid),
			vsafetyareas: ula.DupVirtualSafetyAreaSlice(vscreen.VdispVsafetyAreas[vdspid]),
		}

		workV2RMap[vdspid] = wvdisp
	}

	return workV2RMap, nil
}

func convGlobalToVDisplayCoordinateSub(
	vdisp_vx int,
	vdisp_vw int,
	vlayer_vdx int,
	vlayer_vdw int,
	vlayer_vsx int,
	vlayer_vsw int) (int, int, int, int) {

	var (
		nvlayer_vdx int
		nvlayer_vdw int
		nvlayer_vsx int
		nvlayer_vsw int
	)

	/* Assume src cutout area is the entire layer in this part */
	if vdisp_vx <= vlayer_vdx &&
		vlayer_vdx <= vdisp_vx+vdisp_vw &&
		vdisp_vx+vdisp_vw <= vlayer_vdx+vlayer_vdw {

		nvlayer_vdx = vlayer_vdx - vdisp_vx
		nvlayer_vsx = 0
		nvlayer_vdw = vdisp_vx + vdisp_vw - vlayer_vdx
		nvlayer_vsw = nvlayer_vdw

	} else if vlayer_vdx <= vdisp_vx &&
		vdisp_vx+vdisp_vw <= vlayer_vdx+vlayer_vdw {

		nvlayer_vdx = 0
		nvlayer_vsx = vdisp_vx - vlayer_vdx
		nvlayer_vdw = vdisp_vw
		nvlayer_vsw = nvlayer_vdw

	} else if vlayer_vdx <= vdisp_vx &&
		vdisp_vx <= vlayer_vdx+vlayer_vdw &&
		vlayer_vdx+vlayer_vdw <= vdisp_vx+vdisp_vw {

		nvlayer_vdx = 0
		nvlayer_vsx = vdisp_vx - vlayer_vdx
		nvlayer_vdw = vlayer_vdx + vlayer_vdw - vdisp_vx
		nvlayer_vsw = nvlayer_vdw

	} else if vdisp_vx <= vlayer_vdx &&
		vlayer_vdx+vlayer_vdw <= vdisp_vx+vdisp_vw {

		nvlayer_vdx = vlayer_vdx - vdisp_vx
		nvlayer_vsx = 0
		nvlayer_vdw = vlayer_vdw
		nvlayer_vsw = nvlayer_vdw

	} else {
		nvlayer_vdx = 0
		nvlayer_vsx = 0
		nvlayer_vdw = 0
		nvlayer_vsw = 0
	}

	/* Corresponds to src cutout position */
	nvlayer_vsx = nvlayer_vsx*vlayer_vsw/vlayer_vdw + vlayer_vsx
	nvlayer_vsw = nvlayer_vsw * vlayer_vsw / vlayer_vdw

	return nvlayer_vdx, nvlayer_vdw, nvlayer_vsx, nvlayer_vsw

}

func convGlobalToVDisplayCoordinate(sVlayer *ula.VirtualLayer,
	vdisp *ula.VirtualDisplay) *ula.VirtualLayer {

	vdisp_vx := vdisp.VirtualX
	vdisp_vy := vdisp.VirtualY
	vdisp_vw := vdisp.VirtualW
	vdisp_vh := vdisp.VirtualH

	newVlayer := sVlayer.Dup()

	vlayer_vdx := sVlayer.VdstX
	vlayer_vdy := sVlayer.VdstY
	vlayer_vdw := sVlayer.VdstW
	vlayer_vdh := sVlayer.VdstH

	vlayer_vsx := sVlayer.VsrcX
	vlayer_vsy := sVlayer.VsrcY
	vlayer_vsw := sVlayer.VsrcW
	vlayer_vsh := sVlayer.VsrcH

	nvlayer_vdx, nvlayer_vdw, nvlayer_vsx, nvlayer_vsw :=
		convGlobalToVDisplayCoordinateSub(vdisp_vx, vdisp_vw, vlayer_vdx, vlayer_vdw, vlayer_vsx, vlayer_vsw)
	nvlayer_vdy, nvlayer_vdh, nvlayer_vsy, nvlayer_vsh :=
		convGlobalToVDisplayCoordinateSub(vdisp_vy, vdisp_vh, vlayer_vdy, vlayer_vdh, vlayer_vsy, vlayer_vsh)

	newVlayer.VdstX = nvlayer_vdx
	newVlayer.VdstW = nvlayer_vdw
	newVlayer.VsrcX = nvlayer_vsx
	newVlayer.VsrcW = nvlayer_vsw

	newVlayer.VdstY = nvlayer_vdy
	newVlayer.VsrcY = nvlayer_vsy
	newVlayer.VdstH = nvlayer_vdh
	newVlayer.VsrcH = nvlayer_vsh

	return newVlayer
}

func convToVDisplayCoordinate(sVlayers []ula.VirtualLayer,
	vdisp *ula.VirtualDisplay) []ula.VirtualLayer {

	dVlayers := make([]ula.VirtualLayer, 0)

	for _, vlayer := range sVlayers {

		var newVlayer ula.VirtualLayer
		if vlayer.Coord == ula.COORD_VDISPLAY {
			newVlayer = *vlayer.Dup()
		} else {
			newVlayer = *convGlobalToVDisplayCoordinate(&vlayer, vdisp)
		}

		dVlayers = append(dVlayers, newVlayer)
	}

	return dVlayers
}

func sAreaConvGlobalToVDisplayCoordinateSub(
	vdisp_vx int,
	vdisp_vw int,
	vlayer_vdx int,
	vlayer_vdw int) (int, int) {

	var (
		nvlayer_vdx int
		nvlayer_vdw int
	)

	/* Assume src cutout area is the entire layer in this part */
	if vdisp_vx <= vlayer_vdx &&
		vlayer_vdx <= vdisp_vx+vdisp_vw &&
		vdisp_vx+vdisp_vw <= vlayer_vdx+vlayer_vdw {
		nvlayer_vdx = vlayer_vdx - vdisp_vx
		nvlayer_vdw = vdisp_vx + vdisp_vw - vlayer_vdx
	} else if vlayer_vdx <= vdisp_vx &&
		vdisp_vx+vdisp_vw <= vlayer_vdx+vlayer_vdw {
		nvlayer_vdx = 0
		nvlayer_vdw = vdisp_vw
	} else if vlayer_vdx <= vdisp_vx &&
		vdisp_vx <= vlayer_vdx+vlayer_vdw &&
		vlayer_vdx+vlayer_vdw <= vdisp_vx+vdisp_vw {
		nvlayer_vdx = 0
		nvlayer_vdw = vlayer_vdx + vlayer_vdw - vdisp_vx
	} else if vdisp_vx <= vlayer_vdx &&
		vlayer_vdx+vlayer_vdw <= vdisp_vx+vdisp_vw {
		nvlayer_vdx = vlayer_vdx - vdisp_vx
		nvlayer_vdw = vlayer_vdw
	} else {
		nvlayer_vdx = 0
		nvlayer_vdw = 0
	}
	return nvlayer_vdx, nvlayer_vdw
}

func sAreaConvGlobalToVDisplayCoordinate(sVsafetyArea *ula.VirtualSafetyArea,
	vdisp *ula.VirtualDisplay) *ula.VirtualSafetyArea {

	vdisp_vx := vdisp.VirtualX
	vdisp_vy := vdisp.VirtualY
	vdisp_vw := vdisp.VirtualW
	vdisp_vh := vdisp.VirtualH

	newVsafetyarea := sVsafetyArea.Dup()

	vsafetyarea_vdx := sVsafetyArea.VirtualX
	vsafetyarea_vdy := sVsafetyArea.VirtualY
	vsafetyarea_vdw := sVsafetyArea.VirtualW
	vsafetyarea_vdh := sVsafetyArea.VirtualH

	nvlayer_vdx, nvlayer_vdw :=
		sAreaConvGlobalToVDisplayCoordinateSub(vdisp_vx, vdisp_vw, vsafetyarea_vdx, vsafetyarea_vdw)

	nvlayer_vdy, nvlayer_vdh :=
		sAreaConvGlobalToVDisplayCoordinateSub(vdisp_vy, vdisp_vh, vsafetyarea_vdy, vsafetyarea_vdh)

	newVsafetyarea.VirtualX = nvlayer_vdx
	newVsafetyarea.VirtualY = nvlayer_vdy
	newVsafetyarea.VirtualW = nvlayer_vdw
	newVsafetyarea.VirtualH = nvlayer_vdh

	return newVsafetyarea
}

func sAreaConvToVDisplayCoordinate(sVSafetyAreas []ula.VirtualSafetyArea,
	vdisp *ula.VirtualDisplay) []ula.VirtualSafetyArea {

	dVSafetyAreas := make([]ula.VirtualSafetyArea, 0)

	for _, vSafetyArea := range sVSafetyAreas {

		var newVSafetyArea ula.VirtualSafetyArea
		newVSafetyArea = *sAreaConvGlobalToVDisplayCoordinate(&vSafetyArea, vdisp)

		dVSafetyAreas = append(dVSafetyAreas, newVSafetyArea)
	}

	return dVSafetyAreas
}

func sAreaConvToRDisplayCoordinate(sVSafetyAreas []ula.VirtualSafetyArea,
	vdisp *ula.VirtualDisplay, rdisp *ula.RealDisplay) []ula.VirtualSafetyArea {

	dVSafetyAreas := make([]ula.VirtualSafetyArea, 0)

	vdisp_vw := vdisp.VirtualW
	vdisp_vh := vdisp.VirtualH

	rdisp_pixw := rdisp.PixelW
	rdisp_pixh := rdisp.PixelH

	for _, vSafetyArea := range sVSafetyAreas {

		newVSafetyArea := vSafetyArea.Dup()

		vSafetyArea_vdx := vSafetyArea.VirtualX
		vSafetyArea_vdy := vSafetyArea.VirtualY
		vSafetyArea_vdw := vSafetyArea.VirtualW
		vSafetyArea_vdh := vSafetyArea.VirtualH

		/* Convert RealDisplay Geometory */
		newVSafetyArea.VirtualX = vSafetyArea_vdx * rdisp_pixw / vdisp_vw
		newVSafetyArea.VirtualW = vSafetyArea_vdw * rdisp_pixw / vdisp_vw
		newVSafetyArea.VirtualY = vSafetyArea_vdy * rdisp_pixh / vdisp_vh
		newVSafetyArea.VirtualH = vSafetyArea_vdh * rdisp_pixh / vdisp_vh

		dVSafetyAreas = append(dVSafetyAreas, *newVSafetyArea)
	}

	return dVSafetyAreas
}

func convToRDisplayCoordinate(sVlayers []ula.VirtualLayer,
	vdisp *ula.VirtualDisplay, rdisp *ula.RealDisplay) []ula.VirtualLayer {

	dVlayers := make([]ula.VirtualLayer, 0)

	vdisp_vw := vdisp.VirtualW
	vdisp_vh := vdisp.VirtualH

	rdisp_pixw := rdisp.PixelW
	rdisp_pixh := rdisp.PixelH

	for _, vlayer := range sVlayers {

		newVlayer := vlayer.Dup()

		vlayer_vdx := vlayer.VdstX
		vlayer_vdy := vlayer.VdstY
		vlayer_vdw := vlayer.VdstW
		vlayer_vdh := vlayer.VdstH

		/* Convert RealDisplay Geometory */
		newVlayer.VdstX = vlayer_vdx * rdisp_pixw / vdisp_vw
		newVlayer.VdstW = vlayer_vdw * rdisp_pixw / vdisp_vw
		newVlayer.VdstY = vlayer_vdy * rdisp_pixh / vdisp_vh
		newVlayer.VdstH = vlayer_vdh * rdisp_pixh / vdisp_vh

		dVlayers = append(dVlayers, *newVlayer)
	}

	return dVlayers
}

func convVDisplay2RDisplayCoordinate(workV2RMap map[int]WorkV2R) {

	for key, workV2R := range workV2RMap {
		/*layer*/
		tmpVlayers := convToVDisplayCoordinate(workV2R.vlayers, &workV2R.vdisplay)
		vlayers := convToRDisplayCoordinate(tmpVlayers, &workV2R.vdisplay, &workV2R.rdisplay)
		workV2R.players = convVLayers2PLayers(vlayers)

		/*SafetyArea*/
		tmpVSAreas := sAreaConvToVDisplayCoordinate(workV2R.vsafetyareas, &workV2R.vdisplay)
		vSAreas := sAreaConvToRDisplayCoordinate(tmpVSAreas, &workV2R.vdisplay, &workV2R.rdisplay)
		workV2R.psafetyareas = convVSafetyAreas2PSafetyAreas(vSAreas)

		workV2RMap[key] = workV2R
	}
}

func (vs2rd *Vscreen2RdisplayConverter) GetNodePixelScreens() (*ula.NodePixelScreens, error) {

	pscrns := make([]ula.PixelScreen, 0)

	for _, workV2R := range vs2rd.workV2RMap {

		pscrn, err := ula.NewPixelScreen(&workV2R.rdisplay, workV2R.players, workV2R.psafetyareas)
		if err != nil {
			return nil, err
		}
		pscrns = append(pscrns, *pscrn)
	}

	spscrns, err := ula.NewNodePixelScreens(vs2rd.nodeId, pscrns)
	if err != nil {
		return nil, err
	}

	return spscrns, nil
}
