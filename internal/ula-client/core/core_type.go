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

package core

type CAVsurface struct {
	AppName    string /* who is it created by */
	VID        int
	AppID      string
	PixelW     int
	PixelH     int
	PsrcX      int
	PsrcY      int
	PsrcW      int
	PsrcH      int
	VdstX      int
	VdstY      int
	VdstW      int
	VdstH      int
	Visibility *int
}

type CAVlayer struct {
	AppName    string /* who is it created by */
	VID        int
	ZOrder     int
	Coord      string
	VdisplayId int
	VirtualW   int
	VirtualH   int
	VsrcX      int
	VsrcY      int
	VsrcW      int
	VsrcH      int
	VdstX      int
	VdstY      int
	VdstW      int
	VdstH      int
	Visibility *int
	Vsurfaces  []CAVsurface
}

type CALayout struct {
	Vlayers []CAVlayer
}

type CALayoutTree struct {
	Vlayers []CAVlayer
}

func (vsurf *CAVsurface) Dup() *CAVsurface {
	copied := *vsurf
	return &copied
}

func (vlayer *CAVlayer) Dup() *CAVlayer {

	copiedVLayer := *vlayer
	copiedVLayer.Vsurfaces = make([]CAVsurface, 0)

	for _, vsurf := range vlayer.Vsurfaces {
		copiedVLayer.Vsurfaces = append(copiedVLayer.Vsurfaces, *vsurf.Dup())
	}

	return &copiedVLayer
}
