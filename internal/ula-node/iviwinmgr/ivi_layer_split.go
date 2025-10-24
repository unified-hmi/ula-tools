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

package iviwinmgr

import (
	"ula-tools/internal/ula"
)

type PLayerSplitIdTbl struct {
	RDisplayId int

	IdPair map[int]int
}

var pLayerSplitIDs []PLayerSplitIdTbl

func makeDiffPLayerSplitTbl(diffPLayerSplitIDs *[]PLayerSplitIdTbl, rDisplayId int, layerVID int, newID int) {

	for _, splitID := range *diffPLayerSplitIDs {
		if splitID.RDisplayId == rDisplayId {
			if v, ok := splitID.IdPair[layerVID]; ok {
				if v == newID {
					return
				}
			}
		}
	}

	pLayerSplitID := PLayerSplitIdTbl{
		RDisplayId: rDisplayId,
		IdPair:     make(map[int]int),
	}
	pLayerSplitID.IdPair[layerVID] = newID
	*diffPLayerSplitIDs = append(*diffPLayerSplitIDs, pLayerSplitID)
}

func genSplitLayerID(
	spscrns *ula.NodePixelScreens, layerVID int, rDisplayId int, diffPLayerSplitIDs *[]PLayerSplitIdTbl) int {

	for _, splitID := range pLayerSplitIDs {
		if splitID.RDisplayId == rDisplayId {
			if v, ok := splitID.IdPair[layerVID]; ok {
				makeDiffPLayerSplitTbl(diffPLayerSplitIDs, rDisplayId, layerVID, splitID.IdPair[layerVID])
				return v
			}
		}
	}

	/* FIX-ME. should prepare dedicated ID genelator instead. */
	newID := layerVID*100 + rDisplayId
RETRY:
	for _, spscrn := range spscrns.Pscreens {
		for _, splayer := range spscrn.Players {
			if splayer.VID == newID {
				newID++
				goto RETRY
			}
		}
	}

	for _, splitID := range pLayerSplitIDs {
		if _, ok := splitID.IdPair[newID]; ok {
			newID++
			goto RETRY
		}
		for _, id := range splitID.IdPair {
			if id == newID {
				newID++
				goto RETRY
			}
		}
	}

	for _, splitID := range pLayerSplitIDs {
		if splitID.RDisplayId == rDisplayId {
			splitID.IdPair[layerVID] = newID
			makeDiffPLayerSplitTbl(diffPLayerSplitIDs, rDisplayId, layerVID, newID)
			return newID
		}
	}
	pLayerSplitID := PLayerSplitIdTbl{
		RDisplayId: rDisplayId,
		IdPair:     make(map[int]int),
	}
	pLayerSplitID.IdPair[layerVID] = newID
	pLayerSplitIDs = append(pLayerSplitIDs, pLayerSplitID)

	makeDiffPLayerSplitTbl(diffPLayerSplitIDs, rDisplayId, layerVID, newID)

	return newID
}

func splitIviLayer(
	spscrns *ula.NodePixelScreens) (*ula.NodePixelScreens, error) {

	diffPLayerSplitIDs := make([]PLayerSplitIdTbl, 0)

	for sidx, spscrn := range spscrns.Pscreens {

		for pidx, splayer := range spscrn.Players {
			layerVID := splayer.VID

			var layerCnt int
			for _, ckscrn := range spscrns.Pscreens {
				for _, cklayer := range ckscrn.Players {
					if layerVID == cklayer.VID {
						layerCnt++
					}
				}
			}

			if layerCnt < 2 {
				continue
			}

			for i := sidx; i < len(spscrns.Pscreens); i++ {

				pscrn := spscrns.Pscreens[i]

				for i := pidx; i < len(pscrn.Players); i++ {

					player := pscrn.Players[i]

					if layerVID == player.VID {
						player.VID =
							genSplitLayerID(spscrns, layerVID, pscrn.Rdisplay.RDisplayId, &diffPLayerSplitIDs)
						pscrn.Players[i] = player
					}
				}
				pidx = 0
			}
		}
	}

	pLayerSplitIDs = diffPLayerSplitIDs
	return spscrns, nil

}

func splitLayer(srvPixScreens *ula.NodePixelScreens) (*ula.NodePixelScreens, error) {

	dpscrns, err := splitIviLayer(srvPixScreens)
	if err != nil {
		return nil, err
	}
	return dpscrns, nil
}
