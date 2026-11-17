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

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"testing"

	srvconfig "github.com/containerd/containerd/v2/cmd/containerd/server/config"
	"github.com/containerd/containerd/v2/defaults"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCleanupStaleSocketsRemovesOrphans(t *testing.T) {
	defaultDir := t.TempDir()
	stateDir := t.TempDir()
	sockPath := filepath.Join(defaultDir, "containerd.sock")
	ttrpcPath := sockPath + ".ttrpc"

	listener, err := net.Listen("unix", sockPath)
	require.NoError(t, err)
	listener.Close()

	other, err := net.Listen("unix", ttrpcPath)
	require.NoError(t, err)
	defer other.Close()

	config := &srvconfig.Config{
		State:   stateDir,
		Plugins: map[string]any{},
	}
	cleanupStaleSockets(context.Background(), config, defaultDir, sockPath, filepath.Join(defaultDir, "debug.sock"))

	_, err = os.Lstat(sockPath)
	assert.ErrorIs(t, err, os.ErrNotExist, "stale socket without a listener must be removed")
	_, err = os.Lstat(ttrpcPath)
	assert.NoError(t, err, "live socket served by another listener must be kept")
}

func TestCleanupStaleSocketsKeepsNonSocketFiles(t *testing.T) {
	defaultDir := t.TempDir()
	stateDir := t.TempDir()
	sockPath := filepath.Join(defaultDir, "containerd.sock")

	require.NoError(t, os.WriteFile(sockPath, []byte("not a socket"), 0o600))

	config := &srvconfig.Config{
		State:   stateDir,
		Plugins: map[string]any{},
	}
	cleanupStaleSockets(context.Background(), config, defaultDir, sockPath, filepath.Join(defaultDir, "debug.sock"))

	_, err := os.Lstat(sockPath)
	assert.NoError(t, err, "files that are not sockets must never be removed")
}

func TestCleanupStaleDefaultSocketsSkipsDefaultState(t *testing.T) {
	// With the built-in default state directory the entry guard returns before
	// touching the filesystem, so no socket files may be touched even if the
	// default directory contains stale entries.
	config := &srvconfig.Config{
		State:   defaults.DefaultStateDir,
		Plugins: map[string]any{},
	}
	cleanupStaleDefaultSockets(context.Background(), config)
}

func TestCleanupStaleSocketsSkipsDebugWhenInUse(t *testing.T) {
	defaultDir := t.TempDir()
	stateDir := t.TempDir()
	sockPath := filepath.Join(defaultDir, "containerd.sock")
	debugPath := filepath.Join(defaultDir, "debug.sock")

	listener, err := net.Listen("unix", debugPath)
	require.NoError(t, err)
	defer listener.Close()

	config := &srvconfig.Config{
		State: stateDir,
		Plugins: map[string]any{
			"io.containerd.server.v1.debug": map[string]any{
				"address": debugPath,
			},
		},
	}
	cleanupStaleSockets(context.Background(), config, defaultDir, sockPath, debugPath)

	_, err = os.Lstat(debugPath)
	assert.NoError(t, err, "a debug socket this instance is configured to use must be kept")
}
