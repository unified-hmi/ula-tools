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
	"encoding/json"
	"errors"
	_ "fmt"
	"sync"
	"ula-tools/internal/ula"
	. "ula-tools/internal/ulog"
)

var vScreenMutex sync.RWMutex
var VScreen *VirtualScreen
var NodePixelScreens *ula.NodePixelScreens

func getStringFromJson(mJson map[string]interface{}, key string) (string, error) {

	tval := mJson[key]
	if tval == nil {
		return "", errors.New("Error in getStringFromJson")
	}
	val := tval.(string)

	return val, nil
}

func getIntFromJson(mJson map[string]interface{}, key string) (int, error) {

	tval := mJson[key]
	if tval == nil {
		return 0, errors.New("Error in getIntFromJson")
	}
	val := int(tval.(float64))

	return val, nil
}

func getIntFromJsonDef(mJson map[string]interface{}, key string, defval int) (int, error) {

	tval := mJson[key]
	if tval == nil {
		return defval, nil
	}
	val := int(tval.(float64))

	return val, nil
}

func getSliceFromJson(mJson map[string]interface{}, key string) ([]interface{}, error) {

	tval := mJson[key]
	if tval == nil {
		return make([]interface{}, 0), errors.New("Error in getSliceFromJson")
	}
	val := tval.([]interface{})

	return val, nil
}

func getCoordFromJsonDef(mJson map[string]interface{}, key string, defval ula.Coord) (ula.Coord, error) {

	tval := mJson[key]
	if tval == nil {
		return defval, nil
	}

	if tval == "global" {
		defval = ula.COORD_GLOBAL
	} else if tval == "vdisplay" {
		defval = ula.COORD_VDISPLAY
	} else {
		return 0, errors.New("getCoordFromJsonDef")
	}

	return defval, nil
}

func generateSurfaceFromParam(layerId int, mSurface map[string]interface{}, appli_name string) (*ula.VirtualSurface, error) {

	surfaceId, err := getIntFromJson(mSurface, "VID")
	if err != nil {
		return nil, err
	}

	pixelW, err := getIntFromJson(mSurface, "pixel_w")
	if err != nil {
		return nil, err
	}

	pixelH, err := getIntFromJson(mSurface, "pixel_h")
	if err != nil {
		return nil, err
	}

	psrcX, err := getIntFromJson(mSurface, "psrc_x")
	if err != nil {
		return nil, err
	}

	psrcY, err := getIntFromJson(mSurface, "psrc_y")
	if err != nil {
		return nil, err
	}

	psrcW, err := getIntFromJson(mSurface, "psrc_w")
	if err != nil {
		return nil, err
	}

	psrcH, err := getIntFromJson(mSurface, "psrc_h")
	if err != nil {
		return nil, err
	}

	vdstX, err := getIntFromJson(mSurface, "vdst_x")
	if err != nil {
		return nil, err
	}

	vdstY, err := getIntFromJson(mSurface, "vdst_y")
	if err != nil {
		return nil, err
	}

	vdstW, err := getIntFromJson(mSurface, "vdst_w")
	if err != nil {
		return nil, err
	}

	vdstH, err := getIntFromJson(mSurface, "vdst_h")
	if err != nil {
		return nil, err
	}

	if pixelW < 0 || pixelH < 0 {
		return nil, errors.New("pixel_w and pixel_h should not have negative values")
	}

	if psrcX < 0 || psrcY < 0 || psrcW < 0 || psrcH < 0 {
		return nil, errors.New("psrc regions should not have negative values")
	}

	if vdstX < 0 || vdstY < 0 || vdstW < 0 || vdstH < 0 {
		return nil, errors.New("vdst regions should not have negative values")
	}

	visibility, err := getIntFromJsonDef(mSurface, "visibility", 1)
	if err != nil {
		return nil, err
	}

	vsurface := ula.VirtualSurface{
		AppName:    appli_name,
		ParentVID:  layerId,
		VID:        surfaceId,
		PixelW:     pixelW,
		PixelH:     pixelH,
		PsrcX:      psrcX,
		PsrcY:      psrcY,
		PsrcW:      psrcW,
		PsrcH:      psrcH,
		VdstX:      vdstX,
		VdstY:      vdstY,
		VdstW:      vdstW,
		VdstH:      vdstH,
		Visibility: visibility,
	}

	return &vsurface, nil
}

type VirtualScreen struct {
	VScrnDef ula.VScrnDef

	/* the same as vscrnDef.Size.VirtualW and vscrnDef.Size.VirtualH */
	VirtualWidth  int
	VirtualHeight int

	VirtualDisplays   map[int]ula.VirtualDisplay
	RealDisplays      map[int]ula.RealDisplay
	VdispVlayers      map[int][]ula.VirtualLayer
	VdispVsafetyAreas map[int][]ula.VirtualSafetyArea
}

func NewVirtualScreen(vscrnDef *ula.VScrnDef) (*VirtualScreen, error) {

	vscreen := VirtualScreen{
		VScrnDef:          *vscrnDef,
		VirtualWidth:      vscrnDef.Def2D.Size.VirtualW,
		VirtualHeight:     vscrnDef.Def2D.Size.VirtualH,
		VirtualDisplays:   make(map[int]ula.VirtualDisplay),
		RealDisplays:      make(map[int]ula.RealDisplay),
		VdispVlayers:      make(map[int][]ula.VirtualLayer),
		VdispVsafetyAreas: make(map[int][]ula.VirtualSafetyArea),
	}

	for _, r := range vscrnDef.Def2D.VirtualDisplays {
		vscreen.VirtualDisplays[r.VDisplayId] = ula.VirtualDisplay{
			DispName:   r.DispName,
			VDisplayId: r.VDisplayId,
			VirtualX:   r.VirtualX,
			VirtualY:   r.VirtualY,
			VirtualW:   r.VirtualW,
			VirtualH:   r.VirtualH,
		}
		vscreen.VdispVlayers[r.VDisplayId] = make([]ula.VirtualLayer, 0)
	}

	for _, r := range vscrnDef.RealDisplays {
		vscreen.RealDisplays[r.VDisplayId] = ula.RealDisplay{
			NodeId:     r.NodeId,
			VDisplayId: r.VDisplayId,
			PixelW:     r.PixelW,
			PixelH:     r.PixelH,
			RDisplayId: r.RDisplayId,
		}
	}

	vVdispVsafetyAreas := make([]ula.VirtualSafetyArea, 0)
	for _, r := range vscrnDef.VirtualSafetyArea {
		vVdispVsafetyArea := ula.VirtualSafetyArea{
			VirtualX: r.VirtualX,
			VirtualY: r.VirtualY,
			VirtualW: r.VirtualW,
			VirtualH: r.VirtualH,
		}
		vVdispVsafetyAreas = append(vVdispVsafetyAreas, vVdispVsafetyArea)
	}
	for _, r := range vscrnDef.Def2D.VirtualDisplays {
		vscreen.VdispVsafetyAreas[r.VDisplayId] = vVdispVsafetyAreas
	}

	return &vscreen, nil
}

func (vscreen *VirtualScreen) Dup() *VirtualScreen {
	vScreenMutex.RLock()
	defer vScreenMutex.RUnlock()
	copiedVDsps := make(map[int]ula.VirtualDisplay)
	copiedRDsps := make(map[int]ula.RealDisplay)
	copiedVDispVLayers := make(map[int][]ula.VirtualLayer)
	copiedVdispVsafetyAreas := make(map[int][]ula.VirtualSafetyArea)

	for vdspid, vdisplay := range vscreen.VirtualDisplays {
		copiedVDsp := vdisplay
		copiedRDsp := vscreen.RealDisplays[vdspid]
		copiedVLayers := vscreen.VdispVlayers[vdspid]
		copiedVSafetyAreas := vscreen.VdispVsafetyAreas[vdspid]

		copiedVDsps[vdspid] = copiedVDsp
		copiedRDsps[vdspid] = copiedRDsp
		copiedVDispVLayers[vdspid] = ula.DupVirtualLayerSlice(copiedVLayers)
		copiedVdispVsafetyAreas[vdspid] = ula.DupVirtualSafetyAreaSlice(copiedVSafetyAreas)
	}

	copiedVscreen := *vscreen
	copiedVscreen.VirtualDisplays = copiedVDsps
	copiedVscreen.RealDisplays = copiedRDsps
	copiedVscreen.VdispVlayers = copiedVDispVLayers
	copiedVscreen.VdispVsafetyAreas = copiedVdispVsafetyAreas
	return &copiedVscreen
}

func (vscrn *VirtualScreen) ApplyCommand(mJson map[string]interface{}) (*ula.ApplyCommandData, error) {
	vScreenMutex.RLock()
	defer vScreenMutex.RUnlock()
	command := mJson["command"].(string)
	DLog.Println("command=", command)

	var chgIds []ula.IdPair
	var err error

	switch command {
	case "initial_vscreen":
		DLog.Println("@@INITIAL_VSCREEN@@")
		chgIds, err = initVirtualScreen(vscrn, mJson)
	default:
		chgIds = make([]ula.IdPair, 0)
	}

	if err != nil {
		return nil, err
	}

	var vids []int
	for _, vdisp := range vscrn.VirtualDisplays {
		vids = make([]int, 0)
		for _, vlayer := range vscrn.VdispVlayers[vdisp.VDisplayId] {
			vids = append(vids, vlayer.VID)
		}
		DLog.Println("ApplyCommand command: ", command, " vdispId: ", vdisp.VDisplayId, " vids: ", vids)
	}

	acdata := &ula.ApplyCommandData{Command: command, ChgIds: chgIds}

	return acdata, nil

}

func initVirtualScreen(vscreen *VirtualScreen, mJson map[string]interface{}) ([]ula.IdPair, error) {
	err := fillVscreenFromParam(vscreen, mJson)
	if err != nil {
		return make([]ula.IdPair, 0), err
	}

	return make([]ula.IdPair, 0), nil
}

func ApplyAndGenCommand(command string, nodeId int) (string, error) {
	var applyCommand map[string]interface{}
	if err := json.Unmarshal([]byte(command), &applyCommand); err != nil {
		ELog.Printf("Unmarshal json command error: %s\n", err)
		return "", err
	}

	vscrnCopy := VScreen.Dup()

	acdata, err := vscrnCopy.ApplyCommand(applyCommand)
	if err != nil {
		ELog.Printf("ApplyCommand error: %s\n", err)
		return "", err
	}

	vs2rdConv, err := NewVscreen2RdisplayConverter(vscrnCopy, nodeId)
	if err != nil {
		ELog.Printf("Failed to create converter: %s\n", err)
		return "", err
	}

	var vsconv GeometoryConverter = vs2rdConv
	vsconv.DoConvert()

	acdata.NPScreens, err = vsconv.GetNodePixelScreens()
	if err != nil {
		ELog.Printf("GetNodePixelScreens error: %s\n", err)
		return "", err
	}

	jsonBytes, err := json.Marshal(acdata)
	if err != nil {
		ELog.Printf("Marshal ApplyCommandData error: %s\n", err)
		return "", err
	}

	jsonCommand := string(jsonBytes)

	vScreenMutex.Lock()
	VScreen = vscrnCopy
	vScreenMutex.Unlock()

	return jsonCommand, nil
}

func generateLayerFromParam(mLayer map[string]interface{}, genSurfaces bool, existingVlayer *ula.VirtualLayer) (*ula.VirtualLayer, error) {

	appli_name, err := getStringFromJson(mLayer, "appli_name")
	if err != nil {
		ELog.Println("error in generateLayerFromParam")
		return nil, err
	}

	layerId, err := getIntFromJson(mLayer, "VID")
	if err != nil {
		if existingVlayer == nil {
			return nil, err
		} else {
			layerId = existingVlayer.VID
		}
	}

	var defCoord ula.Coord
	if existingVlayer != nil {
		defCoord = existingVlayer.Coord
	} else {
		defCoord = ula.COORD_GLOBAL
	}
	coord, err := getCoordFromJsonDef(mLayer, "coord", defCoord)
	if err != nil {
		ELog.Println("error in generateLayerFromParam")
		return nil, err
	}

	vdisplayId := -1
	if coord == ula.COORD_VDISPLAY {
		defVdisplayId := -1
		if existingVlayer != nil {
			defVdisplayId = existingVlayer.VDisplayId
		}
		vdisplayId, err = getIntFromJsonDef(mLayer, "vdisplay_id", defVdisplayId)
		if err != nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		}
	}

	virtualW, err := getIntFromJson(mLayer, "virtual_w")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			virtualW = existingVlayer.VirtualW
		}
	}

	virtualH, err := getIntFromJson(mLayer, "virtual_h")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			virtualH = existingVlayer.VirtualH
		}
	}

	vsrcX, err := getIntFromJson(mLayer, "vsrc_x")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vsrcX = existingVlayer.VsrcX
		}
	}

	vsrcY, err := getIntFromJson(mLayer, "vsrc_y")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vsrcY = existingVlayer.VsrcY
		}
	}

	vsrcW, err := getIntFromJson(mLayer, "vsrc_w")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vsrcW = existingVlayer.VsrcW
		}
	}

	vsrcH, err := getIntFromJson(mLayer, "vsrc_h")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vsrcH = existingVlayer.VsrcH
		}
	}

	vdstX, err := getIntFromJson(mLayer, "vdst_x")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vdstX = existingVlayer.VdstX
		}
	}

	vdstY, err := getIntFromJson(mLayer, "vdst_y")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vdstY = existingVlayer.VdstY
		}
	}

	vdstW, err := getIntFromJson(mLayer, "vdst_w")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vdstW = existingVlayer.VdstW
		}
	}

	vdstH, err := getIntFromJson(mLayer, "vdst_h")
	if err != nil {
		if existingVlayer == nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		} else {
			vdstH = existingVlayer.VdstH
		}
	}

	if virtualW < 0 || virtualH < 0 {
		return nil, errors.New("virtual_w and virtual_h should not have negative values")
	}

	if vsrcX < 0 || vsrcY < 0 || vsrcW < 0 || vsrcH < 0 {
		return nil, errors.New("vsrc regions should not have negative values")
	}

	if vdstX < 0 || vdstY < 0 || vdstW < 0 || vdstH < 0 {
		return nil, errors.New("vdst regions should not have negative values")
	}

	if vdstW == 0 || vdstH == 0 {
		return nil, errors.New("vdst_w and vdst_h need to have values other than 0")
	}

	defVisibility := 1
	if existingVlayer != nil {
		defVisibility = existingVlayer.Visibility
	}
	visibility, err := getIntFromJsonDef(mLayer, "visibility", defVisibility)
	if err != nil {
		ELog.Println("error in generateLayerFromParam")
		return nil, err
	}

	vsurfaces := make([]ula.VirtualSurface, 0)
	if genSurfaces {
		surfaces, err := getSliceFromJson(mLayer, "vsurface")
		if err != nil {
			ELog.Println("error in generateLayerFromParam")
			return nil, err
		}

		for _, mSurface := range surfaces {
			newVsurface, err := generateSurfaceFromParam(layerId, mSurface.(map[string]interface{}), appli_name)
			if err != nil {
				ELog.Println("error in generateLayerFromParam")
				return nil, err
			}
			vsurfaces = append(vsurfaces, *newVsurface)
		}
	}

	vlayer := ula.VirtualLayer{
		AppName:    appli_name,
		VID:        layerId,
		Coord:      coord,
		VDisplayId: vdisplayId,
		VirtualW:   virtualW,
		VirtualH:   virtualH,
		VsrcX:      vsrcX,
		VsrcY:      vsrcY,
		VsrcW:      vsrcW,
		VsrcH:      vsrcH,
		VdstX:      vdstX,
		VdstY:      vdstY,
		VdstW:      vdstW,
		VdstH:      vdstH,
		Visibility: visibility,
		Vsurfaces:  vsurfaces,
	}

	return &vlayer, nil
}

/* fill initial virtual layer, virtual window param*/
func fillVscreenFromParam(vscreen *VirtualScreen, mJson map[string]interface{}) error {

	vlayers := make([]ula.VirtualLayer, 0)

	layers, err := getSliceFromJson(mJson, "vlayer")
	if err != nil {
		ELog.Println("err in fillVscreenFromParam")
		return err
	}

	for _, mLayer := range layers {
		var existingVlayer *ula.VirtualLayer
		newVlayer, err := generateLayerFromParam(mLayer.(map[string]interface{}), true, existingVlayer)
		if err != nil {
			ELog.Println("err in fillVscreenFromParam")
			return err
		}
		vlayers = append(vlayers, *newVlayer)
	}

	for _, vdisp := range vscreen.VirtualDisplays {
		vscreen.VdispVlayers[vdisp.VDisplayId] = vlayers
	}

	return nil
}
