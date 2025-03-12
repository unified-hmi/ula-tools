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

package readclusterapp

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path/filepath"
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-client/core"
	. "ula-tools/internal/ulog"
)

const (
	DEF_DWM_DIR   = "/var/local/uhmi-app/dwm"
	initialLayout = "dwm_initial_layout.json"
)

type caCfgInitialLayoutVsurface struct {
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

type caCfgInitialLayoutVLayer struct {
	VID        int                          `json:"VID"`
	Coord      *string                      `json:"coord"`
	VdisplayId *int                         `json:"vdisplay_id"` /* pointer type for check json property existence */
	Zorder     *int                         `json:"z_order"`     /* pointer type for check json property existence */
	VirtualW   int                          `json:"virtual_w"`
	VirtualH   int                          `json:"virtual_h"`
	VsrcX      int                          `json:"vsrc_x"`
	VsrcY      int                          `json:"vsrc_y"`
	VsrcW      int                          `json:"vsrc_w"`
	VsrcH      int                          `json:"vsrc_h"`
	VdstX      int                          `json:"vdst_x"`
	VdstY      int                          `json:"vdst_y"`
	VdstW      int                          `json:"vdst_w"`
	VdstH      int                          `json:"vdst_h"`
	Visibility *int                         `json:"visibility"`
	Vsurface   []caCfgInitialLayoutVsurface `json:"vsurface"`
}

type caCfgInitialLayout struct {
	AppName string                     `json:"application_name"`
	Vlayer  []caCfgInitialLayoutVLayer `json:"vlayer"`
}

func getDwmClusterAppDirs() []string {
	var paths []string
	dwmPath := ula.GetEnvString("DWMPATH", DEF_DWM_DIR)
	files, err := ioutil.ReadDir(dwmPath)
	if err != nil {
		return paths
	}

	for _, file := range files {
		if file.IsDir() == false {
			continue
		}
		paths = append(paths, filepath.Join(dwmPath, file.Name()))
	}

	return paths
}

func readAppInitialLayout(fname string) (*caCfgInitialLayout, error) {

	rdata, err := ioutil.ReadFile(fname)
	if err != nil {
		return nil, err
	}

	ilt := new(caCfgInitialLayout)
	err = json.Unmarshal(rdata, ilt)
	if err != nil {
		return nil, err
	}

	return ilt, nil
}

func newCAVLayoutFromCfg(appDir string) (*core.CALayout, error) {

	fname := filepath.Join(appDir, initialLayout)

	ilt, err := readAppInitialLayout(fname)
	if err != nil {
		return nil, err
	}

	appName := ilt.AppName

	calayout := core.CALayout{
		Vlayers: make([]core.CAVlayer, 0),
	}

	for _, r := range ilt.Vlayer {

		var (
			vdisplayId int
			coord      string
		)

		if r.Zorder == nil {
			WLog.Printf("VID %d will be skipped because tmp_z_order is not specified", r.VID)
			continue
		}

		if r.Coord == nil || *r.Coord == "global" {
			coord = "global"
		} else if *r.Coord == "vdisplay" {
			coord = "vdisplay"
			if r.VdisplayId == nil {
				WLog.Printf("VID %d will be skipped because vdisplay_id is not specified", r.VID)
				continue
			}
			vdisplayId = *r.VdisplayId
		} else {
			WLog.Printf("VID %d will be skipped because coord is not specified", r.VID)
			continue
		}

		vlayer := core.CAVlayer{
			AppName:    appName,
			VID:        r.VID,
			ZOrder:     *r.Zorder,
			Coord:      coord,
			VdisplayId: vdisplayId,
			VirtualW:   r.VirtualW,
			VirtualH:   r.VirtualH,
			VsrcX:      r.VsrcX,
			VsrcY:      r.VsrcY,
			VsrcW:      r.VsrcW,
			VsrcH:      r.VsrcH,
			VdstX:      r.VdstX,
			VdstY:      r.VdstY,
			VdstW:      r.VdstW,
			VdstH:      r.VdstH,
			Visibility: r.Visibility,
			Vsurfaces:  make([]core.CAVsurface, 0),
		}
		for _, s := range r.Vsurface {
			vsurf := core.CAVsurface{
				AppName:    appName,
				VID:        s.VID,
				AppID:      s.AppID,
				PixelW:     s.PixelW,
				PixelH:     s.PixelH,
				PsrcX:      s.PsrcX,
				PsrcY:      s.PsrcY,
				PsrcW:      s.PsrcW,
				PsrcH:      s.PsrcH,
				VdstX:      s.VdstX,
				VdstY:      s.VdstY,
				VdstW:      s.VdstW,
				VdstH:      s.VdstH,
				Visibility: s.Visibility,
			}

			vlayer.Vsurfaces = append(vlayer.Vsurfaces, vsurf)
		}

		calayout.Vlayers = append(calayout.Vlayers, vlayer)
	}

	return &calayout, nil
}

func insertCAVLayoutToCALayoutTree(tree *core.CALayoutTree, nlayout *core.CALayout) {
	for _, nvlayer := range nlayout.Vlayers {

		if len(tree.Vlayers) == 0 {
			tree.Vlayers = append(tree.Vlayers, *nvlayer.Dup())
			continue
		}

		inserted := false

		for idx, r := range tree.Vlayers {
			if r.ZOrder > nvlayer.ZOrder {
				tree.Vlayers = append(tree.Vlayers[:idx+1], tree.Vlayers[idx:]...)
				tree.Vlayers[idx] = *nvlayer.Dup()
				inserted = true
				break
			}
		}

		if inserted {
			continue
		}

		tree.Vlayers = append(tree.Vlayers, *nvlayer.Dup())
	}
}

func ReadCALayoutTreeFromCfg() (*core.CALayoutTree, error) {

	calayoutTree := &core.CALayoutTree{
		Vlayers: make([]core.CAVlayer, 0),
	}

	appDirs := getDwmClusterAppDirs()

	DLog.Println(appDirs)

	for _, appDir := range appDirs {
		calayout, err := newCAVLayoutFromCfg(appDir)
		if err != nil {
			WLog.Printf("newCAVLayoutFromCfg %s error: %s\n", appDir, err)
			continue
		}
		insertCAVLayoutToCALayoutTree(calayoutTree, calayout)
	}

	if len(calayoutTree.Vlayers) == 0 {
		return calayoutTree, errors.New("Cannot read layout config")
	}

	return calayoutTree, nil
}
