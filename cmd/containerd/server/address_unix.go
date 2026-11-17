//go:build !windows

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

import "path/filepath"

// isLocalSocketAddress reports whether address refers to a filesystem-backed
// unix socket rather than a network endpoint. On unix every absolute path is
// a local socket.
func isLocalSocketAddress(address string) bool {
	return filepath.IsAbs(address)
}
