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

func NewRdisplayCommandData(rdisp *ula.RealDisplay, players []ula.PixelLayer) (*RdisplayCommandData, error) {

	dcomm := RdisplayCommandData{
		Rdisplay: *rdisp.Dup(),
		Players:  ula.DupPixelLayerSlice(players),
	}

	return &dcomm, nil
}

func NewRdisplayCommandDataWithSafetyArea(rdisp *ula.RealDisplay, players []ula.PixelLayer, safetyareas []ula.PixelSafetyArea) (*RdisplayCommandData, error) {

	dcomm := RdisplayCommandData{
		Rdisplay:    *rdisp.Dup(),
		Players:     ula.DupPixelLayerSlice(players),
		SafetyAreas: ula.DupPixelSafetyAreaSlice(safetyareas),
	}

	return &dcomm, nil
}

func NewEmptyLocalCommandReq() (*LocalCommandReq, error) {

	ltq := LocalCommandReq{}

	return &ltq, nil
}
