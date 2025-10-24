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
	"errors"
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-node"
)

type workRvgpu struct {
	rdisplay ula.RealDisplay

	/* Geometry is Coordinated by RDisplay  */
	players      []ula.PixelLayer
	psafetyareas []ula.PixelSafetyArea
}

type RvgpuPlugin struct{}

func (plugin RvgpuPlugin) GenerateLocalCommandReq(acdata *ula.ApplyCommandData, sps *ula.NodePixelScreens) ([]*ulanode.LocalCommandReq, error) {
	ltqs := []*ulanode.LocalCommandReq{}

	wRvgpuMap, err := generateWorkRvgpu(acdata.NPScreens.Dup())
	if err != nil {
		return ltqs, errors.New("generateWorkRvgpu error")
	}

	oldwRvgpuMap := make(map[int]workRvgpu)
	if acdata.Command != "initial_vscreen" {
		oldwRvgpuMap, err = generateWorkRvgpu(sps.Dup())
		if err != nil {
			return ltqs, errors.New("generateWorkRvgpu error")
		}
	}

	ret, err := pickupInitialVScreen(wRvgpuMap, oldwRvgpuMap)
	if err == nil && ret != nil {
		ltqs = append(ltqs, ret)
	}

	return ltqs, nil
}

func generateWorkRvgpu(
	spscrns *ula.NodePixelScreens) (map[int]workRvgpu, error) {

	workRvgpuMap := make(map[int]workRvgpu)

	for _, pscrn := range spscrns.Pscreens {
		wvdisp := workRvgpu{
			rdisplay:     *pscrn.Rdisplay.Dup(),
			players:      ula.DupPixelLayerSlice(pscrn.Players),
			psafetyareas: ula.DupPixelSafetyAreaSlice(pscrn.PsafetyAreas),
		}
		workRvgpuMap[wvdisp.rdisplay.RDisplayId] = wvdisp
	}

	err := isValidWorkRvgpuMap(workRvgpuMap)

	return workRvgpuMap, err
}

func isValidWorkRvgpuMap(wRvgpuMap map[int]workRvgpu) error {

	rdisplayMap := make(map[int]bool)
	for _, workRvgpu := range wRvgpuMap {
		if rdisplayMap[workRvgpu.rdisplay.RDisplayId] {
			return errors.New("PixelScreens has duplicate RDisplayId")
		}
		rdisplayMap[workRvgpu.rdisplay.RDisplayId] = true

		playerMap := make(map[int]bool)
		for _, player := range workRvgpu.players {
			if playerMap[player.VID] {
				return errors.New("RealDisplay has duplicate PixelLayer VID")
			}
			playerMap[player.VID] = true

			psurfaceMap := make(map[int]bool)
			for _, psurface := range player.Psurfaces {
				if psurfaceMap[psurface.VID] {
					return errors.New("PixelLayer has duplicate PixelSurface VID")
				}
				psurfaceMap[psurface.VID] = true
			}
		}
	}

	return nil
}

func generateLocalCommandWithSafetyAreas(lcomm string,
	workRvgpuMap map[int]workRvgpu,
	pickupPlayersMap map[int][]ula.PixelLayer,
	pickupPSafetyAreasMap map[int][]ula.PixelSafetyArea) (*ulanode.LocalCommandReq, error) {

	dcomms := make([]ulanode.RdisplayCommandData, 0)

	for key, workRvgpu := range workRvgpuMap {
		pickupPlayers := pickupPlayersMap[key]
		pickupSafetyAreas := pickupPSafetyAreasMap[key]
		if len(pickupPlayers) == 0 {
			continue
		}
		dcomm, err := ulanode.NewRdisplayCommandDataWithSafetyArea(&workRvgpu.rdisplay, pickupPlayers, pickupSafetyAreas)
		if err != nil {
			return nil, err
		}

		dcomms = append(dcomms, *dcomm)
	}

	ltq, err := ulanode.NewEmptyLocalCommandReq()
	if err != nil {
		return nil, err
	}
	ltq.Command = lcomm
	ltq.RDComms = dcomms

	return ltq, nil
}

func pickupInitialVScreen(
	wRvgpuMap map[int]workRvgpu,
	oldwRvgpuMap map[int]workRvgpu) (*ulanode.LocalCommandReq, error) {

	lcomm := ""
	pickupPlayersMap := make(map[int][]ula.PixelLayer, len(wRvgpuMap))
	pickupPsafetyAreasMap := make(map[int][]ula.PixelSafetyArea, len(wRvgpuMap))

	for key, wRvgpu := range wRvgpuMap {
		oldwRvgpu := oldwRvgpuMap[key]
		pickupPlayers := pickupPlayersMap[key]
		pickupPsafetyareas := pickupPsafetyAreasMap[key]

		if len(oldwRvgpu.players) == 0 && len(wRvgpu.players) != 0 {
			lcomm = "initial_vscreen"
			pickupPlayers = ula.DupPixelLayerSlice(wRvgpu.players)
			pickupPsafetyareas = ula.DupPixelSafetyAreaSlice(wRvgpu.psafetyareas)

			oldwRvgpuMap[key] = wRvgpu
			pickupPlayersMap[key] = pickupPlayers
			pickupPsafetyAreasMap[key] = pickupPsafetyareas
		}
	}

	if lcomm == "" {
		return nil, nil
	}
	ltq, err := generateLocalCommandWithSafetyAreas(lcomm, wRvgpuMap, pickupPlayersMap, pickupPsafetyAreasMap)
	if err != nil {
		return nil, err
	}

	return ltq, nil
}
