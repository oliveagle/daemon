// Copyright 2015 Igor Dolzhikov. All rights reserved.
// Use of this source code is governed by
// license that can be found in the LICENSE file.

package daemon

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"text/template"
    "strings"
    "github.com/oliveagle/hickwall/utils"
)

// upstartRecord - standard record (struct) for linux upstart version of daemon package
type upstartRecord struct {
	name        string
	description string
}

// Standard service path for upstart daemons
func (linux *upstartRecord) servicePath() string {
	return "/etc/init/" + linux.name + ".conf"
}

// Check service is installed
func (linux *upstartRecord) checkInstalled() bool {

	if _, err := os.Stat(linux.servicePath()); err == nil {
		return true
	}

	return false
}

// Check service is running
func (linux *upstartRecord) checkRunning() (string, bool) {
	output, err := exec.Command("initctl", "status", linux.name).Output()
	if err == nil {
		if matched, err := regexp.MatchString("start/running", string(output)); err == nil && matched {
			reg := regexp.MustCompile("process [0-9]+")
			data := reg.FindStringSubmatch(string(output))
			if len(data) > 1 {
				return "Service (pid  " + data[1] + ") is running...", true
			}
			return "Service is running...", true
		}
	}

	return "Service is stopped", false
}

func (linux *upstartRecord) InstallFromPath(thePath string) (string, error) {
	installAction := "Install " + linux.description + ":"

	if checkPrivileges() == false {
		return installAction + failed, errors.New(rootPrivileges)
	}

	srvPath := linux.servicePath()

	if linux.checkInstalled() == true {
		return installAction + failed, errors.New(linux.description + " already installed")
	}

	file, err := os.Create(srvPath)
	if err != nil {
		return installAction + failed, err
	}
	defer file.Close()

	isexec, err := IsExecutable(thePath)
	if err != nil {
		return installAction + failed, err
	}
	if !isexec {
		return installAction + failed, fmt.Errorf("target is not executable: %s", thePath)
	}

	templ, err := template.New("upstartConf").Parse(upstartConfig)
	if err != nil {
		return installAction + failed, err
	}

	if err := templ.Execute(
		file,
		&struct {
			Name, Description, Path string
		}{linux.name, linux.description, thePath},
	); err != nil {
		return installAction + failed, err
	}

	if err := exec.Command("initctl", "reload-configuration").Run(); err != nil {
		return installAction + failed, err
	}

	// if err := exec.Command("initctl", "enable", linux.name+".service").Run(); err != nil {
	// 	return installAction + failed, err
	// }

	return installAction + success, nil
}

// Install the service
func (linux *upstartRecord) Install() (string, error) {
	installAction := "Install " + linux.description + ":"

	execPatch, err := executablePath(linux.name)
	if err != nil {
		return installAction + failed, err
	}
	return linux.InstallFromPath(execPatch)
}

func (linux *upstartRecord) stop_only() (bool, error) {
    cmd := exec.Command("initctl", "stop", linux.name)
    output, err := cmd.CombinedOutput()
    line := fmt.Sprintf("%v: %s", err, string(output))
    if err != nil {
        if strings.Contains(line, "Unknown instance") {
            return true, nil
        } else {
            return false, fmt.Errorf("unknown error: %s", line)
        }
    } else {
        if strings.Contains(line, "stop/waiting") {
            return true, nil
        } else {
            return false, fmt.Errorf("unknown output: %v", line)
        }
    }
}

// Remove the service
func (linux *upstartRecord) Remove() (string, error) {
    defer utils.Recover_and_log()
	removeAction := "Removing " + linux.description + ":"

	if checkPrivileges() == false {
		return removeAction + failed, errors.New(rootPrivileges)
	}

	if linux.checkInstalled() == false {
		return removeAction + failed, errors.New(linux.description + " is not installed")
	}

    if _,err := linux.stop_only(); err != nil {
    return removeAction + failed, err
    }
//	if err := exec.Command("initctl", "stop", linux.name).Run(); err != nil {
//		return removeAction + failed, err
//	}

	if err := os.Remove(linux.servicePath()); err != nil {
		return removeAction + failed, err
	}

    if err := exec.Command("initctl", "reload-configuration").Run(); err != nil {
		return removeAction + failed, err
	}

	return removeAction + success, nil
}

// Start the service
func (linux *upstartRecord) Start() (string, error) {
	startAction := "Starting " + linux.description + ":"

	if checkPrivileges() == false {
		return startAction + failed, errors.New(rootPrivileges)
	}

	if linux.checkInstalled() == false {
		return startAction + failed, errors.New(linux.description + " is not installed")
	}

	if _, status := linux.checkRunning(); status == true {
		return startAction + failed, errors.New("service already running")
	}

	if err := exec.Command("initctl", "start", linux.name).Run(); err != nil {
		return startAction + failed, err
	}

	return startAction + success, nil
}

// Stop the service
func (linux *upstartRecord) Stop() (string, error) {
	stopAction := "Stopping " + linux.description + ":"

	if checkPrivileges() == false {
		return stopAction + failed, errors.New(rootPrivileges)
	}

	if linux.checkInstalled() == false {
		return stopAction + failed, errors.New(linux.description + " is not installed")
	}

	if _, status := linux.checkRunning(); status == false {
		return stopAction + failed, errors.New("service already stopped")
	}

	if err := exec.Command("initctl", "stop", linux.name).Run(); err != nil {
		return stopAction + failed, err
	}

	return stopAction + success, nil
}

// Status - Get service status
func (linux *upstartRecord) Status() (string, error) {

	if checkPrivileges() == false {
		return "", errors.New(rootPrivileges)
	}

	if linux.checkInstalled() == false {
		return "Status could not defined", errors.New(linux.description + " is not installed")
	}

	statusAction, _ := linux.checkRunning()

	return statusAction, nil
}

var upstartConfig = `
description "{{.Name}}, {{.Description}}"

start on runlevel [2345]
stop on runlevel [!2345]

respawn
pre-start exec sleep 1
exec {{.Path}}
`
