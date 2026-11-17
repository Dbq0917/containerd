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

import "github.com/containerd/containerd/v2/internal/localaddr"

// isLocalSocketAddress reports whether address refers to a filesystem-backed
// AF_UNIX socket. Named pipes are not filesystem backed and therefore cannot
// be relocated under the state directory.
func isLocalSocketAddress(address string) bool {
	network, _, _ := localaddr.Parse(address)
	return network != localaddr.NamedPipe
}
