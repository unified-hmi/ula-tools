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

package ulacommgen

/*  Ula Protocol for Initial VSurface */
type UPIVsurface struct {
	VID        int    `json:"VID"`
	AppID      string `json:"AppID"`
	PixelW     int    `json:"pixel_w"`
	PixelH     int    `json:"pixel_h"`
	PsrcX      int    `json:"psrc_x"`
	PsrcY      int    `json:"psrc_y"`
	PsrcW      int    `json:"psrc_w"`
	PsrcH      int    `json:"psrc_h"`
	VdstX      int    `json:"vdst_x"`
	VdstY      int    `json:"vdst_y"`
	VdstW      int    `json:"vdst_w"`
	VdstH      int    `json:"vdst_h"`
	Visibility *int   `json:"visibility"`
}

/*  Ula Protocol for Initial Vlayer */
type UPIVlayer struct {
	VID        int           `json:"VID"`
	Coord      string        `json:"coord"`
	VdisplayId int           `json:"vdisplay_id"`
	VirtualW   int           `json:"virtual_w"`
	VirtualH   int           `json:"virtual_h"`
	VsrcX      int           `json:"vsrc_x"`
	VsrcY      int           `json:"vsrc_y"`
	VsrcW      int           `json:"vsrc_w"`
	VsrcH      int           `json:"vsrc_h"`
	VdstX      int           `json:"vdst_x"`
	VdstY      int           `json:"vdst_y"`
	VdstW      int           `json:"vdst_w"`
	VdstH      int           `json:"vdst_h"`
	Visibility *int          `json:"visibility"`
	Surface    []UPIVsurface `json:"vsurface"`
}

/*  Ula Protocol for Initial VScreen */
type UPIVscreen struct {
	Command string      `json:"command"`
	Layer   []UPIVlayer `json:"vlayer"`
}
