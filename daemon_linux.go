// Copyright 2015 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by
// license that can be found in the LICENSE file.

// Package daemon linux version
package daemon

import (
	// "github.com/oliveagle/goos"
	"os"
)

// Get the daemon properly
func newDaemon(name, path, description string) (Daemon, error) {

	if _, err := os.Stat("/run/systemd/system"); err == nil {
		return &systemDRecord{name, path, description}, nil
	}

	if _, err := os.Stat("/etc/init"); err == nil {
		if _, err := os.Stat("/sbin/initctl"); err == nil {
			return &upstartRecord{name, path, description}, nil
		}
	}

	return &systemVRecord{name, path, description}, nil
}

// Get executable path
func execPath() (string, error) {
	return os.Readlink("/proc/self/exe")
}
