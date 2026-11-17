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
	"errors"
	"net"
	"os"
	"path/filepath"
	"time"

	srvconfig "github.com/containerd/containerd/v2/cmd/containerd/server/config"
	"github.com/containerd/containerd/v2/defaults"
	"github.com/containerd/log"
)

// cleanupStaleDefaultSockets removes socket files left behind in the default
// state directory by an installation that used the default runtime directory.
//
// It runs only when this instance uses a non-default state directory, so the
// default behavior (and the default state directory layout) is untouched.
// A file is removed only when it is not backed by a live, accepting listener:
// a socket still served by another instance, or a socket exposed through a
// bind mount, is left in place. This keeps running services, including
// co-located instances, unaffected and makes the operation safe to run on
// every start (upgrades and reinstalls included).
func cleanupStaleDefaultSockets(ctx context.Context, config *srvconfig.Config) {
	if config.State == "" || config.State == defaults.DefaultStateDir {
		return
	}

	cleanupStaleSockets(ctx, config, defaults.DefaultStateDir, defaults.DefaultAddress, defaults.DefaultDebugAddress)
}

// cleanupStaleSockets removes leftover containerd socket files from
// defaultStateDir. It is split out with explicit paths so the behavior can be
// exercised against temporary directories in tests.
func cleanupStaleSockets(ctx context.Context, config *srvconfig.Config, defaultStateDir, defaultAddress, defaultDebugAddress string) {
	candidates := []string{
		defaultAddress,
		defaultAddress + ".ttrpc",
	}
	// Only clean up the debug endpoint when this instance is not going to use
	// it (a debug socket at the default path would otherwise be created by the
	// debug plugin below).
	if debugAddress := readString(config.Plugins, "io.containerd.server.v1.debug", "address"); debugAddress != "" && debugAddress != defaultDebugAddress {
		candidates = append(candidates, defaultDebugAddress)
	}

	for _, path := range candidates {
		if filepath.Dir(path) != defaultStateDir {
			continue
		}
		removeStaleSocket(ctx, path)
	}
}

// removeStaleSocket deletes a single stale socket file. Failures are logged,
// never fatal: a bind-mounted or otherwise protected path must not block
// daemon startup.
func removeStaleSocket(ctx context.Context, path string) {
	info, err := os.Lstat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return
		}
		log.G(ctx).WithError(err).WithField("path", path).Warn("failed to stat stale socket in default state directory")
		return
	}

	if info.Mode()&os.ModeSocket == 0 {
		// Never touch anything that is not a socket (symlink, regular file,
		// or a bind-mount entry represented with a different mode).
		return
	}

	if socketIsLive(path) {
		log.G(ctx).WithField("path", path).Info("keeping socket in default state directory, another instance is serving it")
		return
	}

	if err := os.Remove(path); err != nil {
		log.G(ctx).WithError(err).WithField("path", path).Warn("failed to remove stale socket from default state directory")
		return
	}
	log.G(ctx).WithField("path", path).Info("removed stale socket left in default state directory")
}

// socketIsLive reports whether a unix socket path is served by an accepting
// listener. A successful dial means another process owns the endpoint and it
// must not be removed. A dial failure, including a timeout, means the file is
// a leftover with no active listener.
func socketIsLive(path string) bool {
	conn, err := net.DialTimeout("unix", path, 200*time.Millisecond)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
