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
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	srvconfig "github.com/containerd/containerd/v2/cmd/containerd/server/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCleanupStaleSocketsHandlesBindMount ensures a socket exposed through a
// bind mount (a common container-orchestrator setup) is never unlinked.
// Unlinking the mounted entry would only detach the mount point while the
// serving socket stays live, and mount removal can also fail with EBUSY.
func TestCleanupStaleSocketsHandlesBindMount(t *testing.T) {
	if os.Geteuid() != 0 {
		t.Skip("bind-mount test requires a user/mount namespace; skipping unprivileged run")
	}

	defaultDir := t.TempDir()
	stateDir := t.TempDir()

	sourceDir := t.TempDir()
	sourceSock := filepath.Join(sourceDir, "containerd.sock")
	ln, err := net.Listen("unix", sourceSock)
	require.NoError(t, err)
	defer ln.Close()

	mountTarget := filepath.Join(defaultDir, "containerd.sock")
	placeholder, err := net.Listen("unix", mountTarget)
	require.NoError(t, err)
	placeholder.Close()

	cmd := exec.Command("mount", "--bind", sourceSock, mountTarget)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Skipf("bind mount unavailable: %v: %s", err, string(out))
	}
	t.Cleanup(func() {
		_ = exec.Command("umount", mountTarget).Run()
	})

	config := &srvconfig.Config{
		State:   stateDir,
		Plugins: map[string]any{},
	}
	cleanupStaleSockets(context.Background(), config, defaultDir, mountTarget, filepath.Join(defaultDir, "debug.sock"))

	_, err = os.Lstat(mountTarget)
	assert.NoError(t, err, "bind-mounted socket must remain in place")
}
