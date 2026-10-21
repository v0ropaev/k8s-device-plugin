/**
# Copyright (c) 2026, NVIDIA CORPORATION.  All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
**/

package resource

import (
	"testing"

	"github.com/NVIDIA/go-nvlib/pkg/nvpci"
	"github.com/stretchr/testify/require"
)

func TestVfioDeviceGetTotalMemoryMiB(t *testing.T) {
	const (
		mib = uint64(1024 * 1024)
		gib = 1024 * mib

		// Only the low bits of the resource flags matter here: they mark
		// a BAR as a 32-bit or a 64-bit memory window.
		flags32bit = 0x40200
		flags64bit = 0x14220c
	)

	bar := func(size uint64, flags uint64) *nvpci.MemoryResource {
		return &nvpci.MemoryResource{
			Start: 0,
			End:   uintptr(size - 1),
			Flags: flags,
		}
	}

	testCases := []struct {
		description string
		resources   nvpci.MemoryResources
		expected    uint64
	}{
		{
			description: "single 64-bit BAR",
			resources: nvpci.MemoryResources{
				0: bar(16*mib, flags32bit),
				1: bar(16*gib, flags64bit),
			},
			expected: 16 * 1024,
		},
		{
			description: "64-bit BARs are summed and rounded up to a power of two",
			resources: nvpci.MemoryResources{
				0: bar(16*mib, flags32bit),
				1: bar(64*gib, flags64bit),
				3: bar(32*mib, flags64bit),
			},
			expected: 128 * 1024,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			d := vfioDevice{
				nvidiaPCIDevice: &nvpci.NvidiaPCIDevice{
					Resources: tc.resources,
				},
			}

			memory, err := d.GetTotalMemoryMiB()
			require.NoError(t, err)
			require.Equal(t, tc.expected, memory)
		})
	}
}
