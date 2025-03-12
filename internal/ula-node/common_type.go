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

package ulanode

import (
	"ula-tools/internal/ula"
)

type IdPair struct {
	LayerId   int
	SurfaceId int // should be -1 if SurfaceId is not used
}

type ApplyCommandData struct {
	Command string
	ChgIds  []IdPair
	Vlayers []ula.VirtualLayer
}

func DupVirtualLayerSlice(srcVlayers []ula.VirtualLayer) []ula.VirtualLayer {

	dstVlayers := make([]ula.VirtualLayer, 0)

	for _, sVlayer := range srcVlayers {
		dstVlayers = append(dstVlayers, *sVlayer.Dup())
	}

	return dstVlayers
}

func DupVirtualLayerSliceIfNeed(srcVlayers []ula.VirtualLayer,
	isNeedFunc func(vlayer *ula.VirtualLayer, arg interface{}) bool,
	arg interface{}) []ula.VirtualLayer {

	dstVlayers := make([]ula.VirtualLayer, 0)

	for _, sVlayer := range srcVlayers {

		isNeed := isNeedFunc(&sVlayer, arg)
		if !isNeed {
			continue
		}

		dstVlayers = append(dstVlayers, *sVlayer.Dup())

	}

	return dstVlayers
}

type RealDisplay struct {
	NodeId     int
	PixelW     int
	PixelH     int
	VDisplayId int
	RDisplayId int
}

func (rdsp *RealDisplay) Dup() *RealDisplay {
	copied := *rdsp
	return &copied
}

type PixelSurface struct {
	ParentVID int
	VID       int
	AppID     string

	PixelW int
	PixelH int

	PsrcX int
	PsrcY int
	PsrcW int
	PsrcH int

	PdstX int
	PdstY int
	PdstW int
	PdstH int

	Visibility int
}

func NewEmptyPixelSurface() *PixelSurface {

	psurf := PixelSurface{}

	return &psurf
}

func (sPsurf *PixelSurface) Dup() *PixelSurface {
	copied := *sPsurf
	return &copied
}

func DupPixelSurfaceSlice(srcPsurfaces []PixelSurface) []PixelSurface {

	dstPsurfaces := make([]PixelSurface, 0)

	for _, sPsurface := range srcPsurfaces {
		dstPsurfaces = append(dstPsurfaces, *sPsurface.Dup())
	}

	return dstPsurfaces
}

type PixelLayer struct {
	VID    int
	PixelW int
	PixelH int

	PsrcX int
	PsrcY int
	PsrcW int
	PsrcH int

	PdstX int
	PdstY int
	PdstW int
	PdstH int

	Visibility int

	Psurfaces []PixelSurface
}

func NewEmptyPixelLayer() *PixelLayer {

	player := PixelLayer{
		Psurfaces: make([]PixelSurface, 0),
	}

	return &player
}

func (sPlayer *PixelLayer) Dup() *PixelLayer {

	copiedPLayer := *sPlayer
	copiedPLayer.Psurfaces = make([]PixelSurface, 0)

	for _, sPsurf := range sPlayer.Psurfaces {
		copiedPLayer.Psurfaces = append(copiedPLayer.Psurfaces, *sPsurf.Dup())
	}

	return &copiedPLayer
}

func (sPlayer *PixelLayer) DupWithoutSurface() *PixelLayer {

	copiedPLayer := *sPlayer
	copiedPLayer.Psurfaces = make([]PixelSurface, 0)

	return &copiedPLayer
}

func DupPixelLayerSlice(srcPlayers []PixelLayer) []PixelLayer {

	dstPlayers := make([]PixelLayer, 0)

	for _, sPlayer := range srcPlayers {
		dstPlayers = append(dstPlayers, *sPlayer.Dup())
	}

	return dstPlayers
}

func DupPixelLayerSliceWithoutSurface(srcPlayers []PixelLayer) []PixelLayer {

	dstPlayers := make([]PixelLayer, 0)

	for _, sPlayer := range srcPlayers {
		dstPlayers = append(dstPlayers, *sPlayer.DupWithoutSurface())
	}

	return dstPlayers
}

type PixelScreen struct {
	Rdisplay RealDisplay
	Players  []PixelLayer
}

func NewPixelScreen(rdsp *RealDisplay, players []PixelLayer) (*PixelScreen, error) {

	pscrn := PixelScreen{
		Rdisplay: *rdsp,
		Players:  DupPixelLayerSlice(players),
	}

	return &pscrn, nil
}

func (pscrn *PixelScreen) Dup() *PixelScreen {

	copiedPscrn := *pscrn
	copiedPscrn.Players = DupPixelLayerSlice(pscrn.Players)

	return &copiedPscrn
}

func dupPixelScreensSlice(sPscrns []PixelScreen) []PixelScreen {

	dPscrns := make([]PixelScreen, 0)

	for _, pscrn := range sPscrns {
		dPscrns = append(dPscrns, *pscrn.Dup())
	}

	return dPscrns
}

type NodePixelScreens struct {
	NodeId   int
	Pscreens []PixelScreen
}

func NewNodePixelScreens(nodeId int, pscrns []PixelScreen) (*NodePixelScreens, error) {

	spscrns := NodePixelScreens{
		NodeId:   nodeId,
		Pscreens: dupPixelScreensSlice(pscrns),
	}

	return &spscrns, nil
}

type RdisplayCommandData struct {
	Rdisplay    RealDisplay
	InsertOrder string
	ReferenceId int
	Players     []PixelLayer
}

func NewRdisplayCommandData(rdisp *RealDisplay, players []PixelLayer) (*RdisplayCommandData, error) {

	dcomm := RdisplayCommandData{
		Rdisplay: *rdisp.Dup(),
		Players:  DupPixelLayerSlice(players),
	}

	return &dcomm, nil
}

func (spscrn *NodePixelScreens) Dup() *NodePixelScreens {

	copiedSPscrn := *spscrn
	copiedSPscrn.Pscreens = dupPixelScreensSlice(spscrn.Pscreens)

	return &copiedSPscrn
}

type LocalCommandReq struct {
	Command string
	RDComms []RdisplayCommandData
	Ret     int
}

func NewEmptyLocalCommandReq() (*LocalCommandReq, error) {

	ltq := LocalCommandReq{}

	return &ltq, nil
}

type GeometoryConverter interface {
	DoConvert() error
	GetNodePixelScreens() (*NodePixelScreens, error)
}

type LocalCommandGenerator interface {
	GenerateLocalCommandReq(*ApplyCommandData) ([]*LocalCommandReq, error)
}
