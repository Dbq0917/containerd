/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package server

import (
	"path/filepath"
	"testing"

	srvconfig "github.com/containerd/containerd/v2/cmd/containerd/server/config"
	"github.com/containerd/containerd/v2/defaults"
	"github.com/stretchr/testify/assert"
)

func TestResolveGRPCAddress(t *testing.T) {
	customState := "/data/containerd/run"

	for _, tc := range []struct {
		name    string
		state   string
		address string
		want    string
	}{
		{
			name:    "default state and empty address keeps historical default",
			state:   defaults.DefaultStateDir,
			address: "",
			want:    defaults.DefaultAddress,
		},
		{
			name:    "default state and default address keeps historical default",
			state:   defaults.DefaultStateDir,
			address: defaults.DefaultAddress,
			want:    defaults.DefaultAddress,
		},
		{
			name:    "custom state derives socket under state directory",
			state:   customState,
			address: "",
			want:    filepath.Join(customState, filepath.Base(defaults.DefaultAddress)),
		},
		{
			name:    "custom state with default address still follows state",
			state:   customState,
			address: defaults.DefaultAddress,
			want:    filepath.Join(customState, filepath.Base(defaults.DefaultAddress)),
		},
		{
			name:    "explicit address is always honored",
			state:   customState,
			address: "/srv/run/containerd.sock",
			want:    "/srv/run/containerd.sock",
		},
		{
			name:    "explicit address under default state is honored",
			state:   customState,
			address: defaults.DefaultAddress,
			want:    filepath.Join(customState, filepath.Base(defaults.DefaultAddress)),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !isLocalSocketAddress(defaults.DefaultAddress) {
				t.Skip("address derivation only applies to filesystem-backed sockets")
			}
			config := &srvconfig.Config{State: tc.state}
			assert.Equal(t, tc.want, resolveGRPCAddress(config, tc.address))
		})
	}
}

func TestResolveTTRPCAddress(t *testing.T) {
	customState := "/data/containerd/run"
	defaultTTRPC := defaults.DefaultAddress + ".ttrpc"

	for _, tc := range []struct {
		name      string
		state     string
		ttrpc     string
		wantGRPC  string
		wantTTRPC string
	}{
		{
			name:      "default state keeps historical default",
			state:     defaults.DefaultStateDir,
			ttrpc:     "",
			wantGRPC:  defaults.DefaultAddress,
			wantTTRPC: defaultTTRPC,
		},
		{
			name:      "custom state derives ttrpc next to grpc socket",
			state:     customState,
			ttrpc:     "",
			wantGRPC:  filepath.Join(customState, filepath.Base(defaults.DefaultAddress)),
			wantTTRPC: filepath.Join(customState, filepath.Base(defaults.DefaultAddress)) + ".ttrpc",
		},
		{
			name:      "custom state with default ttrpc still follows state",
			state:     customState,
			ttrpc:     defaultTTRPC,
			wantGRPC:  filepath.Join(customState, filepath.Base(defaults.DefaultAddress)),
			wantTTRPC: filepath.Join(customState, filepath.Base(defaults.DefaultAddress)) + ".ttrpc",
		},
		{
			name:      "explicit ttrpc address is honored",
			state:     customState,
			ttrpc:     "/srv/run/containerd.sock.ttrpc",
			wantGRPC:  filepath.Join(customState, filepath.Base(defaults.DefaultAddress)),
			wantTTRPC: "/srv/run/containerd.sock.ttrpc",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if !isLocalSocketAddress(defaultTTRPC) {
				t.Skip("address derivation only applies to filesystem-backed sockets")
			}
			config := &srvconfig.Config{State: tc.state}
			grpc := resolveGRPCAddress(config, "")
			assert.Equal(t, tc.wantGRPC, grpc)
			assert.Equal(t, tc.wantTTRPC, resolveTTRPCAddress(config, grpc, tc.ttrpc))
		})
	}
}
