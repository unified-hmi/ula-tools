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

const VSCRNDEF_FILE = "/etc/uhmi-framework/virtual-screen-def.json"

type Coord int

const (
	COORD_GLOBAL Coord = iota
	COORD_VDISPLAY
)

type VirtualDisplay struct {
	DispName   string
	VDisplayId int
	VirtualX   int
	VirtualY   int
	VirtualW   int
	VirtualH   int
}

type VirtualSurface struct {
	AppName string

	ParentVID int
	VID       int

	PixelW int
	PixelH int

	PsrcX int
	PsrcY int
	PsrcW int
	PsrcH int

	VdstX int
	VdstY int
	VdstW int
	VdstH int

	Visibility int
}

type VirtualLayer struct {
	AppName string

	VID int

	Coord      Coord
	VDisplayId int /* only used if COORD is COORD_VDISPLAY */

	VirtualW int
	VirtualH int

	VsrcX int
	VsrcY int
	VsrcW int
	VsrcH int

	VdstX int
	VdstY int
	VdstW int
	VdstH int

	Visibility int

	Vsurfaces []VirtualSurface
}

type VirtualSafetyArea struct {
	VirtualX int
	VirtualY int
	VirtualW int
	VirtualH int
}
