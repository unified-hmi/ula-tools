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

import (
	"encoding/json"
	"ula-tools/internal/ula"
	"ula-tools/internal/ula-client/core"
	. "ula-tools/internal/ulog"
)

func convCAVSurface2UPIVsurface(csurf *core.CAVsurface, appName string) *UPIVsurface {

	usurf := new(UPIVsurface)

	usurf.AppName = appName

	usurf.VID = csurf.VID

	usurf.PixelW = csurf.PixelW
	usurf.PixelH = csurf.PixelH

	usurf.PsrcX = csurf.PsrcX
	usurf.PsrcY = csurf.PsrcY
	usurf.PsrcW = csurf.PsrcW
	usurf.PsrcH = csurf.PsrcH

	usurf.VdstX = csurf.VdstX
	usurf.VdstY = csurf.VdstY
	usurf.VdstW = csurf.VdstW
	usurf.VdstH = csurf.VdstH

	usurf.Visibility = csurf.Visibility
	return usurf
}

func convCAVlayer2UPIVlayer(clayer *core.CAVlayer) *UPIVlayer {

	ulayer := new(UPIVlayer)

	ulayer.AppName = clayer.AppName

	ulayer.VID = clayer.VID
	ulayer.Coord = clayer.Coord
	ulayer.VdisplayId = clayer.VdisplayId

	ulayer.VirtualW = clayer.VirtualW
	ulayer.VirtualH = clayer.VirtualH

	ulayer.VsrcX = clayer.VsrcX
	ulayer.VsrcY = clayer.VsrcY
	ulayer.VsrcW = clayer.VsrcW
	ulayer.VsrcH = clayer.VsrcH

	ulayer.VdstX = clayer.VdstX
	ulayer.VdstY = clayer.VdstY
	ulayer.VdstW = clayer.VdstW
	ulayer.VdstH = clayer.VdstH

	ulayer.Visibility = clayer.Visibility

	ulayer.Surface = make([]UPIVsurface, 0)

	for _, cVsurf := range clayer.Vsurfaces {
		usurf := convCAVSurface2UPIVsurface(&cVsurf, ulayer.AppName)
		ulayer.Surface = append(ulayer.Surface, *usurf)
	}

	return ulayer
}

func convCAVLayoutTree2UPIVscreen(ctree *core.CALayoutTree) *UPIVscreen {

	uscreen := new(UPIVscreen)

	uscreen.Command = "initial_vscreen"

	for _, cLayer := range ctree.Vlayers {
		ulayer := convCAVlayer2UPIVlayer(&cLayer)
		uscreen.Layer = append(uscreen.Layer, *ulayer)
	}

	return uscreen
}

func convVirtualSurface2UPIVsurface(vsurf *ula.VirtualSurface, appName string) *UPIVsurface {
	usurf := new(UPIVsurface)

	usurf.AppName = appName

	usurf.VID = vsurf.VID

	usurf.PixelW = vsurf.PixelW
	usurf.PixelH = vsurf.PixelH

	usurf.PsrcX = vsurf.PsrcX
	usurf.PsrcY = vsurf.PsrcY
	usurf.PsrcW = vsurf.PsrcW
	usurf.PsrcH = vsurf.PsrcH

	usurf.VdstX = vsurf.VdstX
	usurf.VdstY = vsurf.VdstY
	usurf.VdstW = vsurf.VdstW
	usurf.VdstH = vsurf.VdstH

	usurf.Visibility = &vsurf.Visibility
	return usurf
}

func GenerateUlaCommInitialVscreen(ctree *core.CALayoutTree) (string, error) {

	var msg string
	upiscreen := convCAVLayoutTree2UPIVscreen(ctree)

	jsonBytes, err := json.Marshal(upiscreen)
	if err != nil {
		ELog.Println("JSON Marshal error: ", err)
		return msg, err
	}

	msg = string(jsonBytes)

	return msg, nil
}
