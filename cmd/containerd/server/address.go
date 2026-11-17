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

	srvconfig "github.com/containerd/containerd/v2/cmd/containerd/server/config"
	"github.com/containerd/containerd/v2/defaults"
)

// resolveGRPCAddress determines the address of the main gRPC endpoint.
//
// When the address was not explicitly configured (it is empty or still holds
// the built-in default) it is derived from the state directory so that every
// instance keeps its sockets under its own runtime directory. Explicitly
// configured addresses, including non-unix addresses such as TCP endpoints,
// are always honored verbatim. With the default state directory the derived
// value is identical to defaults.DefaultAddress, preserving the historical
// default behavior.
func resolveGRPCAddress(config *srvconfig.Config, address string) string {
	if address != "" && address != defaults.DefaultAddress {
		return address
	}
	if !isLocalSocketAddress(defaults.DefaultAddress) {
		// Non-filesystem defaults (e.g. Windows named pipes) cannot be
		// relocated under the state directory.
		return defaults.DefaultAddress
	}
	return filepath.Join(config.State, filepath.Base(defaults.DefaultAddress))
}

// resolveTTRPCAddress determines the address of the TTRPC endpoint. It mirrors
// resolveGRPCAddress and, for an unset or default value, derives the address
// from the gRPC address to preserve the historical "<grpc>.ttrpc"
// relationship.
func resolveTTRPCAddress(config *srvconfig.Config, grpcAddress, ttrpcAddress string) string {
	defaultTTRPCAddress := defaults.DefaultAddress + ".ttrpc"
	if ttrpcAddress != "" && ttrpcAddress != defaultTTRPCAddress {
		return ttrpcAddress
	}
	if !isLocalSocketAddress(defaultTTRPCAddress) {
		return defaultTTRPCAddress
	}
	return grpcAddress + ".ttrpc"
}
