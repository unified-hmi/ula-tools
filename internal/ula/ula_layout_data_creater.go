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

package ula

func (vdsp *VirtualDisplay) Dup() *VirtualDisplay {
	copied := *vdsp
	return &copied
}

func (sVsurf *VirtualSurface) Dup() *VirtualSurface {
	copied := *sVsurf
	return &copied
}

func (sVlayer *VirtualLayer) Dup() *VirtualLayer {

	copiedVLayer := *sVlayer
	copiedVLayer.Vsurfaces = make([]VirtualSurface, 0)

	for _, sVsurf := range sVlayer.Vsurfaces {
		copiedVLayer.Vsurfaces = append(copiedVLayer.Vsurfaces, *sVsurf.Dup())
	}

	return &copiedVLayer
}

func (vsaftyarea *VirtualSafetyArea) Dup() *VirtualSafetyArea {
	copied := *vsaftyarea
	return &copied
}

func (sPSafetyArea *PixelSafetyArea) Dup() *PixelSafetyArea {

	copiedPSafetyArea := *sPSafetyArea
	return &copiedPSafetyArea
}

func (rdsp *RealDisplay) Dup() *RealDisplay {
	copied := *rdsp
	return &copied
}

func (sPsurf *PixelSurface) Dup() *PixelSurface {
	copied := *sPsurf
	return &copied
}

func (sPlayer *PixelLayer) Dup() *PixelLayer {

	copiedPLayer := *sPlayer
	copiedPLayer.Psurfaces = make([]PixelSurface, 0)

	for _, sPsurf := range sPlayer.Psurfaces {
		copiedPLayer.Psurfaces = append(copiedPLayer.Psurfaces, *sPsurf.Dup())
	}

	return &copiedPLayer
}

func (pscrn *PixelScreen) Dup() *PixelScreen {

	copiedPscrn := *pscrn
	copiedPscrn.Players = DupPixelLayerSlice(pscrn.Players)

	return &copiedPscrn
}

func (spscrn *NodePixelScreens) Dup() *NodePixelScreens {

	copiedSPscrn := *spscrn
	copiedSPscrn.Pscreens = DupPixelScreensSlice(spscrn.Pscreens)

	return &copiedSPscrn
}

func DupVirtualLayerSlice(srcVlayers []VirtualLayer) []VirtualLayer {

	dstVlayers := make([]VirtualLayer, 0)

	for _, sVlayer := range srcVlayers {
		dstVlayers = append(dstVlayers, *sVlayer.Dup())
	}

	return dstVlayers
}

func DupVirtualLayerSliceIfNeed(srcVlayers []VirtualLayer,
	isNeedFunc func(vlayer *VirtualLayer, arg interface{}) bool,
	arg interface{}) []VirtualLayer {

	dstVlayers := make([]VirtualLayer, 0)

	for _, sVlayer := range srcVlayers {

		isNeed := isNeedFunc(&sVlayer, arg)
		if !isNeed {
			continue
		}

		dstVlayers = append(dstVlayers, *sVlayer.Dup())

	}

	return dstVlayers
}

func NewEmptyPixelSurface() *PixelSurface {

	psurf := PixelSurface{}

	return &psurf
}

func NewEmptyPixelSafetyArea() *PixelSafetyArea {

	psafetyarea := PixelSafetyArea{}

	return &psafetyarea
}

func DupPixelSafetyAreaSlice(srcPSafetyAreas []PixelSafetyArea) []PixelSafetyArea {

	dstPSafetyAreas := make([]PixelSafetyArea, 0)

	for _, sPSafetyArea := range srcPSafetyAreas {
		dstPSafetyAreas = append(dstPSafetyAreas, *sPSafetyArea.Dup())
	}

	return dstPSafetyAreas
}

func DupVirtualSafetyAreaSlice(srcVSafetyAreas []VirtualSafetyArea) []VirtualSafetyArea {

	dstVSafetyAreas := make([]VirtualSafetyArea, 0)

	for _, sVSafetyArea := range srcVSafetyAreas {
		dstVSafetyAreas = append(dstVSafetyAreas, *sVSafetyArea.Dup())
	}

	return dstVSafetyAreas
}

func NewEmptyPixelLayer() *PixelLayer {

	player := PixelLayer{
		Psurfaces: make([]PixelSurface, 0),
	}

	return &player
}

func NewPixelScreen(rdsp *RealDisplay, players []PixelLayer, psafetyareas []PixelSafetyArea) (*PixelScreen, error) {

	pscrn := PixelScreen{
		Rdisplay:     *rdsp,
		Players:      DupPixelLayerSlice(players),
		PsafetyAreas: DupPixelSafetyAreaSlice(psafetyareas),
	}

	return &pscrn, nil
}

func NewNodePixelScreens(nodeId int, pscrns []PixelScreen) (*NodePixelScreens, error) {

	spscrns := NodePixelScreens{
		NodeId:   nodeId,
		Pscreens: DupPixelScreensSlice(pscrns),
	}

	return &spscrns, nil
}

func DupPixelLayerSlice(srcPlayers []PixelLayer) []PixelLayer {

	dstPlayers := make([]PixelLayer, 0)

	for _, sPlayer := range srcPlayers {
		dstPlayers = append(dstPlayers, *sPlayer.Dup())
	}

	return dstPlayers
}

func DupPixelScreensSlice(sPscrns []PixelScreen) []PixelScreen {

	dPscrns := make([]PixelScreen, 0)

	for _, pscrn := range sPscrns {
		dPscrns = append(dPscrns, *pscrn.Dup())
	}

	return dPscrns
}
