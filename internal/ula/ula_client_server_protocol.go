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

// should be -1 if SurfaceId is not used
type IdPair struct {
	LayerId   int `json:"LayerId"`
	SurfaceId int `json:"SurfaceId"`
}

type ApplyCommandData struct {
	Command   string            `json:"Command"`
	ChgIds    []IdPair          `json:"ChgIds"`
	NPScreens *NodePixelScreens `json:"NPScreens"`
}

type NodePixelScreens struct {
	NodeId   int           `json:"NodeId"`
	Pscreens []PixelScreen `json:"Pscreens"`
}

type PixelScreen struct {
	Rdisplay     RealDisplay       `json:"Rdisplay"`
	Players      []PixelLayer      `json:"Players"`
	PsafetyAreas []PixelSafetyArea `json:"PsafetyAreas"`
}

type RealDisplay struct {
	NodeId     int `json:"NodeId"`
	PixelW     int `json:"PixelW"`
	PixelH     int `json:"PixelH"`
	VDisplayId int `json:"VDisplayId"`
	RDisplayId int `json:"RDisplayId"`
}

type PixelLayer struct {
	AppName string `json:"AppName"`

	VID    int `json:"VID"`
	PixelW int `json:"PixelW"`
	PixelH int `json:"PixelH"`

	PsrcX int `json:"PsrcX"`
	PsrcY int `json:"PsrcY"`
	PsrcW int `json:"PsrcW"`
	PsrcH int `json:"PsrcH"`

	PdstX int `json:"PdstX"`
	PdstY int `json:"PdstY"`
	PdstW int `json:"PdstW"`
	PdstH int `json:"PdstH"`

	Visibility int `json:"Visibility"`

	Psurfaces []PixelSurface `json:"Psurfaces"`
}

type PixelSurface struct {
	AppName string `json:"AppName"`

	ParentVID int
	VID       int `json:"VID"`

	PixelW int `json:"PixelW"`
	PixelH int `json:"PixelH"`

	PsrcX int `json:"PsrcX"`
	PsrcY int `json:"PsrcY"`
	PsrcW int `json:"PsrcW"`
	PsrcH int `json:"PsrcH"`

	PdstX int `json:"PdstX"`
	PdstY int `json:"PdstY"`
	PdstW int `json:"PdstW"`
	PdstH int `json:"PdstH"`

	Visibility int `json:"Visibility"`
}

type PixelSafetyArea struct {
	PixelX int `json:"PixelX"`
	PixelY int `json:"PixelY"`
	PixelW int `json:"PixelW"`
	PixelH int `json:"PixelH"`
}
